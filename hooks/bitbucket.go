package hooks

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// bitbucketHookPayload reflects the parts of the Bitbucket
// webhook JSON structure that we are interested in
type bitbucketHookPayload struct {
	BaseUrl    string `json:"canon_url"`
	Repository struct {
		RepoPath string `json:"absolute_url"`
	}
}

// A BitbucketHook contains push info in JSON within an x-www-form-urlencoded POST body
type BitbucketHook struct{}

func (h BitbucketHook) GetGitRepoUri(req *http.Request) (string, error) {
	form, err := getRequestForm(req)
	if err != nil {
		return "", err
	}

	formValue := form.Get("payload")

	var payload bitbucketHookPayload
	json.Unmarshal([]byte(formValue), &payload)
	repoHttpUrl := strings.TrimSuffix(payload.BaseUrl+payload.Repository.RepoPath, "/")
	if repoHttpUrl == "" {
		return "", errors.New("No URL found in webhook payload")
	}

	return getSshUriForUrl(repoHttpUrl)
}
