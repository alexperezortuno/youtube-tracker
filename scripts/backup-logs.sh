#!/bin/bash

set -euo pipefail

LOG_DIR="./logs"
BACKUP_DIR="./backups"

REMOTE_ENABLED=false
CLEAN_LOGS=false

while [[ $# -gt 0 ]]
do
    case "$1" in
        --rsync)
            REMOTE_ENABLED=true
            shift
            ;;
        --clean)
            CLEAN_LOGS=true
            shift
            ;;
        *)
            echo "Unknown flag: $1"
            exit 1
            ;;
    esac
done

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
HOSTNAME=$(hostname)

ARCHIVE="${BACKUP_DIR}/logs_${HOSTNAME}_${TIMESTAMP}.tar.gz"

REMOTE_USER="alex"
REMOTE_HOST="10.0.0.10"
REMOTE_PATH="/data/youtube-tracker/logs"

mkdir -p "$BACKUP_DIR"

tar -czf "$ARCHIVE" "$LOG_DIR"

echo "[OK] Backup generated: $ARCHIVE"

if [ "$REMOTE_ENABLED" = true ]; then

    rsync -avz \
        "$ARCHIVE" \
        "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}/"

    echo "[OK] Backup sent"

fi

if [ "$CLEAN_LOGS" = true ]; then

    find "$LOG_DIR" \
        -type f \
        -name "*.log" \
        -exec truncate -s 0 {} \;

    echo "[OK] Cleaned logs"

fi
