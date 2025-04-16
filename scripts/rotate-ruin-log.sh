#!/bin/bash
# Rotate ruin.log manually on macOS

LOG_FILE="/var/log/ruin.log"
MAX_SIZE_BYTES=$((5 * 1024 * 1024)) # 5MB

if [ -f "$LOG_FILE" ]; then
  FILE_SIZE=$(stat -f%z "$LOG_FILE")
  if [ "$FILE_SIZE" -gt "$MAX_SIZE_BYTES" ]; then
    mv "$LOG_FILE" "$LOG_FILE.$(date +%Y%m%d%H%M%S)"
    touch "$LOG_FILE"
    chown root:admin "$LOG_FILE"
    chmod 664 "$LOG_FILE"
  fi
fi