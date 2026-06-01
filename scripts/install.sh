#!/usr/bin/env bash

set -e

REPO="alexperezortuno/youtube-tracker"
BINARY="yt-tracker"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# normalizar arch
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
fi

# detectar windows (git bash / wsl)
if [[ "$OS" == *"mingw"* ]] || [[ "$OS" == *"msys"* ]]; then
  OS="windows"
  EXT=".exe"
fi

FILE="$BINARY-$OS-$ARCH$EXT"

echo "Installing $FILE..."

URL="https://github.com/$REPO/releases/latest/download/$FILE"

curl -L "$URL" -o "$BINARY$EXT"

chmod +x "$BINARY$EXT"

# mover a PATH
if [ "$OS" = "windows" ]; then
  echo "Move $BINARY.exe to a folder in your PATH manually"
else
  sudo mv "$BINARY" /usr/local/bin/$BINARY
  echo "Installed to /usr/local/bin/$BINARY"
fi

echo "Done!"
