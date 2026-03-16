package register

import "os/exec"

// lsregisterCmd returns an exec.Cmd for lsregister with the given args.
// Defined here so it can call os/exec without CGO header conflicts.
func lsregisterCmd(path string, args ...string) *exec.Cmd {
	return exec.Command(path, args...)
}
