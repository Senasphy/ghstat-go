#!/usr/bin/env sh
set -eu

REPO="${GHSTAT_REPO:-Senasphy/ghstat-go}"
BIN_NAME="${GHSTAT_BIN:-ghstat-go}"
VERSION="${1:-${GHSTAT_VERSION:-}}"
INSTALL_DIR="${GHSTAT_INSTALL_DIR:-/usr/local/bin}"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd tar
need_cmd uname

detect_os() {
  os="$(uname -s)"
  case "$os" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)
      echo "unsupported OS: $os" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      echo "unsupported architecture: $arch" >&2
      exit 1
      ;;
  esac
}

resolve_latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n 1
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ -z "$VERSION" ]; then
  VERSION="$(resolve_latest_version)"
fi

if [ -z "$VERSION" ]; then
  echo "failed to resolve release version" >&2
  exit 1
fi

ARTIFACT="${BIN_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARTIFACT}"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

ARCHIVE_PATH="${TMP_DIR}/${ARTIFACT}"
curl -fL "$URL" -o "$ARCHIVE_PATH"
tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

SOURCE_BIN="${TMP_DIR}/${BIN_NAME}"
if [ ! -f "$SOURCE_BIN" ]; then
  echo "binary not found in archive: $SOURCE_BIN" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR" 2>/dev/null || true
TARGET_BIN="${INSTALL_DIR%/}/${BIN_NAME}"

if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "$SOURCE_BIN" "$TARGET_BIN"
else
  if command -v sudo >/dev/null 2>&1; then
    sudo install -m 0755 "$SOURCE_BIN" "$TARGET_BIN"
  else
    echo "cannot write to ${INSTALL_DIR}. run with elevated privileges or set GHSTAT_INSTALL_DIR" >&2
    exit 1
  fi
fi

echo "installed ${BIN_NAME} ${VERSION} to ${TARGET_BIN}"
if ! printf '%s' "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
  echo "note: ${INSTALL_DIR} is not in PATH"
fi
