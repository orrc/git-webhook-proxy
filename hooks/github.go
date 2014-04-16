package hooks

import (
	"encoding/json"
	"errors"
	"net/http"
)

// gitHubHookPayload reflects the parts of the GitHub
// webhook JSON structure that we are interested in
type gitHubHookPayload struct {
	Repository struct {
		Url string
	}
}

// A GitHubFormHook contains push info in JSON within an x-www-form-urlencoded POST body
type GitHubFormHook struct{}

func (h GitHubFormHook) GetGitRepoUri(req *http.Request) (string, error) {
	form, err := getRequestForm(req)
	if err != nil {
		return "", err
	}

	formValue := form.Get("payload")
	return getSshUriFromGitHubWebhookJson(formValue)
}

// A GitHubFormHook contains push info in JSON
type GitHubJsonHook struct{}

func (h GitHubJsonHook) GetGitRepoUri(req *http.Request) (string, error) {
	body, err := getRequestBody(req)
	if err != nil {
		return "", err
	}
	return getSshUriFromGitHubWebhookJson(body)
}

func getSshUriFromGitHubWebhookJson(body string) (string, error) {
	var payload gitHubHookPayload
	json.Unmarshal([]byte(body), &payload)
	repoHttpUrl := payload.Repository.Url
	if repoHttpUrl == "" {
		return "", errors.New("No URL found in webhook payload")
	}

	return getSshUriForUrl(repoHttpUrl)
}
