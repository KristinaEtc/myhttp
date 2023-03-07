package myhttp

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// URL should never be longer than 2,048 characters;
// any long than this some browsers won’t be able to load your page.
const maxURLLen = 2048

const defaultTimeout = time.Second * 5

// Result struct represents the result of a URL encoding operation
type Result struct {
	URL     string
	Encoded [md5.Size]byte
	Err     error
}

type Client interface {
	Get(string) (*http.Response, error)
}

func (r *Result) GetHexString() string {
	return fmt.Sprintf("%x", r.Encoded)
}

// MyHTTP struct represents myhttp tool. It contains a pool of worker goroutines
// to generate MD5 hash for multiple URLs concurrently.
type MyHTTP struct {
	maxWorkers int
	httpClient Client
	inputCh    chan string
	outputCh   chan *Result
}

// Send sends an input URL to the input channel of MyHTTP struct.
func (m *MyHTTP) Send(input string) {
	m.inputCh <- input
}

// Recv returns a receive-only channel of pointers to Result structs.
// It allows clients of `MyHTTP` to receive responses from the worker goroutines.
func (m *MyHTTP) Recv() <-chan *Result {
	return m.outputCh
}

// New is a constructor for `MyHTTP`
func New(parallel int) *MyHTTP {
	return &MyHTTP{
		httpClient: &http.Client{Timeout: defaultTimeout},
		maxWorkers: parallel,
		inputCh:    make(chan string, parallel),
		outputCh:   make(chan *Result, parallel),
	}
}

// Run URL processing: read from the input channel, process
func (m *MyHTTP) Run(ctx context.Context) {
	// this buffered channel will block at the concurrency limit
	semaphoreChan := make(chan struct{}, m.maxWorkers)
	defer close(semaphoreChan)

	for {
		select {
		case <-ctx.Done():
			return

		// write commands one by one
		case originURL := <-m.inputCh:
			go func(semCh chan struct{}) {
				// use semaphoreChan to control number of concurrent urls processed
				semCh <- struct{}{}
				defer func() { <-semCh }()
				md5Hash, err := m.process(originURL)
				m.outputCh <- &Result{
					URL:     originURL,
					Encoded: md5Hash,
					Err:     err,
				}
			}(semaphoreChan)
		}
	}
}

// validate url, make an http request and encode to md5 response body
func (m *MyHTTP) process(u string) ([md5.Size]byte, error) {
	emptyResp := [md5.Size]byte{}
	// make sure url is valid before making a http request
	parsedURL, err := m.validateAndUpdate(u)
	if err != nil {
		return emptyResp, err
	}

	resp, err := m.httpClient.Get(parsedURL)
	if err != nil {
		return emptyResp, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return emptyResp, err
	}
	return m.encodeMD5(body), nil
}

// ValidateURL contains basic rules to detect invalid url
// google says it's expected that url len is smaller than 2048 -
// I also added this rule here.
// It also adds scheme if there isn't such to request urls like "example.com"
func (m *MyHTTP) validateAndUpdate(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	switch {
	case err != nil:
		return "", err
	case parsedURL.Scheme == "":
		parsedURL.Scheme = "http"
	case len(u) >= maxURLLen:
		return "", errors.New("url len is too long")
	}
	return parsedURL.String(), nil
}

// A simple function to encode byte array. It makes sense to operate bytes buffer,
// if we want further optimisation
func (m *MyHTTP) encodeMD5(data []byte) [md5.Size]byte {
	return md5.Sum(data)
}

func (m *MyHTTP) Close() {
	close(m.inputCh)
	close(m.outputCh)
}
