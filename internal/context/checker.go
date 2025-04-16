package context

import (
    "os/exec"
    "strings"
)

func GetCurrentContext() (string, error) {
    out, err := exec.Command("kubectl", "config", "current-context").Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}
