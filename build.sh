#!/bin/bash

APP_NAME="UpToDate"
VERSION="1.0.0"

# Create build directory
mkdir -p dist

# Build for different platforms
echo "Building $APP_NAME version $VERSION..."
go mod tidy

# Linux AMD64
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "dist/${APP_NAME}${VERSION}-linux"

# Linux ARM64
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -o "dist/${APP_NAME}${VERSION}-linux-arm"

# Windows AMD64
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "dist/${APP_NAME}${VERSION}-windows.exe"

# Windows ARM64
echo "Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build -o "dist/${APP_NAME}${VERSION}-windows-arm64.exe"

# Mac Intel
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o "dist/${APP_NAME}${VERSION}-macos"

# Mac Apple Silicon
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o "dist/${APP_NAME}${VERSION}-macos-arm"

echo "Build complete!"