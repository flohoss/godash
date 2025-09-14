#!/bin/sh
set -e

APP="./godash"
USER="appuser"
HOME="/app"

# Create user if PUID/PGID provided
if [ -n "$PUID" ] || [ -n "$PGID" ]; then
    PGID="${PGID:-1000}"
    PUID="${PUID:-1000}"

    # Ensure group exists
    if ! getent group "$PGID" >/dev/null 2>&1; then
        addgroup -g "$PGID" "$USER"
    fi

    # Ensure user exists
    if ! id -u "$PUID" >/dev/null 2>&1 2>/dev/null; then
        adduser -h "$HOME" -g "" -G "$USER" -D -H -u "$PUID" "$USER"
    fi

    chown -R "$PUID:$PGID" "$HOME"
    printf "\nUID: %s GID: %s\n\n" "$PUID" "$PGID"

    exec su-exec "$USER" "$APP" "$@"
else
    printf "\nWARNING: Running docker as root\n\n"
    exec "$APP" "$@"
fi
