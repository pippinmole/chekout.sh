package tray

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/chekout-sh/chekout/config"
	"github.com/getlantern/systray"
)

// Run initialises and blocks on the system tray. Call from the main goroutine.
func Run(cfg *config.Config) {
	systray.Run(
		func() { onReady(cfg) },
		func() { /* onExit — nothing to do */ },
	)
}

func onReady(cfg *config.Config) {
	systray.SetIcon(iconData())
	fmt.Println("tray icon set")
	systray.SetTooltip("chekout")
	fmt.Println("tray ready")

	// Recent opens section (up to 5, greyed-out display items).
	addRecentOpens(cfg)

	systray.AddSeparator()

	// Open config file.
	mConfig := systray.AddMenuItem("Open Configg", "Edit ~/.config/chekout/config.json")
	go func() {
		for range mConfig.ClickedCh {
			openConfig()
		}
	}()

	systray.AddSeparator()

	// Quit.
	mQuit := systray.AddMenuItem("Quit chekout", "Quit the chekout app")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

// addRecentOpens adds the last 5 recent opens as disabled menu items.
func addRecentOpens(cfg *config.Config) {
	recents := cfg.RecentOpens
	if len(recents) == 0 {
		item := systray.AddMenuItem("No recent opens", "")
		item.Disable()
		return
	}
	for _, r := range recents {
		label := r
		if len(label) > 50 {
			label = "…" + label[len(label)-47:]
		}
		item := systray.AddMenuItem(label, r)
		item.Disable()
	}
}

// openConfig opens the config file with the OS default editor.
func openConfig() {
	path := config.Path()
	// Ensure the file exists so the editor doesn't complain.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = config.Save(&config.Config{
			DefaultIDE:  "vscode",
			Repos:       make(map[string]config.RepoEntry),
			RecentOpens: []string{},
		})
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("notepad", path)
	default:
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "xdg-open"
		}
		cmd = exec.Command(editor, path)
	}
	_ = cmd.Start()
}
