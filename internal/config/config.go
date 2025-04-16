package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type KubectlConfig struct {
	ProtectedContexts  []string `yaml:"protected_contexts"`
	AuthMethod         string   `yaml:"auth_method"`
	GracePeriodSeconds int      `yaml:"grace_period_seconds"`
}

type Config struct {
	Kubectl KubectlConfig `yaml:"kubectl"`
}

func LoadConfig() Config {
	paths := []string{"/etc/ruin/config", filepath.Join(os.Getenv("HOME"), ".ruin/config")}
	final := Config{}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			fmt.Println("Reading config:", path) // ← Add this line
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
}

func (kc KubectlConfig) IsProtected(context string) bool {
	if len(kc.ProtectedContexts) == 0 {
		return true // No list = protect everything
	}

	for _, c := range kc.ProtectedContexts {
		if c == "*" || c == "ALL" || c == context {
			return true
		}
	}

	return false
}
