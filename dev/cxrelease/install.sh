#!/bin/sh

# Chronix Server Installation Script
# https://chronixhq.com/install.sh
# 
# This script detects your platform, downloads the latest Chronix binary,
# installs it to your system, and helps you get started.

set -e

# Colors for output
if [ -t 1 ]; then
    # Use actual escape characters for compatibility with different printf implementations
    # This ensures colors work even when passed as arguments to printf %s
    ESC=$(printf '\033')
    RED="${ESC}[0;31m"
    GREEN="${ESC}[0;32m"
    YELLOW="${ESC}[1;33m"
    BLUE="${ESC}[0;34m"
    NC="${ESC}[0m"
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

info() { printf "%sinfo:%s %s\n" "${BLUE}" "${NC}" "$1"; }
success() { printf "%ssuccess:%s %s\n" "${GREEN}" "${NC}" "$1"; }
warn() { printf "%swarn:%s %s\n" "${YELLOW}" "${NC}" "$1"; }
error() { printf "%serror:%s %s\n" "${RED}" "${NC}" "$1"; }

printf "${GREEN}
   ______ __                            _        
  / ____// /_   _____ ____   ____   (_) _  __
 / /    / __ \\ / ___// __ \\ / __ \\ / / | |/_/
/ /___ / / / // /   / /_/ // / / // /  >  <  
\\____//_/ /_//_/    \\____//_/ /_//_/  /_/|_| 
                                             
${NC}"
printf "Chronix Server Installation\n\n"

# Check for curl
if ! command -v curl >/dev/null 2>&1; then
    error "curl is required but not installed. Please install curl and try again."
    exit 1
fi

# Detect OS
OS_TYPE="$(uname -s)"
case "$OS_TYPE" in
    Linux*)     OS='linux';;
    Darwin*)    OS='darwin';;
    CYGWIN*|MINGW*|MSYS*) OS='windows';;
    *)          OS='unknown';;
esac

# Detect Architecture
ARCH_TYPE="$(uname -m)"
case "$ARCH_TYPE" in
    x86_64|amd64) ARCH='amd64';;
    arm64|aarch64) ARCH='arm64';;
    *)             ARCH='unknown';;
esac

PLATFORM="${OS}-${ARCH}"

# Verify platform
info "Detected platform: ${YELLOW}${PLATFORM}${NC}"
printf "Is this correct? [Y/n] "
read confirm_platform < /dev/tty
confirm_platform=${confirm_platform:-Y}

case "$confirm_platform" in
    [Yy]*) ;;
    *)
        printf "Available platforms:\n"
        printf "  1) linux-amd64\n"
        printf "  2) linux-arm64\n"
        printf "  3) windows-amd64\n"
        printf "  4) darwin-arm64\n"
        printf "  5) darwin-amd64\n"
        printf "Select your platform [1-5]: "
        read platform_choice < /dev/tty
        case "$platform_choice" in
            1) PLATFORM="linux-amd64";;
            2) PLATFORM="linux-arm64";;
            3) PLATFORM="windows-amd64";;
            4) PLATFORM="darwin-arm64";;
            5) PLATFORM="darwin-amd64";;
            *) error "Invalid choice. Exiting."; exit 1;;
        esac
        ;;
esac

# Determine default installation directory
if [ "$OS" = "windows" ]; then
    # Default for Windows (assuming Git Bash or similar)
    INSTALL_DIR="/usr/bin"
else
    if [ "$(id -u)" -eq 0 ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        # Try to find a good user bin directory
        if [ -d "$HOME/.local/bin" ]; then
            INSTALL_DIR="$HOME/.local/bin"
        elif [ -d "$HOME/bin" ]; then
            INSTALL_DIR="$HOME/bin"
        else
            INSTALL_DIR="$HOME/.local/bin"
            mkdir -p "$INSTALL_DIR"
        fi
    fi
fi

info "Default installation directory: ${YELLOW}${INSTALL_DIR}${NC}"
printf "Enter installation directory [${INSTALL_DIR}]: "
read custom_install_dir < /dev/tty
INSTALL_DIR=${custom_install_dir:-$INSTALL_DIR}

# Ensure directory exists
if [ ! -d "$INSTALL_DIR" ]; then
    info "Creating directory ${INSTALL_DIR}..."
    mkdir -p "$INSTALL_DIR" || sudo mkdir -p "$INSTALL_DIR"
fi

# Full path for the binary
BINARY_NAME="chronix"
if [ "$PLATFORM" = "windows-amd64" ]; then
    BINARY_NAME="chronix.exe"
fi
INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"

# Download
DOWNLOAD_URL="https://dist.chronixhq.com/latest/${PLATFORM}/chronix"
if [ "$PLATFORM" = "windows-amd64" ]; then
    DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
fi

info "Downloading ${BINARY_NAME} from ${DOWNLOAD_URL}..."
TMP_BIN="/tmp/chronix-$$.tmp"
if curl -fsSL --progress-bar "$DOWNLOAD_URL" -o "$TMP_BIN"; then
    success "Download successful."
else
    error "Download failed. Please check the URL or your connection."
    rm -f "$TMP_BIN"
    exit 1
fi

# Move to destination
if [ -w "$INSTALL_DIR" ]; then
    mv -f "$TMP_BIN" "$INSTALL_PATH"
else
    info "Permission denied for ${INSTALL_DIR}. Using sudo..."
    sudo mv -f "$TMP_BIN" "$INSTALL_PATH"
fi

# Set permissions
if [ "$OS" != "windows" ]; then
    if [ -w "$INSTALL_PATH" ]; then
        chmod +x "$INSTALL_PATH"
    else
        sudo chmod +x "$INSTALL_PATH"
    fi
fi

success "Chronix has been installed to ${INSTALL_PATH}"

# Check if INSTALL_DIR is in PATH
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) warn "${INSTALL_DIR} is not in your PATH. You may need to add it to your shell profile." ;;
esac

printf "\n${YELLOW}Post-installation steps:${NC}\n"
printf "1. Run '${BINARY_NAME}' to initialize the server.\n"
printf "2. To install as a system service: '${BINARY_NAME} service install'\n"
printf "   Note: Initialization must be done in the foreground first.\n\n"

printf "Do you want to start Chronix now? [Y/n] "
read start_now < /dev/tty
start_now=${start_now:-Y}

case "$start_now" in
    [Yy]*)
        printf "Which protocols should be started?\n"
        printf "  1) HTTP only\n"
        printf "  2) HTTPS only\n"
        printf "  3) Both\n"
        printf "Select [3]: "
        read proto_choice < /dev/tty
        proto_choice=${proto_choice:-3}

        RUN_FLAGS=""
        START_HTTP=true
        START_HTTPS=true

        case "$proto_choice" in
            1) 
                RUN_FLAGS="--disable-https"
                START_HTTPS=false
                ;;
            2) 
                RUN_FLAGS="--disable-http"
                START_HTTP=false
                ;;
            3) ;;
            *) warn "Invalid choice. Using defaults (Both).";;
        esac

        if [ "$START_HTTP" = true ]; then
            printf "HTTP port [5170]: "
            read http_port < /dev/tty
            http_port=${http_port:-5170}
            RUN_FLAGS="$RUN_FLAGS --force-http-port=$http_port"
        fi

        if [ "$START_HTTPS" = true ]; then
            printf "HTTPS port [5171]: "
            read https_port < /dev/tty
            https_port=${https_port:-5171}
            RUN_FLAGS="$RUN_FLAGS --force-https-port=$https_port"
        fi

        info "Starting Chronix in the foreground for initialization..."
        # shellcheck disable=SC2086
        "$INSTALL_PATH" run $RUN_FLAGS
        ;;
    *)
        info "You can start Chronix later by running '${BINARY_NAME} run'."
        ;;
esac
