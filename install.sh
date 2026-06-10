#!/usr/bin/env bash
set -euo pipefail

REPO="dohwi/tmux-manager"
BIN_DIR="$HOME/go/bin"
BIN="$BIN_DIR/tmux-manager"
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

TARBALL=""

cleanup() {
  if [ -n "$TARBALL" ] && [ -f "/tmp/$TARBALL" ]; then
    rm -f "/tmp/$TARBALL"
  fi
}
trap cleanup EXIT

info()  { echo -e "${GREEN}→${NC} $1"; }
err()   { echo -e "${RED}→${NC} $1"; exit 1; }

install_go() {
  if command -v go &>/dev/null; then return 0; fi
  info "Go not found, installing..."

  local os arch
  case "$(uname -s)" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *)      err "Unsupported OS: $(uname -s)" ;;
  esac
  case "$(uname -m)" in
    x86_64)  arch="amd64" ;;
    aarch64) arch="arm64" ;;
    arm64)   arch="arm64" ;;
    *)       err "Unsupported arch: $(uname -m)" ;;
  esac

  local latest_go
  latest_go=$(curl -sSfL 'https://go.dev/VERSION?m=text' 2>/dev/null | head -1)
  [ -z "$latest_go" ] && err "Cannot detect latest Go version"

  TARBALL="${latest_go}.${os}-${arch}.tar.gz"
  local url="https://go.dev/dl/${TARBALL}"
  local go_local="$HOME/.local/go"

  info "Downloading Go ${latest_go}..."
  curl -sSfL "$url" -o "/tmp/${TARBALL}"

  info "Extracting to $go_local..."
  rm -rf "$go_local"
  mkdir -p "$go_local"
  tar -C "$go_local" --strip-components=1 -xzf "/tmp/${TARBALL}"

  export GOROOT="$go_local"
  export PATH="$go_local/bin:$PATH"
  info "Go ${latest_go} installed"
}

echo ""
info "tmux-manager installer"
install_go

info "Resolving latest release..."
LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
[ -z "$LATEST_TAG" ] && err "Cannot detect latest release tag"

info "go install github.com/${REPO}/cmd/tmux-manager@${LATEST_TAG}"
GOBIN="$BIN_DIR" go install "github.com/${REPO}/cmd/tmux-manager@${LATEST_TAG}"

info "Running setup..."
"$BIN" setup

echo ""
info "✓ Installed to $BIN"
echo ""
echo "  Run this to use immediately:"
echo "    export PATH=\"\$PATH:\$HOME/go/bin\""
echo "  Or restart your shell."
echo "  Then: tm"
echo "  Then: tm update  (to upgrade later)"
echo "  To uninstall: tm setup --uninstall"
echo ""
