#!/bin/sh

set -e

CONFIG_FILE="/app/storage/bookmarks.yaml"
APP="./godash"
USER="appuser"
HOME="/app"

log_message() {
    local message="$1"
    local timestamp
    timestamp=$(date +"%Y/%m/%d %H:%M:%S")
    echo "${timestamp} INFO ${message}"
}

cat logo.txt

# Create user if PUID/PGID are provided
if [ -n "$PUID" ] || [ -n "$PGID" ]; then
    if ! grep -q "$USER" /etc/passwd; then
        addgroup -g "${PGID:-1000}" "$USER"
        adduser -h "$HOME" -g "" -G "$USER" -D -H -u "${PUID:-1000}" "$USER"
    fi
    chown -R "$USER:$USER" "$HOME"
    printf "\nUID: %s GID: %s\n\n" "${PUID:-1000}" "${PGID:-1000}"

    RUN_AS="su-exec $USER"
else
    printf "\nWARNING: Running docker as root\n\n"
    RUN_AS=""
fi

while true; do
    log_message "Starting GoDash..."
    $RUN_AS $APP &
    PID=$!
    log_message "GoDash started with PID: $PID"

    log_message "Waiting for $CONFIG_FILE to exist..."
    while [ ! -f "$CONFIG_FILE" ]; do
        sleep 1
    done
    log_message "$CONFIG_FILE now exists. Proceeding..."

    log_message "Watching for changes in $CONFIG_FILE..."
    inotifywait -qq -e modify "$CONFIG_FILE"
    log_message "Detected change in $CONFIG_FILE."

    log_message "Config changed. Reloading..."
    log_message "Killing GoDash (PID: $PID)..."
    kill $PID
    log_message "Waiting for GoDash (PID: $PID) to terminate..."
    wait $PID
    log_message "GoDash (PID: $PID) terminated. Restarting loop..."
done

log_message "Entrypoint script exited. This should not happen in the loop."
