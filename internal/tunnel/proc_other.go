//go:build !windows

package tunnel

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {
	// No-op on non-Windows platforms
}
