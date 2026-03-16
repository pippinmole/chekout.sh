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
	u, err := url.Parse(rawURL)
	if err != nil {
		notify(false, "chekout", fmt.Sprintf("Invalid URL: %v", err))
		return
	}

	if u.Host != "open" && u.Path != "/open" {
		// Accept both chekout://open?... and chekout:///open?...
		if !(u.Host == "open" || strings.TrimPrefix(u.Path, "/") == "open") {
			notify(false, "chekout", "Unknown URL format: "+rawURL)
			return
		}
	}

	q := u.Query()
	repo := q.Get("repo")
	branch := q.Get("branch")

	if repo == "" || branch == "" {
		notify(false, "chekout", "Missing repo or branch in URL")
		return
	}

	// Normalise repo key: strip leading "github.com/" if accidentally doubled,
	// but keep the full "github.com/org/repo" form.
	repoKey := repo

	// Look up local path.
	entry, ok := cfg.GetRepo(repoKey)
	if !ok {
		// Ask the user to locate the repo.
		localPath, err := pickDirectory(fmt.Sprintf("Locate local clone of %s", repoKey))
		if err != nil {
			notify(false, "chekout", fmt.Sprintf("Could not find repo %s: %v", repoKey, err))
			return
		}
		entry = config.RepoEntry{Path: localPath}
		cfg.SetRepo(repoKey, entry)
		_ = config.Save(cfg)
	}

	localPath := entry.Path

	// Verify the path exists.
	if _, err := os.Stat(localPath); err != nil {
		notify(false, "chekout", fmt.Sprintf("Repo path not found: %s", localPath))
		return
	}

	// Optionally read .chekout.json from repo root for IDE override.
	selectedIDE := entry.IDE
	if selectedIDE == "" {
		selectedIDE = cfg.DefaultIDE
	}
	if override, err := readChekoutJSON(localPath); err == nil && override != "" {
		selectedIDE = override
	}

	// git fetch + checkout.
	if err := gitCheckout(localPath, branch); err != nil {
		notify(false, "chekout: checkout failed", err.Error())
		return
	}

	// Open in IDE.
	if err := ide.Open(selectedIDE, localPath); err != nil {
		notify(false, "chekout: IDE open failed", err.Error())
		return
	}

	// Update recent opens.
	cfg.AddRecentOpen(repoKey)
	_ = config.Save(cfg)

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
