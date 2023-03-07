package myhttp

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockClient struct{}

func (m *mockClient) Get(url string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader(url)),
	}, nil
}

func BenchmarkMD5(b *testing.B) {
	s := New(100)
	s.httpClient = &mockClient{}
	defer s.Close()

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		_, err := s.process("example.com/" + time.Now().String())
		assert.NoError(b, err)
		wg.Done()
	}

	wg.Wait()
}
