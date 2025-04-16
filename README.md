# ruin 🌧

> "Ruin everything. Except your important resources."

`ruin-kubectl` is a secure wrapper for `kubectl` that:
- Prevents accidental production mistakes
- Requires sudo or awareness prompts for flagged contexts
- Logs all sensitive command usage to a secure file or remote endpoint

---

## ✨ Features

- ⛔ Context-based protection using wildcards or exact names
- 🔑 Sudo-auth enforced with grace period (like `sudo`)
- ⏰ Awareness prompts (countdown or y/n) to catch "oops"
- 🔍 `init` flow to scaffold config per-user or system-wide
- 🗃️ Secure logging with append-only MAC-signed entries
- 📩 Optional log forwarding (e.g., to Vector, HTTP, or syslog)

---

## 🚀 Quick Install

Clone this project and run the installation script.

```bash
git clone https://github.com/shadetree-dev/ruin.git
chmod +x scripts/*
sudo ./scripts/install.sh
```

## 🗑️ Uninstall

```bash
# Prompted uninstall
sudo ./scripts/uninstall.sh

# Full/quiet uninstall
sudo ./scripts/uninstall.sh --full-clean
```

## 🔧 Usage

### Step 1: Initialize

```bash
# Launches interactive setup
ruin-kubectl init
```

### Step 2: Use it like `kubectl`

```bash
ruin-kubectl get pods
ruin-kubectl delete ns prod
```

Protected contexts will prompt or pause before dangerous actions, but by default allow read actions like `get`, `describe`, etc.

> [!NOTE]
> If you enable symlinking during ruin-kubectl init, you can simply use `kubectl` and can pass through any normal `kubectl` command through your protected `ruin-kubectl` wrapper!

## 🧠 Config

Default file paths:
- System-wide: `/etc/ruin/config`
- User: `~/.ruin/config`

Example config:

```yaml
kubectl:
  protected_contexts:
    - "*"
  auth_method: sudo
  grace_period_seconds: 300
  awareness_prompt:
    mode: pause # pause | prompt | none
    pause_seconds: 5
    only_on_write: true

audit:
  log_path: /var/log/ruin.log
  forward_url: ""
  fallback_local: true
  max_log_size_bytes: 5242880
```

## 🌀 Log Rotation

- On Linux: installs `logrotate.d` rule
- On macOS: installs `launchd` task and a rotation script

Append-only logs are enforced using `chattr +a` (Linux only).

All logs are:
- JSON-formatted and optionally signed with HMAC
- Designed for integration with Vector, syslog, or SIEMs

## 🔭 Roadmap
-	`ruin-aws` (IAM and CLI guardrails)
-	`ruin-git` (e.g. block force-push to main)
-	`.deb`, `.rpm`, and `Homebrew` packages
-	macOS `.dmg` installer
- Additional log push/audit capabilities

## ✏️ Contributing

PRs welcome!

## 📚 License

MIT © 2024–present [shadetree-dev](https://shadetree.dev/)

I'm here to ruin responsibly 💥🛡️