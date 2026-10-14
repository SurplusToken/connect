//go:build windows

package discovery

import (
	"os/exec"
	"syscall"
)

func configureProbe(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
