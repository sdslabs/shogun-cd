#!/usr/bin/env bash
set -euo pipefail

required_tools=(curl openssl awk tr)
missing_tools=()

for tool in "${required_tools[@]}"; do
  if ! command -v "$tool" > /dev/null 2>&1; then
    missing_tools+=("$tool")
  fi
done

if [ ${#missing_tools[@]} -eq 0 ]; then
  echo "All required tools are present"
  exit 0
fi

echo "Missing required tools: ${missing_tools[*]}, attempting installation"

if command -v apt-get > /dev/null 2>&1; then
  if [ "$EUID" -eq 0 ]; then
    apt-get update -qq
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "${missing_tools[@]}"
  elif command -v sudo > /dev/null 2>&1; then
    sudo apt-get update -qq
    sudo env DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "${missing_tools[@]}"
  else
    echo "::error::Missing required tools (${missing_tools[*]}) and no root or sudo to install them"
    exit 1
  fi
elif command -v apk > /dev/null 2>&1; then
  apk add --no-cache "${missing_tools[@]}"
else
  echo "::error::Missing required tools (${missing_tools[*]}) and no supported package manager (apt-get, apk) found"
  exit 1
fi

for tool in "${missing_tools[@]}"; do
  if ! command -v "$tool" > /dev/null 2>&1; then
    echo "::error::$tool is still missing after installation"
    exit 1
  fi
done

echo "Installed: ${missing_tools[*]}"
