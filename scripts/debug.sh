#!/bin/sh

go build -gcflags="all=-N -l" -o /tmp/app
dlv --listen=:4001 --headless=true --api-version=2 --accept-multiclient exec /tmp/app
