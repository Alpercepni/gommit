#!/usr/bin/env sh
# Usage (1-liner):
#   curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
set -e

OWNER=${OWNER:-Hangell}
REPO=${REPO:-gommit}
API="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
UA="gommit-installer"

uname_s=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$uname_s" in
  linux)  OS=linux ;;
  darwin) OS=darwin ;;
  *) echo "Unsupported OS: $uname_s"; exit 1 ;;
esac

uname_m=$(uname -m)
case "$uname_m" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported arch: $uname_m"; exit 1 ;;
esac

tmp="$(mktemp -d 2>/dev/null || mktemp -d -t gommit)"
url=$(
  curl -fsSL -H "User-Agent: $UA" "$API" \
  | awk -v os="$OS" -v arch="$ARCH" -F '"' '/browser_download_url/ {print $4}' \
  | grep "gommit_.*_${OS}_${ARCH}\.tar\.gz$" | head -n1
)

if [ -z "$url" ]; then
  echo "No matching asset for ${OS}_${ARCH}"; exit 1
fi

echo "Downloading $url"
cd "$tmp"
curl -fsSL -o gommit.tar.gz "$url"
tar -xzf gommit.tar.gz

bin="$(find . -type f -name gommit -perm -111 2>/dev/null | head -n1)"
[ -z "$bin" ] && bin="$(find . -type f -name gommit | head -n1)"
chmod +x "$bin"

"$bin" --install

rm -rf "$tmp"
echo ""
echo "Done. Restart your shell and run: gommit --version"
