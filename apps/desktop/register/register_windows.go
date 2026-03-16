//go:build windows

package register

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

// Register writes the chekout:// URL scheme handler into the Windows registry.
func Register() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	// HKEY_CURRENT_USER\Software\Classes\chekout (no admin required)
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\chekout`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("creating registry key: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue("", "URL:chekout Protocol"); err != nil {
		return err
	}
	if err := key.SetStringValue("URL Protocol", ""); err != nil {
		return err
	}

	// HKEY_CURRENT_USER\Software\Classes\chekout\shell\open\command
	cmdKey, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\chekout\shell\open\command`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("creating command key: %w", err)
	}
	defer cmdKey.Close()

	cmdValue := fmt.Sprintf(`"%s" "%%1"`, exe)
	return cmdKey.SetStringValue("", cmdValue)
}

// ListenForURLs is a no-op on Windows — URLs arrive as CLI arguments handled
// in main.go via the IPC relay path.
func ListenForURLs(_ chan<- string) {}
