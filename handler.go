package main

import (
	"errors"
	"fmt"
	"github.com/orrc/git-webhook-proxy/hooks"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
)

type Handler struct {
	gitPath       string
	mirrorRootDir string
	remoteUrl     string
	proxy         http.Handler
	requests      map[string]*sync.Mutex
}

func NewHandler(gitPath, mirrorRootDir, remoteUrl string) (h *Handler, err error) {
	backendUrl, err := url.Parse(remoteUrl)
	proxy := httputil.NewSingleHostReverseProxy(backendUrl)

	// Ensure we send the correct Host header to the backend
	defaultDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		defaultDirector(req)
		req.Host = backendUrl.Host
	}

	h = &Handler{
		gitPath:       gitPath,
		mirrorRootDir: mirrorRootDir,
		remoteUrl:     remoteUrl,
		proxy:         proxy,
		requests:      make(map[string]*sync.Mutex),
	}
	return
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Log request
	log.Printf("Incoming webhook from %s %s %s", req.RemoteAddr, req.Method, req.URL)

	// Determine which handler to use
	// TODO: This won't work well for e.g. "/jenkins/git/notifyCommit"
	var hookType hooks.Webhook
	switch req.URL.Path {
	case "/git/notifyCommit":
		hookType = hooks.JenkinsHook{}
	case "/github-webhook/":
		hookType = hooks.GitHubFormHook{}
	default:
		log.Println("No hook handler found!")
		http.NotFound(w, req)
		return
	}

	// Parse the Git repo URI from the webhook request
	repoUri, err := hookType.GetGitRepoUri(req)
	if err != nil {
		msg := fmt.Sprintf("%s returned error: %s", reflect.TypeOf(hookType), err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	if repoUri == "" {
		msg := fmt.Sprintf("%s could not determine the repository URL from this request", reflect.TypeOf(hookType))
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// Check whether we're already working on updating this repo
	// TODO: Coalesce multiple blocked requests
	if _, exists := h.requests[repoUri]; !exists {
		h.requests[repoUri] = &sync.Mutex{}
	}
	lock := h.requests[repoUri]
	lock.Lock()
	defer lock.Unlock()

	// Clone or mirror the repo
	// TODO: Test what happens if the HTTP client disappears in the middle of a long clone
	err = h.updateOrCloneRepoMirror(repoUri)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Proxy the original webhook request to the backend
	log.Printf("Proxying webhook request to %s/%s\n", h.remoteUrl, req.URL)
	h.proxy.ServeHTTP(w, req)
}

func (h *Handler) updateOrCloneRepoMirror(repoUri string) error {
	// Check whether we have cloned this repo already
	repoPath := h.getMirrorPathForRepo(repoUri)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// TODO: Also need to somehow detect whether a directory has a full clone, or failed...
		err = h.cloneRepo(repoUri)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to clone %s: %s", repoUri, err.Error()))
		}
		return err
	}

	// If we already have clone the repo, ensure that it is up-to-date
	log.Printf("Updating mirror at %s", repoPath)
	cmd := exec.Command(h.gitPath, "remote", "update", "-p")
	cmd.Dir = repoPath
	err := cmd.Run()
	if err == nil {
		log.Printf("Successfully updated %s", repoPath)

		// Also run "git gc", if required, to clean up afterwards
		cmd := exec.Command(h.gitPath, "gc", "--aggressive", "--auto")
		cmd.Dir = repoPath

		// But we don't really care about the outcome
		cmd.Run()
	} else {
		err = fmt.Errorf("Failed to update %s: %s", repoPath, err.Error())
	}
	return err
}

func (h *Handler) cloneRepo(repoUri string) error {
	// Ensure the mirror root directory exists
	err := os.MkdirAll(h.mirrorRootDir, 0700)
	if err != nil {
		return err
	}

	// Delete the directory if cloning fails
	defer func() {
		if err != nil {
			os.Remove(h.getMirrorPathForRepo(repoUri))
		}
	}()

	// TODO: We may need to transform incoming repo URIs to add user credentials so they can be cloned
	log.Printf("Cloning %s to %s", repoUri, h.mirrorRootDir)
	cmd := exec.Command(h.gitPath, "clone", "--mirror", repoUri, getDirNameForRepo(repoUri))
	cmd.Dir = h.mirrorRootDir
	err = cmd.Run()
	if err == nil {
		log.Printf("Successfully cloned %s", repoUri)
	}
	return err
}

func (h *Handler) getMirrorPathForRepo(repoUri string) string {
	return fmt.Sprintf("%s/%s", h.mirrorRootDir, getDirNameForRepo(repoUri))
}

func getDirNameForRepo(repoUri string) string {
	repoUri = strings.TrimSpace(repoUri)
	repoUri = strings.TrimSuffix(repoUri, "/")
	repoUri = strings.TrimSuffix(repoUri, ".git")
	repoUri = strings.ToLower(repoUri)

	if strings.Contains(repoUri, "://") {
		uri, _ := url.Parse(repoUri)
		if i := strings.Index(uri.Host, ":"); i != -1 {
			uri.Host = uri.Host[:i]
		}
		return fmt.Sprintf("%s/%s.git", uri.Host, uri.Path[1:])
	}

	if i := strings.Index(repoUri, "@"); i != -1 {
		repoUri = repoUri[i+1:]
	}
	repoUri = strings.Replace(repoUri, ":", "/", 1)
	repoUri = strings.Replace(repoUri, "//", "/", -1)
	return repoUri + ".git"
}
