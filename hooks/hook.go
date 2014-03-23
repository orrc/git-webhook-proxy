package hooks

import "net/http"

// A Webhook knows how to parse a particular webhook format call
// in order to determine the Git repository URI it refers to.
type Webhook interface {
	// Determines the Git repository URI from a webhook request
	GetGitRepoUri(*http.Request) (string, error)
}
