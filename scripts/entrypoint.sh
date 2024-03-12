#!/bin/sh

cat logo.txt
CMD=./godash

if [ -n "$PUID" ] || [ -n "$PGID" ]; then
    USER=appuser
    HOME=/app

    if ! grep -q "$USER" /etc/passwd; then
        # Usage: addgroup [-g GID] [-S] [USER] GROUP
        #
        # Add a group or add a user to a group
        #    -g GID       Group id
        addgroup -g "$PGID" "$USER"

        # Usage: adduser [OPTIONS] USER [GROUP]
        # Create new user, or add USER to GROUP
        #    -h DIR       Home directory
        #    -g GECOS     GECOS field
        #    -G GRP       Group
        #    -D           Don't assign a password
        #    -H           Don't create home directory
        #    -u UID       User id
        adduser -h "$HOME" -g "" -G "$USER" -D -H -u "$PUID" "$USER"
    fi

    chown "$USER":"$USER" "$HOME" -R
    printf "\nUID: %s GID: %s\n\n" "$PUID" "$PGID"
    exec su -c - $USER "$CMD"
else
    printf "\nWARNING: Running docker as root\n\n"
    exec "$CMD"
fi
