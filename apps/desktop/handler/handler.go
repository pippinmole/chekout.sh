package handler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chekout-sh/chekout/config"
	"github.com/chekout-sh/chekout/ide"
	"github.com/gen2brain/beeep"
	"github.com/ncruces/zenity"
)

// chekoutJSON is the optional per-repo config file.
type chekoutJSON struct {
	IDE string `json:"ide"`
}

// ProcessURL parses a chekout:// URL and performs the full checkout + IDE open flow.
func ProcessURL(rawURL string, cfg *config.Config) {
	fmt.Printf("handler: processing %s\n", rawURL)

	u, err := url.Parse(rawURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "handler: invalid URL: %v\n", err)
		notify(false, "chekout", fmt.Sprintf("Invalid URL: %v", err))
		return
	}

	if u.Host != "open" && u.Path != "/open" {
		// Accept both chekout://open?... and chekout:///open?...
		if !(u.Host == "open" || strings.TrimPrefix(u.Path, "/") == "open") {
			fmt.Fprintf(os.Stderr, "handler: unknown URL format: %s\n", rawURL)
			notify(false, "chekout", "Unknown URL format: "+rawURL)
			return
		}
	}

	q := u.Query()
	repo := q.Get("repo")
	branch := q.Get("branch")

	if repo == "" || branch == "" {
		fmt.Fprintf(os.Stderr, "handler: missing repo or branch in URL\n")
		notify(false, "chekout", "Missing repo or branch in URL")
		return
	}

	fmt.Printf("handler: repo=%s branch=%s\n", repo, branch)

	// Normalise repo key: strip scheme so that "https://github.com/org/repo"
	// and "github.com/org/repo" both resolve to the same config key.
	// The host is preserved so github.com and gitlab.com repos stay distinct.
	repoKey := repo
	if parsed, err := url.Parse(repo); err == nil && parsed.Host != "" {
		repoKey = parsed.Host + parsed.Path
	}

	// Look up local path.
	entry, ok := cfg.GetRepo(repoKey)
	if !ok {
		fmt.Printf("handler: repo %s not in config — prompting user to locate\n", repoKey)
		localPath, err := pickDirectory(fmt.Sprintf("Locate local clone of %s", repoKey))
		if err != nil {
			fmt.Fprintf(os.Stderr, "handler: user did not pick a directory: %v\n", err)
			notify(false, "chekout", fmt.Sprintf("Could not find repo %s: %v", repoKey, err))
			return
		}
		fmt.Printf("handler: user selected path %s — saved to config\n", localPath)
		entry = config.RepoEntry{Path: localPath}
		cfg.SetRepo(repoKey, entry)
		_ = config.Save(cfg)
	} else {
		fmt.Printf("handler: found repo at %s\n", entry.Path)
	}

	localPath := entry.Path

	// Verify the path exists.
	if _, err := os.Stat(localPath); err != nil {
		fmt.Fprintf(os.Stderr, "handler: repo path not found: %s\n", localPath)
		notify(false, "chekout", fmt.Sprintf("Repo path not found: %s", localPath))
		return
	}

	// Optionally read .chekout.json from repo root for IDE override.
	selectedIDE := entry.IDE
	if selectedIDE == "" {
		selectedIDE = cfg.DefaultIDE
	}
	if override, err := readChekoutJSON(localPath); err == nil && override != "" {
		fmt.Printf("handler: .chekout.json overrides IDE to %s\n", override)
		selectedIDE = override
	}

	fmt.Printf("handler: using IDE %q\n", selectedIDE)

	// git fetch + checkout.
	fmt.Printf("handler: running git fetch + checkout for branch %s\n", branch)
	if err := gitCheckout(localPath, branch); err != nil {
		fmt.Fprintf(os.Stderr, "handler: git checkout failed: %v\n", err)
		notify(false, "chekout: checkout failed", err.Error())
		return
	}
	fmt.Printf("handler: checked out branch %s\n", branch)

	// Open in IDE.
	fmt.Printf("handler: opening %s in %s\n", localPath, selectedIDE)
	if err := ide.Open(selectedIDE, localPath); err != nil {
		fmt.Fprintf(os.Stderr, "handler: IDE open failed: %v\n", err)
		notify(false, "chekout: IDE open failed", err.Error())
		return
	}

	// Update recent opens.
	cfg.AddRecentOpen(repoKey)
	_ = config.Save(cfg)

	fmt.Printf("handler: success — opened %s (%s)\n", repoKey, branch)
	notify(true, "chekout", fmt.Sprintf("Opened %s (%s)", repoKey, branch))
}

// gitCheckout runs git fetch origin <branch> && git checkout <branch> in dir.
func gitCheckout(dir, branch string) error {
	fetch := exec.Command("git", "fetch", "origin", branch)
	fetch.Dir = dir
	if out, err := fetch.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	checkout := exec.Command("git", "checkout", branch)
	checkout.Dir = dir
	if out, err := checkout.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

// readChekoutJSON reads the IDE preference from .chekout.json in the repo root.
func readChekoutJSON(repoPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repoPath, ".chekout.json"))
	if err != nil {
		return "", err
	}
	var cj chekoutJSON
	if err := json.Unmarshal(data, &cj); err != nil {
		return "", err
	}
	return cj.IDE, nil
}

// pickDirectory shows a native folder picker dialog and returns the chosen path.
func pickDirectory(title string) (string, error) {
	_ = title // zenity uses the title on some platforms
	return zenity.SelectFile(
		zenity.Title(title),
		zenity.Directory(),
	)
}

// notify shows an OS notification. success=true uses beeep.Notify, false uses beeep.Alert.
func notify(success bool, title, body string) {
	if success {
		_ = beeep.Notify(title, body, appIconPath())
	} else {
		_ = beeep.Alert(title, body, appIconPath())
	}
}

// appIconPath returns the path to the app icon for notifications, or empty string.
func appIconPath() string {
	if runtime.GOOS == "darwin" {
		exe, err := os.Executable()
		if err != nil {
			return ""
		}
		// Walk up to find the .app bundle.
		dir := filepath.Dir(exe)
		for i := 0; i < 5; i++ {
			if strings.HasSuffix(dir, ".app") {
				return filepath.Join(dir, "Contents", "Resources", "AppIcon.icns")
			}
			dir = filepath.Dir(dir)
		}
	}
	return ""
}
