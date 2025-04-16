package main

import (
	"fmt"
	"os"

	"ruin/internal/auth"
	"ruin/internal/config"
	"ruin/internal/context"
	"ruin/internal/executor"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Printf("Loaded config: %+v\n", cfg)

	currentContext, err := context.GetCurrentContext()
	if err != nil {
		fmt.Println("Error determining current context:", err)
		os.Exit(1)
	}

	if cfg.Kubectl.IsProtected(currentContext) {
		if !auth.CheckAuthCache(currentContext, cfg.Kubectl.GracePeriodSeconds) {
			if !auth.Authenticate(cfg.Kubectl.AuthMethod) {
				fmt.Println("Authentication failed.")
				os.Exit(1)
			}
			auth.TouchAuthCache(currentContext)
		}
	}

	executor.RunKubectl(os.Args[1:])
}
