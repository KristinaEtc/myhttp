package myhttp

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	input    string
	expected map[string]string
}

var testCases = []testCase{
	{"http://www.adjust.com http://google.com", map[string]string{
		"http://www.adjust.com": "73efca339a408a919615181dd7d83b0a",
		"http://google.com":     "c7b920f57e553df2bb68272f61570210",
	}},
	{"adjust.com", map[string]string{
		"http://adjust.com": "b53f3f2ec2e7e01d9e1130baac274a90",
	}},
	{
		"-parallel 3 adjust.com google.com facebook.com yahoo.com\nyandex.com twitter.com\nreddit.com/r/funny reddit.com/r/notfunny baroquemusiclibrary.com", map[string]string{
			"http://adjust.com":              "b53f3f2ec2e7e01d9e1130baac274a90",
			"http://baroquemusiclibrary.com": "369dcc7299cc08bc6bed46ae321da030",
			"http://facebook.com":            "eaa025b9fb1ddeb0042fcfc69205244b",
			"http://google.com":              "c7b920f57e553df2bb68272f61570210",
			"http://reddit.com/r/funny":      "a700cd30f93446b941a501527064f998",
			"http://reddit.com/r/notfunny":   "8185464370a476039cba56a57fc2b12d",
			"http://twitter.com":             "77e8fd27d49a8be49e1acfc9743bf805",
			"http://yahoo.com":               "873c87c71f8bf1d15a53ce0c0676971f",
			"http://yandex.com":              "4cdcf41befe197c6958a8679111dfc4c",
		}},
}

func TestMyHTTP(t *testing.T) {

	for _, ts := range testCases {
		args := append([]string{"./myhttp"}, strings.Fields(ts.input)...)
		cmd := exec.Command(args[0], args[1:]...)

		output, err := cmd.CombinedOutput()
		assert.NoError(t, err)

		res, err := convertToMap(output)
		assert.NoError(t, err)
		assert.Equal(t, ts.expected, res)
	}

}

func convertToMap(inputArray []byte) (map[string]string, error) {
	m := make(map[string]string)
	for _, line := range bytes.Split(inputArray, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		fields := bytes.Split(line, []byte(" "))
		if len(fields) != 2 {
			return nil, fmt.Errorf("expected 2 elements, got %s", string(line))
		}

		key := string(fields[0])
		value := string(fields[1])
		m[key] = value
	}
	return m, nil
}

func BenchmarkMD5(b *testing.B) {
	s := New(10)
	defer s.Close()
	for i := 0; i < b.N; i++ {
		_, err := s.encodeMD5(time.Now().String())
		require.NoError(b, err)
	}
}
