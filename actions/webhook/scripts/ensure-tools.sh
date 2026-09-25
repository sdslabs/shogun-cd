#!/usr/bin/env bash
set -euo pipefail

required_tools=(curl openssl awk tr mktemp rm)
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

packages=()
for tool in "${missing_tools[@]}"; do
  case "$tool" in
    awk) package=gawk ;;
    tr|mktemp|rm) package=coreutils ;;
    *) package="$tool" ;;
  esac
  # Several commands may come from the same package.
  case " ${packages[*]-} " in
    *" $package "*) ;;
    *) packages+=("$package") ;;
  esac
done

installer=()
if [ "$EUID" -ne 0 ]; then
  if command -v sudo > /dev/null 2>&1; then
    installer=(sudo -n)
  else
    echo "::error::Missing tools (${missing_tools[*]}). Install packages ${packages[*]} on the runner, or allow root/passwordless sudo installation."
    exit 1
  fi
fi

if command -v apt-get > /dev/null 2>&1; then
  "${installer[@]}" apt-get update -qq
  "${installer[@]}" env DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "${packages[@]}"
elif command -v apk > /dev/null 2>&1; then
  "${installer[@]}" apk add --no-cache "${packages[@]}"
else
  echo "::error::Missing tools (${missing_tools[*]}). Automatic installation supports apt-get and apk; install these tools manually on this runner."
  exit 1
fi

for tool in "${missing_tools[@]}"; do
  if ! command -v "$tool" > /dev/null 2>&1; then
    echo "::error::$tool is still missing after installation"
    exit 1
  fi
done

echo "Installed: ${missing_tools[*]}"
