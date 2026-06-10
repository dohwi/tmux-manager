#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/dohwi/tmux-manager.git"
INSTALL_DIR="$HOME/.local/share/tmux-manager"
GO_VERSION="1.23.0"

BOLD="$(tput bold 2>/dev/null || printf '')"
GREEN="$(tput setaf 2 2>/dev/null || printf '')"
CYAN="$(tput setaf 6 2>/dev/null || printf '')"
RESET="$(tput sgr0 2>/dev/null || printf '')"

info()  { printf "  ${CYAN}%s${RESET}\n" "$*"; }
done_() { printf " ${GREEN}✓${RESET} %s\n" "$*"; }
abort() { printf "✗ %s\n" "$*" >&2; exit 1; }

# --- go detection & install ---
have_go() { command -v go >/dev/null 2>&1; }

install_go() {
  info "Go not found, installing Go $GO_VERSION..."

  local os arch
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$(uname -m)" in
    x86_64)  arch="amd64" ;;
    aarch64) arch="arm64" ;;
    armv7l)  arch="armv6l" ;;
    *)       abort "unsupported architecture: $(uname -m)" ;;
  esac

  local tarball="go${GO_VERSION}.${os}-${arch}.tar.gz"
  local url="https://go.dev/dl/${tarball}"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT

  curl -fsSL "$url" -o "$tmpdir/$tarball" || abort "failed to download Go"
  sudo tar -C /usr/local -xzf "$tmpdir/$tarball" || abort "failed to extract Go"
  export PATH="/usr/local/go/bin:$PATH"
  done_ "Go $GO_VERSION installed"
}

# --- main ---
echo "${BOLD}tmux-manager installer${RESET}"
echo

have_go || install_go

if [[ -d "$INSTALL_DIR" ]]; then
  info "Updating existing install at $INSTALL_DIR"
  git -C "$INSTALL_DIR" pull --ff-only origin main 2>/dev/null || true
else
  info "Cloning into $INSTALL_DIR"
  git clone "$REPO" "$INSTALL_DIR"
fi

info "Building..."
(cd "$INSTALL_DIR" && go build -o tmux-manager .) || abort "build failed"

info "Running setup..."
"$INSTALL_DIR/tmux-manager" setup

echo
done_ "Install complete. Run ${BOLD}tm${RESET} to start."
