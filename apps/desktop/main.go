package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chekout-sh/chekout/config"
	"github.com/chekout-sh/chekout/discover"
	"github.com/chekout-sh/chekout/handler"
	"github.com/chekout-sh/chekout/ipc"
	"github.com/chekout-sh/chekout/register"
	"github.com/chekout-sh/chekout/tray"
)

func main() {
	// Check if launched as a URL handler (Windows/Linux relay, or direct URL arg)
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if strings.HasPrefix(arg, "chekout://") {
			// Try to relay to an already-running instance via IPC.
			// If that fails (no running instance), handle directly.
			if err := ipc.Send(arg); err != nil {
				// No running instance — handle directly in this process.
				cfg, loadErr := config.Load()
				if loadErr != nil {
					fmt.Fprintf(os.Stderr, "failed to load config: %v\n", loadErr)
					os.Exit(1)
				}
				handler.ProcessURL(arg, cfg)
			}
			return
		}
	}

	// Normal startup — load or create config.
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Register URL scheme on first run.
	if !cfg.Registered {
		if err := register.Register(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to register URL scheme: %v\n", err)
		} else {
			cfg.Registered = true
			_ = config.Save(cfg)
		}
	}

	// Auto-discover repos on first run.
	if len(cfg.Repos) == 0 {
		discovered := discover.Repos()
		if len(discovered) > 0 {
			if cfg.Repos == nil {
				cfg.Repos = make(map[string]config.RepoEntry)
			}
			for k, v := range discovered {
				cfg.Repos[k] = v
			}
			_ = config.Save(cfg)
		}
	}

	// Channel for incoming chekout:// URLs (from IPC or Apple Events).
	urlCh := make(chan string, 16)

	// Start IPC server so other instances can relay URLs to us.
	go ipc.Serve(urlCh)

	// Register Apple Events handler (macOS) — no-op on other platforms.
	register.ListenForURLs(urlCh)

	// Handle incoming URLs in the background.
	go func() {
		for url := range urlCh {
			go handler.ProcessURL(url, cfg)
		}
	}()

	// Run the system tray (blocks until quit).
	tray.Run(cfg)
}
