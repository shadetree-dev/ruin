#!/bin/bash
# scripts/install.sh

set -e

BIN_NAME="ruin-kubectl"
INSTALL_PATH="/usr/local/bin"
CONFIG_PATH="/etc/ruin"
CONFIG_FILE="$CONFIG_PATH/config"
LOG_PATH="/var/log/ruin.log"
LOGROTATE_SRC="./etc/logrotate.d/ruin"
LOGROTATE_DEST="/etc/logrotate.d/ruin"

# Ensure binary exists
if [ ! -f "$BIN_NAME" ]; then
  echo "❌ $BIN_NAME not found in current directory."
  exit 1
fi

# Install binary
echo "[*] Installing $BIN_NAME to $INSTALL_PATH..."
install -m 755 "$BIN_NAME" "$INSTALL_PATH/$BIN_NAME"

# Copy example config
echo "[*] Setting up config in $CONFIG_PATH..."
mkdir -p "$CONFIG_PATH"
cp ./etc/ruin/config.example.yaml "$CONFIG_FILE"

# Set up log file with safe permissions
echo "[*] Creating log file at $LOG_PATH..."
touch "$LOG_PATH"
if command -v groupadd >/dev/null 2>&1 && command -v usermod >/dev/null 2>&1; then
  groupadd -f ruinlog
  usermod -a -G ruinlog "$SUDO_USER"
  chown root:ruinlog "$LOG_PATH"
else
    PRIMARY_GROUP=$(id -gn "$SUDO_USER")
    chown root:"$PRIMARY_GROUP" "$LOG_PATH"
fi
chmod 664 "$LOG_PATH"

# Enforce append-only (immutable) flag
echo "[*] Enforcing append-only logging with chattr +a..."
if command -v chattr >/dev/null 2>&1; then
  chattr +a "$LOG_PATH" || echo "⚠️ Could not set immutable append-only attribute. You may need to run as root."
else
  echo "⚠️ chattr not available, skipping append-only enforcement."
fi

# Set up logrotate
if [[ "$(uname -s)" == "Darwin" ]]; then
  echo "[*] Setting up launchd log rotation for macOS..."
  cp ./scripts/rotate-ruin-log.sh /usr/local/bin/rotate-ruin-log.sh
  chmod +x /usr/local/bin/rotate-ruin-log.sh
  cp ./etc/macos/ruin.logrotate.plist /Library/LaunchDaemons/com.yourorg.ruinlogrotate.plist
  launchctl load /Library/LaunchDaemons/com.yourorg.ruinlogrotate.plist
else
  echo "[*] Installing logrotate config for Linux..."
  mkdir -p "$(dirname "$LOGROTATE_DEST")"
  cp "$LOGROTATE_SRC" "$LOGROTATE_DEST"
fi

# Prompt for symlink
read -rp "🌀 Do you want to alias 'kubectl' to use 'ruin-kubectl'? [y/N] " linkme
if [[ "$linkme" =~ ^[Yy]$ ]]; then
  ln -sf "$INSTALL_PATH/$BIN_NAME" "$INSTALL_PATH/kubectl"
  echo "[*] Symlink created: kubectl -> $BIN_NAME"
fi

echo "✅ ruin-kubectl installed successfully."