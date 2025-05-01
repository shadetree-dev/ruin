# Changelog

All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/) (assuming I do not screw that up...)

---

## [0.1.1] - 2025-04-30
### Added
- In `internal/awareness/prompt.go` added the `drain` action as a write operation.

---

## [0.1.0] - 2025-04-16
### Added
- Initial pre-release of `ruin-kubectl`!
- Created `Go` binary that enables prompting for protected Kubernetes contexts, logging, checking for context updates, and more.
- `ruin-kubectl init` flow for first-time setup.
- `sudo` authentication with grace period for protected contexts.
- Awareness prompts: pause, prompt, none. These give flexibility to the user to decide what guardrails to place on themselves.
- Append-only audit logging (`/var/log/ruin.log`).
    - macOS `launchd` and Linux `logrotate` integration.
- Install + uninstall scripts.
- Optional symlink to `kubectl` on installation (`kubectl -> /usr/local/bin/ruin-kubectl`).
- Local + system config fallback support to sane defaults, with override capabilities in the `~/.ruin/config` file.