package discover

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/chekout-sh/chekout/config"
)

// searchDirs are the well-known root directories searched for git repos.
var searchDirs = []string{
	"code", "projects", "dev", "workspace", "src",
}

// Repos walks well-known directories under $HOME looking for git repos whose
// remote origin points to github.com. Returns a map of "github.com/org/repo"
// keys to RepoEntry values.
func Repos() map[string]config.RepoEntry {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	result := make(map[string]config.RepoEntry)

	for _, dir := range searchDirs {
		root := filepath.Join(home, dir)
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		// Walk up to 3 levels deep — repos are rarely nested deeper.
		_ = walkDepth(root, 3, func(path string, d os.DirEntry) error {
			if d.Name() == ".git" && d.IsDir() {
				repoDir := filepath.Dir(path)
				key, err := parseOrigin(filepath.Join(path, "config"))
				if err != nil || key == "" {
					return nil
				}
				result[key] = config.RepoEntry{Path: repoDir}
				// Don't descend into .git subdirectories.
				return filepath.SkipDir
			}
			return nil
		})
	}

	return result
}

// walkDepth is like filepath.WalkDir but stops descending beyond maxDepth.
func walkDepth(root string, maxDepth int, fn func(path string, d os.DirEntry) error) error {
	return walkDepthInner(root, 0, maxDepth, fn)
}

func walkDepthInner(path string, depth, maxDepth int, fn func(string, os.DirEntry) error) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil // skip unreadable directories
	}

	for _, e := range entries {
		fullPath := filepath.Join(path, e.Name())
		if err := fn(fullPath, e); err == filepath.SkipDir {
			continue
		}
		if e.IsDir() && depth < maxDepth {
			_ = walkDepthInner(fullPath, depth+1, maxDepth, fn)
		}
	}
	return nil
}

// parseOrigin parses a git config file and extracts the GitHub repo key
// (e.g. "github.com/org/repo") from the [remote "origin"] section.
func parseOrigin(gitConfigPath string) (string, error) {
	f, err := os.Open(gitConfigPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	inOrigin := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detect section headers.
		if strings.HasPrefix(line, "[") {
			inOrigin = strings.EqualFold(line, `[remote "origin"]`)
			continue
		}

		if inOrigin && strings.HasPrefix(line, "url") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			rawURL := strings.TrimSpace(parts[1])
			return githubKeyFromURL(rawURL), nil
		}
	}

	return "", nil
}

// githubKeyFromURL converts a git remote URL into a "github.com/org/repo" key.
// Handles both SSH (git@github.com:org/repo.git) and HTTPS forms.
func githubKeyFromURL(rawURL string) string {
	rawURL = strings.TrimSuffix(rawURL, ".git")

	// HTTPS: https://github.com/org/repo
	if strings.HasPrefix(rawURL, "https://") {
		rawURL = strings.TrimPrefix(rawURL, "https://")
		if strings.HasPrefix(rawURL, "github.com/") {
			return rawURL
		}
		return ""
	}

	// HTTP: http://github.com/org/repo
	if strings.HasPrefix(rawURL, "http://") {
		rawURL = strings.TrimPrefix(rawURL, "http://")
		if strings.HasPrefix(rawURL, "github.com/") {
			return rawURL
		}
		return ""
	}

	// SSH: git@github.com:org/repo
	if strings.HasPrefix(rawURL, "git@github.com:") {
		return "github.com/" + strings.TrimPrefix(rawURL, "git@github.com:")
	}

	return ""
}
