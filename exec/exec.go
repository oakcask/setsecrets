//+build !windows

package exec

import (
	"os/exec"
	"syscall"
)

func Exec(argv0 string, argv []string, envv []string) error {
	argv0, e := exec.LookPath(argv0)
	if e != nil {
		return e
	}

	return syscall.Exec(argv0, argv, envv)
}
