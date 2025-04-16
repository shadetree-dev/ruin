package test

import (
    "testing"
    "ruin/internal/config"
)

func TestProtectedContext(t *testing.T) {
    cfg := config.Config{
        Kubectl: config.KubectlConfig{
            ProtectedContexts: []string{"prod"},
        },
    }

    if !cfg.Kubectl.IsProtected("prod") {
        t.Fatal("Expected context 'prod' to be protected")
    }

    if cfg.Kubectl.IsProtected("dev") {
        t.Fatal("Expected context 'dev' to be unprotected")
    }
}
