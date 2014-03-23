package hooks

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type JenkinsHook struct{}

func (h JenkinsHook) GetGitRepoUri(req *http.Request) (string, error) {
	params, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return "", err
	}
	return params.Get("url"), nil
}

type JenkinsGitHubHook struct{}

func (h JenkinsGitHubHook) GetGitRepoUri(req *http.Request) (string, error) {
	// TODO: This won't work yet, as we need to make a copy of the request body
	if err := req.ParseForm(); err != nil {
		return "", err
	}

	// TODO: Detect mime-type and decode accordingly
	formValue := req.Form.Get("payload")

	var payload gitHubHookPayload
	json.Unmarshal([]byte(formValue), &payload)
	repoHttpUrl := payload.Repository.Url

	return getSshUriForGitHubUrl(repoHttpUrl)
}
