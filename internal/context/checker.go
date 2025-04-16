package context

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func GetCurrentContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")

	// Pass full user environment
	cmd.Env = os.Environ()

	// Capture output
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("kubectl failed: %v\nstderr: %s", err, stderr.String())
	}

	return out.String(), nil
}
