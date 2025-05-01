package awareness

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"ruin/internal/config"
)

var writeOps = map[string]bool{
	"apply": true, "create": true, "delete": true, "edit": true,
	"patch": true, "replace": true, "scale": true, "rollout": true,
	"drain": true,
}

func ShouldSkipAwareness(cfg config.Config, args []string) bool {
	override := os.Getenv("RUIN_AWARENESS_MODE")
	if override == "none" {
		fmt.Fprintf(os.Stderr, "[ruin] awareness disabled via env override\n")
		return true
	}
	if cfg.Kubectl.Awareness.Mode == "none" {
		fmt.Fprintf(os.Stderr, "[ruin] awareness disabled via config\n")
		return true
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "[ruin] no command args passed, skipping awareness\n")
		return true
	}

	primary := args[0]
	if primary == "auth" && len(args) > 1 {
		primary = args[1]
	}

	skip := cfg.Kubectl.Awareness.OnlyOnWrite && !writeOps[primary]

	fmt.Fprintf(os.Stderr, "[ruin] awareness check: command='%s', skip=%v\n", primary, skip)

	return skip
}
func Run(cfg config.Config, context string, args []string) bool {
	mode := os.Getenv("RUIN_AWARENESS_MODE")
	if mode == "" {
		mode = cfg.Kubectl.Awareness.Mode
	}

	switch mode {
	case "prompt":
		fmt.Printf("⚠️  Using protected context '%s'. Proceed with '%s'? [y/N]: ", context, strings.Join(args, " "))
		var response string
		fmt.Scanln(&response)
		return strings.ToLower(response) == "y"

	case "pause":
		sec := cfg.Kubectl.Awareness.PauseSeconds
		if envPause := os.Getenv("RUIN_PAUSE_SECONDS"); envPause != "" {
			if s, err := strconv.Atoi(envPause); err == nil && s >= 0 && s <= 30 {
				sec = s
			}
		}

		fmt.Printf("\n⚠️  Using protected context '%s'\n", context)
		fmt.Printf("🛑 Executing '%s' in %d seconds... (press 'y' + Enter to proceed immediately or Ctrl+C to cancel)\n", strings.Join(args, " "), sec)

		reader := bufio.NewReader(os.Stdin)
		inputCh := make(chan string, 1)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT)

		go func() {
			text, _ := reader.ReadString('\n')
			inputCh <- strings.TrimSpace(text)
		}()

		for i := sec; i > 0; i-- {
			fmt.Printf("\r%d seconds remaining... ", i)
			time.Sleep(time.Second)

			select {
			case txt := <-inputCh:
				if strings.ToLower(txt) == "y" {
					fmt.Print("\n✅ Proceeding immediately.\n")
					return true
				}
			case <-sigCh:
				fmt.Println("\n❌ Cancelled.")
				os.Exit(1)
			default:
			}
		}

		fmt.Println("\n⏳ Proceeding after timeout.")
		return true

	default:
		return true
	}
}
