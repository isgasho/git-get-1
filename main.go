// Command git-get clones Git repositories with an implicitly relative URL
// and always to a path under source root regardless of working directory.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mitchellh/go-homedir"
)

const defaultGit = "git"
const sourceRoot = "~/src"

const usage = `Usage:

git-get implements a command for git that clones to
`

// defaultPrefix is prefixed to implicitly relative clone URLs
var defaultPrefix = "git@github.com:"

func main() {
	logger := log.New(os.Stderr, "", 0)

	gitPath := defaultGit
	gitPath, err := exec.LookPath(gitPath)
	if err != nil {
		logger.Fatalf("git-get: failed to find git command %q in PATH", gitPath)
	}

	targetPath := sourceRoot
	targetPath, err = homedir.Expand(targetPath)
	cloneURL := expand(os.Args[1], defaultPrefix)
	td, err := targetDir(cloneURL)
	if err != nil {
		logger.Fatalf("git-get: %v", err)
	}

	// Replace current process with git
	gitArgv := []string{"git", "clone", cloneURL, filepath.Join(targetPath, td)}
	err = syscall.Exec(gitPath, gitArgv, os.Environ())
	if err != nil {
		logger.Fatalf("git-get: %v", err)
	}
}

// expand completes the implicitly relative clone URL s of form "project/repo" into
// absolute clone URL.
//
// If s does not have the required form, it is returned unchanged.
// unmodified.
func expand(s, defaultPrefix string) string {
	if strings.Contains(s, ":") {
		return s
	}
	parts := strings.SplitN(s, "/", 2)
	if len(parts) < 2 {
		return s
	}
	return defaultPrefix + parts[0] + "/" + parts[1]
}

// targetDir resolves the cloneURL to a relative directory path.
func targetDir(cloneURL string) (string, error) {
	cleanedCloneURL := strings.TrimSuffix(cloneURL, ".git")

	var hostname, path string

	// URLs like https:// and ssh://
	if parts := strings.SplitN(cleanedCloneURL, "://", 2); len(parts) == 2 {
		if addressParts := strings.SplitN(parts[1], "/", 2); len(addressParts) == 2 {
			hostname = addressParts[0]
			path = addressParts[1]
		} else {
			return "", fmt.Errorf(`expected path in URL, got %q`, cloneURL)
		}
		// URLs like user@hostname:project/repo
	} else if parts := strings.SplitN(cleanedCloneURL, ":", 2); len(parts) == 2 {
		hostname = parts[0]
		path = parts[1]
	} else {
		return "", fmt.Errorf(`expected PROJECT/REPO or absolute git clone URL, got %q`, cloneURL)
	}

	// ignore username
	parts := strings.Split(hostname, "@")
	hostname = parts[len(parts)-1]

	pathparts := strings.Split(path, "/")
	target := append([]string{hostname}, pathparts...)
	return strings.ToLower(filepath.Join(target...)), nil
}