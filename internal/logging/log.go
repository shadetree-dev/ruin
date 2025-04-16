package logging

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"ruin/internal/config"
)

type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	User      string   `json:"user"`
	Context   string   `json:"context"`
	Command   []string `json:"command"`
	Auth      string   `json:"auth"`
	Result    string   `json:"result"`
	MAC       string   `json:"mac"`
}

func Log(cfg config.Config, context, user string, cmd []string, success bool) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User:      user,
		Context:   context,
		Command:   cmd,
		Auth:      cfg.Kubectl.AuthMethod,
		Result:    map[bool]string{true: "allowed", false: "denied"}[success],
	}

	// MAC generation (could be salted with env key)
	mac := hmac.New(sha256.New, []byte("ruin-secret-key"))
	data, _ := json.Marshal(entry)
	mac.Write(data)
	entry.MAC = hex.EncodeToString(mac.Sum(nil))

	jsonData, _ := json.Marshal(entry)

	if cfg.Audit.FallbackLocal {
		writeLocalLog(jsonData, cfg)
	}

	if cfg.Audit.ForwardURL != "" {
		go forwardLog(jsonData, cfg)
	}
}

func writeLocalLog(data []byte, cfg config.Config) {
	path := cfg.Audit.LogPath
	if path == "" {
		path = "/var/log/ruin.log"
	}

	fallback := filepath.Join(os.Getenv("HOME"), ".ruin/ruin.log")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		path = fallback
		os.MkdirAll(filepath.Dir(path), 0700)
		f, _ = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	}
	defer f.Close()

	syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	stat, _ := f.Stat()
	if stat.Size() > cfg.Audit.MaxLogSizeBytes {
		_ = f.Truncate(0)
		f.Seek(0, 0)
	}

	_, _ = f.WriteString(string(data) + "\n")
}

func forwardLog(data []byte, cfg config.Config) {
	_, err := http.Post(cfg.Audit.ForwardURL, "application/json", bytes.NewBuffer(data))
	if err != nil && cfg.Audit.FallbackLocal {
		fmt.Fprintf(os.Stderr, "Failed to forward log: %v\n", err)
	}
}
