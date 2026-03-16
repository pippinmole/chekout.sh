package register

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework CoreServices

#include <stdlib.h>

// Declarations only — definitions live in url_handler_darwin.m which CGO
// compiles automatically as part of this package.
extern void RegisterAppleEventHandler(void);
extern char *PopURL(void);
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"
)

// Register uses lsregister to register the app bundle with LaunchServices.
func Register() error {
	bundlePath := findBundle()
	if bundlePath == "" {
		return fmt.Errorf("could not locate .app bundle — skipping lsregister")
	}

	lsregister := "/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/" +
		"LaunchServices.framework/Versions/A/Support/lsregister"

	cmd := lsregisterCmd(lsregister, "-f", bundlePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lsregister: %w\n%s", err, out)
	}
	return nil
}

// ListenForURLs registers the Apple Events handler and starts a polling
// goroutine that forwards received URLs to ch.
func ListenForURLs(ch chan<- string) {
	C.RegisterAppleEventHandler()

	go func() {
		for {
			cstr := C.PopURL()
			if cstr != nil {
				url := C.GoString(cstr)
				C.free(unsafe.Pointer(cstr))
				if url != "" {
					ch <- url
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

// findBundle walks up from the executable path looking for a .app bundle.
func findBundle() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	for i := 0; i < 6; i++ {
		if strings.HasSuffix(dir, ".app") {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}
