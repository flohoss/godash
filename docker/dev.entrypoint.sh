#!/bin/sh
set -e

mkdir -p /app/assets/js

cp /app/node_modules/htmx.org/dist/htmx.min.js /app/assets/js/htmx.min.js
cp /app/node_modules/htmx-ext-sse/dist/sse.min.js /app/assets/js/htmx-sse.min.js

cat /logo.txt

templ generate --watch --proxybind="0.0.0.0" --proxy="http://localhost:8156" --cmd="go run ." --open-browser=false
