#!/bin/sh
set -e

APP="./godash"
USER="appuser"
HOME="/app"

# Create user if PUID/PGID provided
if [ -n "$PUID" ] || [ -n "$PGID" ]; then
    PGID="${PGID:-1000}"
    PUID="${PUID:-1000}"

    # Check if group with PGID exists
    if ! getent group "$PGID" >/dev/null 2>&1; then
        addgroup -g "$PGID" "$USER"
    else
        GROUP_NAME="$(getent group "$PGID" | cut -d: -f1)"
    fi

    # Check if user with PUID exists
    if ! getent passwd "$PUID" >/dev/null 2>&1; then
        adduser -h "$HOME" -g "" -G "${GROUP_NAME:-$USER}" -D -H -u "$PUID" "$USER"
    else
        USER="$(getent passwd "$PUID" | cut -d: -f1)"
    fi

    chown -R "$PUID:$PGID" "$HOME"
    printf "\nUID: %s GID: %s (user: %s)\n\n" "$PUID" "$PGID" "$USER"

    exec su-exec "$USER" "$APP" "$@"
else
    printf "\nWARNING: Running docker as root\n\n"
    exec "$APP" "$@"
fi
