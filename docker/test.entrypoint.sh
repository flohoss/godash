#!/bin/sh
set -e

templ generate
go test ./... -covermode=atomic -coverprofile=coverage.out "$@"
go tool cover -func=coverage.out | tee coverage.txt
