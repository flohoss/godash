#!/bin/sh
set -e

templ generate --watch --proxybind="0.0.0.0" --proxyport=7331 --proxy="http://localhost:8156" --cmd="go run ." --open-browser=false
