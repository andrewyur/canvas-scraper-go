#/bin/bash

set -e

GOOS=linux GOARCH=amd64 go build -o canvas-scraper-linux
GOOS=windows GOARCH=amd64 go build -o canvas-scraper-windows.exe
GOOS=darwin GOARCH=arm64 go build -o canvas-scraper-mac-arm64
