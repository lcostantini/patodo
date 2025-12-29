#!/bin/bash

# Install script for patodo

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$HOME/.local/bin"

echo "Building patodo..."
cd "$SCRIPT_DIR"
go build -o patodo .

echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp patodo "$INSTALL_DIR/"

echo "✓ patodo installed successfully!"
echo "Make sure $INSTALL_DIR is in your PATH"
echo "Run 'patodo' to start"
