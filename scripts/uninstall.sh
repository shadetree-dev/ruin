#!/bin/bash
# scripts/uninstall.sh

set -e

BIN_NAME="ruin-kubectl"
INSTALL_PATH="/usr/local/bin"
CONFIG_PATH="/etc/ruin"
LOG_PATH="/var/log/ruin.log"
LOGROTATE_PATH="/etc/logrotate.d/ruin"
LAUNCHD_PLIST="/Library/LaunchDaemons/com.yourorg.ruinlogrotate.plist"
ROTATE_SCRIPT="/usr/local/bin/rotate-ruin-log.sh"

FULL_CLEAN=false
if [[ "$1" == "--full-clean" ]]; then
  FULL_CLEAN=true
fi

echo "[*] Uninstalling $BIN_NAME..."

# Remove binary
rm -f "$INSTALL_PATH/$BIN_NAME"

# Remove symlink if it exists and points to us
if [ -L "$INSTALL_PATH/kubectl" ] && [ "$(readlink "$INSTALL_PATH/kubectl")" == "$BIN_NAME" ]; then
  rm -f "$INSTALL_PATH/kubectl"
  echo "[*] Removed symlink: kubectl -> $BIN_NAME"
fi

# Remove config
rm -rf "$CONFIG_PATH"
echo "[*] Removed config directory: $CONFIG_PATH"

# Remove logrotate config
if [[ "$(uname -s)" == "Linux" ]]; then
  rm -f "$LOGROTATE_PATH"
  echo "[*] Removed logrotate config: $LOGROTATE_PATH"
fi

# Remove append-only flag (only Linux)
if [[ "$(uname -s)" == "Linux" && -f "$LOG_PATH" ]]; then
  echo "[*] Unlocking append-only flag on $LOG_PATH..."
  chattr -a "$LOG_PATH" || echo "⚠️ Could not remove append-only attribute."
fi

if [ "$FULL_CLEAN" = true ]; then
  echo "[*] Full clean mode enabled. Deleting log file..."
  rm -f "$LOG_PATH"
  echo "[*] Deleted log file: $LOG_PATH"
else
  # Prompt to delete log file
  read -rp "🧹 Do you want to delete the log file at $LOG_PATH? [y/N] " nuke_logs
  if [[ "$nuke_logs" =~ ^[Yy]$ ]]; then
    rm -f "$LOG_PATH"
    echo "[*] Deleted log file: $LOG_PATH"
  else
    echo "[!] Log file retained: $LOG_PATH"
  fi
fi

# macOS-specific: remove launchd job and script
if [[ "$(uname -s)" == "Darwin" ]]; then
  echo "[*] Removing launchd logrotate job..."
  launchctl unload "$LAUNCHD_PLIST" 2>/dev/null || true
  rm -f "$LAUNCHD_PLIST"
  rm -f "$ROTATE_SCRIPT"
  echo "[*] Removed macOS launchd config and script."
fi

echo "✅ $BIN_NAME has been uninstalled."