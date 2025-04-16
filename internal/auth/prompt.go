package auth

import (
	"os"
	"os/exec"
)

func Authenticate(method string) bool {
	if method == "sudo" {
		return runSudoCheck()
	}
	return false
}

func runSudoCheck() bool {
	cmd := exec.Command("sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}
