#!/usr/bin/env bash
#
# Install asbuilt from a GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/Balanov/asbuilt/main/install.sh | sh
#
# Environment:
#   ASBUILT_VERSION       version to install (default: the latest release)
#   ASBUILT_INSTALL_DIR   target directory (default: $HOME/.local/bin)
#
# The script installs into the user directory, needs no root, and can be run
# again at any time: it replaces the binary in place.

set -euo pipefail

REPO="Balanov/asbuilt"
API="https://api.github.com/repos/${REPO}"
DOWNLOAD="https://github.com/${REPO}/releases/download"
INSTALL_DIR="${ASBUILT_INSTALL_DIR:-${HOME}/.local/bin}"
TAG=""

log()  { printf 'asbuilt-install: %s\n' "$*"; }
warn() { printf 'asbuilt-install: warning: %s\n' "$*" >&2; }
die()  { printf 'asbuilt-install: error: %s\n' "$*" >&2; exit 1; }

have() { command -v "$1" >/dev/null 2>&1; }

need_curl() {
    have curl || die "curl is required to download asbuilt"
}

detect_os() {
    case "$(uname -s)" in
        Linux)  printf 'linux\n' ;;
        Darwin) printf 'darwin\n' ;;
        *)      die "unsupported operating system: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  printf 'amd64\n' ;;
        aarch64|arm64) printf 'arm64\n' ;;
        *)             die "unsupported architecture: $(uname -m)" ;;
    esac
}

# latest_tag prints the tag of the latest published release.
latest_tag() {
    need_curl
    local json tag
    json="$(curl -fsSL -H 'Accept: application/vnd.github+json' "${API}/releases/latest")" \
        || die "could not read the latest release from ${API}"
    # The whole document is read before the first match is taken, so the
    # pipeline never ends early on a closed pipe.
    tag="$(printf '%s' "$json" | awk -F'"' '/"tag_name"/{print $4}')"
    tag="${tag%%$'\n'*}"
    [ -n "$tag" ] || die "the latest release carries no tag_name"
    printf '%s\n' "$tag"
}

# sha256_of prints the SHA-256 of a file, or fails when no tool is available.
sha256_of() {
    if have sha256sum; then
        sha256sum "$1" | awk '{print $1}'
    elif have shasum; then
        shasum -a 256 "$1" | awk '{print $1}'
    elif have openssl; then
        openssl dgst -sha256 "$1" | awk '{print $NF}'
    else
        return 1
    fi
}

# verify_checksum compares the downloaded file with the checksum published in
# the release. A mismatch is a warning, never a fatal error.
verify_checksum() {
    local asset="$1" file="$2" sums expected actual

    if ! have curl; then
        warn "curl is not available, skipping checksum verification"
        return 0
    fi
    sums="${TMP_DIR}/checksums.txt"
    if ! curl -fsSL -o "$sums" "${DOWNLOAD}/${TAG}/checksums.txt"; then
        warn "could not download checksums.txt, skipping checksum verification"
        return 0
    fi

    expected="$(awk -v want="$asset" '{name=$2; sub(/^\*/, "", name); if (name == want) {print $1; exit}}' "$sums")"
    if [ -z "$expected" ]; then
        warn "no checksum is published for ${asset}, skipping checksum verification"
        return 0
    fi
    if ! actual="$(sha256_of "$file")"; then
        warn "no sha256 tool found, skipping checksum verification"
        return 0
    fi
    if [ "$actual" != "$expected" ]; then
        warn "checksum mismatch for ${asset}"
        warn "  expected: ${expected}"
        warn "  actual:   ${actual}"
        return 0
    fi
    log "checksum verified"
}

main() {
    need_curl

    local version os arch target asset url dest

    if [ -n "${ASBUILT_VERSION:-}" ]; then
        version="${ASBUILT_VERSION#v}"
        TAG="v${version}"
    else
        TAG="$(latest_tag)"
        version="${TAG#v}"
    fi

    os="$(detect_os)"
    arch="$(detect_arch)"
    target="${os}-${arch}"
    asset="asbuilt-${version}-${target}"
    url="${DOWNLOAD}/${TAG}/${asset}"

    TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/asbuilt-install.XXXXXX")"
    trap 'rm -rf "$TMP_DIR"' EXIT

    log "installing asbuilt ${version} for ${target}"
    log "downloading ${url}"
    curl -fL --retry 3 --retry-delay 2 -o "${TMP_DIR}/${asset}" "$url" \
        || die "download failed: ${url}"

    chmod +x "${TMP_DIR}/${asset}"
    verify_checksum "$asset" "${TMP_DIR}/${asset}"

    mkdir -p "$INSTALL_DIR"
    dest="${INSTALL_DIR}/asbuilt"
    if have install; then
        install -m 0755 "${TMP_DIR}/${asset}" "$dest" || die "could not install into ${INSTALL_DIR}"
    else
        cp "${TMP_DIR}/${asset}" "$dest" || die "could not install into ${INSTALL_DIR}"
        chmod 0755 "$dest"
    fi

    log "installed asbuilt ${version} to ${dest}"

    case ":${PATH}:" in
        *":${INSTALL_DIR}:"*) ;;
        *) warn "${INSTALL_DIR} is not in your PATH; add it with: export PATH=\"${INSTALL_DIR}:\$PATH\"" ;;
    esac

    printf '\nCheck the installation:\n  %s --version\n' "$dest"
    printf 'Audit this machine:\n  %s check\n' "$dest"
    printf 'List the rule catalog:\n  %s rules\n\n' "$dest"
}

main "$@"
