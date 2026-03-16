package ide

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
)

// ideScheme maps an IDE name to its deep-link URL scheme template.
// {path} is replaced with the URL-encoded local path.
var ideScheme = map[string]string{
	"vscode":     "vscode://file/{path}",
	"cursor":     "cursor://file/{path}",
	"intellij":   "jetbrains://idea/open?file={path}",
	"webstorm":   "jetbrains://web-storm/open?file={path}",
	"goland":     "jetbrains://go-land/open?file={path}",
	"phpstorm":   "jetbrains://php-storm/open?file={path}",
}

// Open fires the IDE deep link for the given IDE name and local path.
// Falls back to "vscode" if the IDE name is unrecognised.
func Open(ideName, localPath string) error {
	tmpl, ok := ideScheme[ideName]
	if !ok {
		tmpl = ideScheme["vscode"]
	}

	encodedPath := url.PathEscape(localPath)
	ideURL := replacePath(tmpl, encodedPath)

	return openURL(ideURL)
}

// replacePath substitutes {path} in the template with encoded.
func replacePath(tmpl, encoded string) string {
	result := ""
	for i := 0; i < len(tmpl); i++ {
		if i+6 <= len(tmpl) && tmpl[i:i+6] == "{path}" {
			result += encoded
			i += 5
		} else {
			result += string(tmpl[i])
		}
	}
	return result
}

// openURL fires the given URL using the OS default handler.
func openURL(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default: // linux and others
		cmd = exec.Command("xdg-open", u)
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("opening IDE URL %q: %w\n%s", u, err, out)
	}
	return nil
}
