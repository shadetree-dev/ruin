package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type AwarenessPrompt struct {
	Mode         string `yaml:"mode"`          // "pause", "prompt", "none"
	PauseSeconds int    `yaml:"pause_seconds"` // only applies if mode is "pause"
	OnlyOnWrite  bool   `yaml:"only_on_write"`
}

type KubectlConfig struct {
	ProtectedContexts  []string        `yaml:"protected_contexts"`
	AuthMethod         string          `yaml:"auth_method"`
	GracePeriodSeconds int             `yaml:"grace_period_seconds"`
	Awareness          AwarenessPrompt `yaml:"awareness_prompt"`
}

type AuditConfig struct {
	ForwardURL      string `yaml:"forward_url"`
	FallbackLocal   bool   `yaml:"fallback_local"`
	MaxLogSizeBytes int64  `yaml:"max_log_size_bytes"`
	LogPath         string `yaml:"log_path"`
}

type Config struct {
	Kubectl KubectlConfig `yaml:"kubectl"`
	Audit   AuditConfig   `yaml:"audit"`
}

var builtInDefaults = Config{
	Kubectl: KubectlConfig{
		ProtectedContexts:  []string{"*"},
		AuthMethod:         "sudo",
		GracePeriodSeconds: 300,
		Awareness: AwarenessPrompt{
			Mode:         "pause",
			PauseSeconds: 5,
			OnlyOnWrite:  true,
		},
	},
	Audit: AuditConfig{
		FallbackLocal:   true,
		MaxLogSizeBytes: 5 * 1024 * 1024,
	},
}

func LoadConfig() Config {
	paths := []string{
		os.Getenv("RUIN_CONFIG"),
		filepath.Join(os.Getenv("HOME"), ".ruin/config"),
		"/etc/ruin/config",
	}
	final := builtInDefaults

	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				mergeConfigs(&final, &cfg)
			}
		}
	}

	return final
}

func mergeConfigs(base *Config, override *Config) {
	base.Kubectl.ProtectedContexts = append(base.Kubectl.ProtectedContexts, override.Kubectl.ProtectedContexts...)
	if base.Kubectl.AuthMethod == "" {
		base.Kubectl.AuthMethod = override.Kubectl.AuthMethod
	}
	if base.Kubectl.GracePeriodSeconds == 0 {
		base.Kubectl.GracePeriodSeconds = override.Kubectl.GracePeriodSeconds
	}
	if override.Kubectl.Awareness.Mode != "" {
		base.Kubectl.Awareness = override.Kubectl.Awareness
	}
	if override.Audit.ForwardURL != "" {
		base.Audit.ForwardURL = override.Audit.ForwardURL
	}
	if override.Audit.MaxLogSizeBytes > 0 {
		base.Audit.MaxLogSizeBytes = override.Audit.MaxLogSizeBytes
	}
	base.Audit.FallbackLocal = override.Audit.FallbackLocal
}

func (kc KubectlConfig) IsProtected(context string) bool {
	if len(kc.ProtectedContexts) == 0 {
		return true
	}
	for _, c := range kc.ProtectedContexts {
		if c == "*" || c == "ALL" || c == context {
			return true
		}
	}
	return false
}
