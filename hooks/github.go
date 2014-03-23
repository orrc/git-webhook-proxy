package hooks

import (
	"fmt"
	"net/url"
)

type gitHubHookPayload struct {
	Repository gitHubHookRepository
}

type gitHubHookRepository struct {
	Url string
}

func getSshUriForGitHubUrl(httpUrl string) (string, error) {
	u, err := url.Parse(httpUrl)
	return fmt.Sprintf("git@%s:%s.git", u.Host, u.Path[1:]), err
}
