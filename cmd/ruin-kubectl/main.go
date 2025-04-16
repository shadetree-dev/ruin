package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"ruin/internal/auth"
	"ruin/internal/awareness"
	"ruin/internal/config"
	"ruin/internal/context"
	"ruin/internal/executor"
	"ruin/internal/logging"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runInit()
		return
	}

	cfg := config.LoadConfig()
	checkForNewContexts()

	currentContext, err := context.GetCurrentContext()
	if err != nil {
		fmt.Println("Error determining current context:", err)
		os.Exit(1)
	}

	if cfg.Kubectl.IsProtected(currentContext) {
		if !auth.CheckAuthCache(currentContext, cfg.Kubectl.GracePeriodSeconds) {
			if !auth.Authenticate(cfg.Kubectl.AuthMethod) {
				fmt.Println("Authentication failed.")
				logging.Log(cfg, currentContext, os.Getenv("USER"), os.Args[1:], false)
				os.Exit(1)
			}
			auth.TouchAuthCache(currentContext)
		}

		if !awareness.ShouldSkipAwareness(cfg, os.Args[1:]) {
			if !awareness.Run(cfg, currentContext, os.Args[1:]) {
				fmt.Println("Command cancelled.")
				os.Exit(0)
			}
		}
	}

	logging.Log(cfg, currentContext, os.Getenv("USER"), os.Args[1:], true)
	executor.RunKubectl(os.Args[1:])
}

func runInit() {
	reader := bufio.NewReader(os.Stdin)

	useEtc := false
	if os.Geteuid() != 0 {
		fmt.Print("🛡  Not running as root. Relaunch with sudo to configure /etc/ruin? [y/N]: ")
		resp, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(resp)) == "y" {
			cmd := exec.Command("sudo", os.Args[0], "init")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			os.Exit(0)
		}
	} else {
		useEtc = true
	}

	out, err := exec.Command("kubectl", "config", "get-contexts", "-o", "name").Output()
	if err != nil {
		fmt.Println("❌ Failed to list contexts:", err)
		os.Exit(1)
	}
	contexts := strings.Split(strings.TrimSpace(string(out)), "\n")
	fmt.Println("🔧 Found Kubernetes contexts:")
	for i, c := range contexts {
		fmt.Printf("  [%d] %s\n", i+1, c)
	}

	fmt.Print("\nEnter comma-separated indexes or '*' for all [*]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var protected []string
	if input == "" || input == "*" {
		protected = []string{"*"}
	} else {
		indices := strings.Split(input, ",")
		for _, idx := range indices {
			n, err := strconv.Atoi(strings.TrimSpace(idx))
			if err == nil && n > 0 && n <= len(contexts) {
				protected = append(protected, contexts[n-1])
			}
		}
	}

	fmt.Print("Awareness mode? (pause/prompt/none) [pause]: ")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = "pause"
	}

	pauseSeconds := "5"
	if mode == "pause" {
		fmt.Print("Pause seconds? [5]: ")
		ps, _ := reader.ReadString('\n')
		ps = strings.TrimSpace(ps)
		if ps != "" {
			pauseSeconds = ps
		}
	}

	fmt.Print("Grace period for sudo auth? [300]: ")
	grace, _ := reader.ReadString('\n')
	grace = strings.TrimSpace(grace)
	if grace == "" {
		grace = "300"
	}

	fmt.Print("Log path? [/var/log/ruin.log]: ")
	logPath, _ := reader.ReadString('\n')
	logPath = strings.TrimSpace(logPath)
	if logPath == "" {
		logPath = "/var/log/ruin.log"
	}

	configYAML := fmt.Sprintf(`
kubectl:
  protected_contexts:
%s
  auth_method: sudo
  grace_period_seconds: %s
  awareness_prompt:
    mode: %s
    pause_seconds: %s
    only_on_write: true

audit:
  log_path: %s
  forward_url: ""
  fallback_local: true
  max_log_size_bytes: 5242880
`, formatListYAML("    - ", protected), grace, mode, pauseSeconds, logPath)

	var path string
	if useEtc {
		path = "/etc/ruin/config"
		os.MkdirAll("/etc/ruin", 0755)
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".ruin/config")
		os.MkdirAll(filepath.Join(os.Getenv("HOME"), ".ruin"), 0700)
	}

	if err := os.WriteFile(path, []byte(configYAML), 0644); err != nil {
		fmt.Println("❌ Could not write config:", err)
		os.Exit(1)
	}

	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); err == nil {
		defer f.Close()
		fmt.Println("📓 Log file created or writable at:", logPath)
	} else {
		fmt.Println("⚠️  Could not create log file:", err)
	}

	fmt.Println("✅ Config written to:", path)
}

func formatListYAML(prefix string, values []string) string {
	var sb strings.Builder
	for _, v := range values {
		sb.WriteString(fmt.Sprintf("%s%s\n", prefix, v))
	}
	return sb.String()
}

func checkForNewContexts() {
	out, err := exec.Command("kubectl", "config", "get-contexts", "-o", "name").Output()
	if err != nil {
		return // silently skip if kubectl fails
	}

	hash := sha256.Sum256(out)
	hashHex := fmt.Sprintf("%x", hash[:])

	cachePath := filepath.Join(os.Getenv("HOME"), ".ruin/.kube_contexts.hash")
	_ = os.MkdirAll(filepath.Dir(cachePath), 0700)

	existing, err := os.ReadFile(cachePath)
	if err == nil && strings.TrimSpace(string(existing)) == hashHex {
		return // same, do nothing
	}

	// If we're here, hash changed
	fmt.Println("📣 New Kubernetes contexts detected.")
	fmt.Println("👉 Run `ruin-kubectl init` to update protected contexts.")
	_ = os.WriteFile(cachePath, []byte(hashHex), 0600)
}
