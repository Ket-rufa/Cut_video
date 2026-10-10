#!/bin/bash
if command -v ffmpeg &>/dev/null; then
  echo "FFmpeg already installed: $(ffmpeg -version 2>&1 | head -1)"
  exit 0
fi
if [ -f /etc/alpine-release ]; then
  apk add --no-cache ffmpeg
elif [ -f /etc/debian_version ]; then
  apt-get update && apt-get install -y ffmpeg
elif [ -f /etc/fedora-release ]; then
  dnf install -y ffmpeg
elif [[ "$OSTYPE" == "darwin"* ]]; then
  brew install ffmpeg
else
  echo "Please install FFmpeg manually from https://ffmpeg.org/download.html"
  exit 1
fi
echo "FFmpeg installed: $(ffmpeg -version 2>&1 | head -1)"
