// Package hooks contains various Webhook implementations
package hooks

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
)

// A Webhook knows how to parse a particular webhook format call
// in order to determine the Git repository URI it refers to.
type Webhook interface {
	// GetGitRepoUri determines the Git repository URI a webhook refers to
	GetGitRepoUri(*http.Request) (string, error)
}

// bodyHolder is a type implementing io.ReadCloser,
// used to make a copy of http.Request.Body
type bodyHolder struct {
	*bytes.Buffer
}

// Required to implement the io.Closer part
func (r bodyHolder) Close() error { return nil }

// getRequest form returns the form parameters from an HTTP request,
// including POST parameters, without modifying the request Body
func getRequestForm(req *http.Request) (url.Values, error) {
	// Take a copy of the request body
	bodyBuf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// Replace the request body, so we can parse the form from it
	req.Body = bodyHolder{bytes.NewBuffer(bodyBuf)}
	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	// Replace the request body once again for the next consumer
	req.Body = bodyHolder{bytes.NewBuffer(bodyBuf)}

	// Return the parsed form
	return req.Form, nil
}

// getRequestBody returns the HTTP request body as a string,
// without modifying the original request Body
func getRequestBody(req *http.Request) (string, error) {
	// Take a copy of the request body
	bodyBuf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	// Replace the request body for the next consumer
	req.Body = bodyHolder{bytes.NewBuffer(bodyBuf)}

	// Return the body to the caller
	return string(bodyBuf), nil
}
