#!/bin/sh
set -e

templ generate
./scripts/coverage.sh coverage.out "$@"
go tool cover -func=coverage.out | tee coverage.txt
