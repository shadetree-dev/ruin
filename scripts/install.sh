#!/bin/bash
# scripts/install.sh

set -e

if [[ $EUID -ne 0 ]]; then
  echo "❌ This script must be run as root. Try: sudo ./scripts/install.sh"
  exit 1
fi

BIN_NAME="ruin-kubectl"
INSTALL_PATH="/usr/local/bin"
CONFIG_PATH="/etc/ruin"
CONFIG_FILE="$CONFIG_PATH/config"
LOG_PATH="/var/log/ruin.log"
LOGROTATE_SRC="./etc/logrotate.d/ruin"
LOGROTATE_DEST="/etc/logrotate.d/ruin"
LAUNCHD_PLIST_SRC="./etc/macos/ruin.logrotate.plist"
LAUNCHD_PLIST_DEST="/Library/LaunchDaemons/com.yourorg.ruinlogrotate.plist"
ROTATE_SCRIPT_SRC="./scripts/rotate-ruin-log.sh"
ROTATE_SCRIPT_DEST="/usr/local/bin/rotate-ruin-log.sh"

TARGET_USER="${SUDO_USER:-$USER}"
PRIMARY_GROUP=$(id -gn "$TARGET_USER")

# Ensure binary exists
if [ ! -f "$BIN_NAME" ]; then
  echo "❌ $BIN_NAME not found in current directory."
  exit 1
fi

echo "[*] Installing $BIN_NAME to $INSTALL_PATH..."
install -m 755 "$BIN_NAME" "$INSTALL_PATH/$BIN_NAME"

echo "[*] Setting up config in $CONFIG_PATH..."
mkdir -p "$CONFIG_PATH"
cp ./etc/ruin/config.example.yaml "$CONFIG_FILE"

echo "[*] Creating log file at $LOG_PATH..."
touch "$LOG_PATH"

if command -v groupadd >/dev/null 2>&1 && command -v usermod >/dev/null 2>&1; then
  groupadd -f ruinlog
  usermod -a -G ruinlog "$TARGET_USER"
  chown root:ruinlog "$LOG_PATH"
else
  chown root:"$PRIMARY_GROUP" "$LOG_PATH"
fi

chmod 664 "$LOG_PATH"

echo "[*] Enforcing append-only logging with chattr +a..."
if command -v chattr >/dev/null 2>&1; then
  chattr +a "$LOG_PATH" || echo "⚠️ Could not set immutable append-only attribute. You may need to run as root."
else
  echo "⚠️ chattr not available, skipping append-only enforcement."
fi

if [[ "$(uname -s)" == "Darwin" ]]; then
  echo "[*] Setting up launchd log rotation for macOS..."
  cp "$ROTATE_SCRIPT_SRC" "$ROTATE_SCRIPT_DEST"
  chmod +x "$ROTATE_SCRIPT_DEST"
  cp "$LAUNCHD_PLIST_SRC" "$LAUNCHD_PLIST_DEST"
  launchctl load "$LAUNCHD_PLIST_DEST"
else
  echo "[*] Installing logrotate config for Linux..."
  mkdir -p "$(dirname "$LOGROTATE_DEST")"
  cp "$LOGROTATE_SRC" "$LOGROTATE_DEST"
fi

read -rp "🌀 Do you want to alias 'kubectl' to use 'ruin-kubectl'? [y/N] " linkme
if [[ "$linkme" =~ ^[Yy]$ ]]; then
  ln -sf "$INSTALL_PATH/$BIN_NAME" "$INSTALL_PATH/kubectl"
  echo "[*] Symlink created: kubectl -> $BIN_NAME"
fi

if ! command -v ruin-kubectl >/dev/null 2>&1; then
  echo "⚠️ ruin-kubectl is not currently in your \$PATH."
  echo "👉 Add this to your shell config:"
  echo "  export PATH=\"/usr/local/bin:\$PATH\""
  echo ""
  echo "Or:"
  echo "  mkdir -p ~/.local/bin"
  echo "  mv /usr/local/bin/ruin-kubectl ~/.local/bin/"
  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
  echo ""
fi

echo "✅ ruin-kubectl installed successfully."
echo ""
echo "To enable kubectl-style autocompletion for ruin-kubectl (zsh):"
echo ""
echo "  autoload -Uz compinit"
echo "  compinit"
echo "  source <(kubectl completion zsh)"
echo "  compdef ruin-kubectl=kubectl"