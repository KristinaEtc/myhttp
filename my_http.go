package myhttp

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/url"
)

// URL should never be longer than 2,048 characters;
// any long than this some browsers won’t be able to load your page.
const maxURLLen = 2048

// Response struct represents the result of a URL encoding operation
type Response struct {
	OriginWithScheme string
	Encoded          string
	Err              error
}

// MyHTTP struct represents myhttp tool. It contains a pool of worker goroutines
// to generate MD5 hash for multiple URLs concurrently.
type MyHTTP struct {
	maxWorkers int
	inputCh    chan string
	outputCh   chan *Response
}

// Send sends an input URL to the input channel of MyHTTP struct.
func (m *MyHTTP) Send(input string) {
	m.inputCh <- input
}

// Recv returns a receive-only channel of pointers to Response structs.
// It allows clients of `MyHTTP` to receive responses from the worker goroutines.
func (m *MyHTTP) Recv() <-chan *Response {
	return m.outputCh
}

// New is a constructor for `MyHTTP`
func New(parallel int) *MyHTTP {
	return &MyHTTP{
		maxWorkers: parallel,
		inputCh:    make(chan string, parallel),
		outputCh:   make(chan *Response, parallel),
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
				resp := m.process(originURL)
				m.outputCh <- resp
			}(semaphoreChan)
		}
	}
}

func (m *MyHTTP) process(originURL string) *Response {
	// make sure url is valid before getting md5 hash for it
	updatedURL, err := m.validateAndUpdate(originURL)
	if err != nil {
		return &Response{
			OriginWithScheme: originURL,
			Err:              err,
		}
	}

	md5Hash, err := m.encodeMD5(updatedURL)
	if err != nil {
		return &Response{
			OriginWithScheme: updatedURL,
			Err:              err,
		}
	}

	return &Response{
		OriginWithScheme: updatedURL,
		Encoded:          md5Hash,
	}
}

// ValidateURL contains basic rules to detect invalid url
// google says it's expected that url len is smaller than 2048 -
// I also added this rule here.
// It also adds scheme if there isn't such
func (m *MyHTTP) validateAndUpdate(u string) (string, error) {
	parced, err := url.Parse(u)
	switch {
	case err != nil:
		return "", err
	case len(u) >= maxURLLen:
		return "", errors.New("url len is too long")
	case parced.Scheme == "":
		return "http://" + u, nil
	default:
		return u, nil
	}
}

// A simple function to encode a string. It makes sense to operate bytes buffer,
// if we want further optimisation
func (m *MyHTTP) encodeMD5(url string) (string, error) {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:]), nil
}

func (m *MyHTTP) Close() {
	close(m.inputCh)
	close(m.outputCh)
}
