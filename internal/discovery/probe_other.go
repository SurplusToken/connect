//go:build !windows

package discovery

import "os/exec"

func configureProbe(cmd *exec.Cmd) {}
