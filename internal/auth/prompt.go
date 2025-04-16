package auth

import (
	"fmt"
	"os"
	"os/exec"
)

func Authenticate(method string) bool {
	switch method {
	case "sudo":
		return runSudoCheck()
	default:
		fmt.Printf("Unsupported auth method: %s\n", method)
		return false
	}
}

func runSudoCheck() bool {
	cmd := exec.Command("sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil
}
