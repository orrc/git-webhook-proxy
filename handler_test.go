package main

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type repoTestData struct {
	input  string
	expect string
}

func TestGetDirNameForRepo(t *testing.T) {
	data := []repoTestData{
		{"ssh://host.xz/path/to/repo.git/", "host.xz/path/to/repo.git"},
		{"ssh://host.xz:22/path/to/repo.git/", "host.xz/path/to/repo.git"},
		{"ssh://user@host.xz:22/path/to/repo.git/", "host.xz/path/to/repo.git"},

		{"git://host.xz/path/to/repo.git/", "host.xz/path/to/repo.git"},
		{"git://host.xz:22/path/to/repo.git/", "host.xz/path/to/repo.git"},
		{"git://user@host.xz:9418/path/to/repo.git/", "host.xz/path/to/repo.git"},

		{"http://git.example.com/user/My-Repo", "git.example.com/user/my-repo.git"},
		{"https://git.example.com:8443/user/My-Repo", "git.example.com/user/my-repo.git"},
		{"https://scm@git.example.com:8443/user/My-Repo.git", "git.example.com/user/my-repo.git"},

		{"example.com:/a/b/c/", "example.com/a/b/c.git"},
		{"git@github.com:example/testing", "github.com/example/testing.git"},
		{"git@git.assembla.com:foo-bar-app.git", "git.assembla.com/foo-bar-app.git"},
	}

	for _, d := range data {
		Convey(fmt.Sprintf("Dir for '%s' should be '%s'", d.input, d.expect), t, func() {
			result := getDirNameForRepo(d.input)
			So(result, ShouldEqual, d.expect)
		})
	}
}
