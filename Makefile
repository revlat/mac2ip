.PHONY: build build-all clean

# Force Bash shell (wichtig für Windows make)
SHELL := /bin/bash

# Detect OS
ifeq ($(OS),Windows_NT)
	BINARY_NAME=mac2ip.exe
else
	BINARY_NAME=mac2ip
endif

# Lokaler Build für aktuelles System
build:
	go build -o $(BINARY_NAME) .

# Alle Plattformen bauen (mit Plattform-Info im Namen für Releases)
build-all:
	GOOS=windows GOARCH=amd64 go build -o build/mac2ip-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -o build/mac2ip-windows-arm64.exe .
	GOOS=linux GOARCH=amd64 go build -o build/mac2ip-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o build/mac2ip-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o build/mac2ip-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o build/mac2ip-darwin-arm64 .
	@echo "✅ All builds in build/"

clean:
	rm -rf build/
	rm -f mac2ip.exe mac2ip
