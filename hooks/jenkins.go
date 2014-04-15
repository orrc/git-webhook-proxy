package hooks

import (
	"net/http"
	"net/url"
)

// A JenkinsHook contains the repository URI in the "url" GET parameter
type JenkinsHook struct{}

func (h JenkinsHook) GetGitRepoUri(req *http.Request) (string, error) {
	params, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return "", err
	}
	return params.Get("url"), nil
}
