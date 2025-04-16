package executor

import (
    "os"
    "os/exec"
)

func RunKubectl(args []string) {
    cmd := exec.Command("kubectl", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    cmd.Run()
}
