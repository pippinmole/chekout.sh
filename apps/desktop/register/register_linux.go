//go:build linux

package register

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const desktopTemplate = `[Desktop Entry]
Name=chekout
Comment=Open GitHub PRs in your local IDE
Exec={{.Exe}} %u
Type=Application
NoDisplay=true
MimeType=x-scheme-handler/chekout;
`

// Register writes a .desktop file and registers the chekout:// MIME type.
func Register() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}

	desktopDir := filepath.Join(desktopAppsDir(), "applications")
	if err := os.MkdirAll(desktopDir, 0o755); err != nil {
		return fmt.Errorf("creating applications dir: %w", err)
	}

	desktopPath := filepath.Join(desktopDir, "chekout.desktop")
	f, err := os.Create(desktopPath)
	if err != nil {
		return fmt.Errorf("creating .desktop file: %w", err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("desktop").Parse(desktopTemplate))
	if err := tmpl.Execute(f, struct{ Exe string }{Exe: exe}); err != nil {
		return fmt.Errorf("writing .desktop file: %w", err)
	}

	// Register the MIME handler.
	cmd := exec.Command("xdg-mime", "default", "chekout.desktop", "x-scheme-handler/chekout")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("xdg-mime: %w\n%s", err, out)
	}

	// Update the MIME database.
	_ = exec.Command("update-desktop-database", desktopDir).Run()

	return nil
}

// ListenForURLs is a no-op on Linux — URLs arrive as CLI arguments handled
// in main.go via the IPC relay path.
func ListenForURLs(ch chan<- string) {}

func desktopAppsDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}
