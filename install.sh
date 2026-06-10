#!/usr/bin/env bash
set -e

REPO="dohwi/tmux-manager"
BIN_DIR="$HOME/go/bin"
BIN="$BIN_DIR/tmux-manager"
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

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

  local tarball="${latest_go}.${os}-${arch}.tar.gz"
  local url="https://go.dev/dl/${tarball}"
  local go_local="$HOME/.local/go"

  info "Downloading Go ${latest_go}..."
  curl -sSfL "$url" -o "/tmp/${tarball}"

  info "Extracting to $go_local..."
  rm -rf "$go_local"
  mkdir -p "$go_local"
  tar -C "$go_local" --strip-components=1 -xzf "/tmp/${tarball}"
  rm -f "/tmp/${tarball}"

  export GOROOT="$go_local"
  export PATH="$go_local/bin:$PATH"
  info "Go ${latest_go} installed"
}

echo ""
info "tmux-manager installer"
install_go

info "go install github.com/${REPO}/cmd/tmux-manager@main"
GOBIN="$BIN_DIR" GOPROXY=direct go install "github.com/${REPO}/cmd/tmux-manager@main"

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
echo ""
