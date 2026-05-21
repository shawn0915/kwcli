#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="shawn0915/kwcli"
BINARY_NAME="kwcli"
INSTALL_DIR="${HOME}/.kwcli/bin"
BASHRC="${HOME}/.bashrc"

# Detect OS
detect_os() {
    if [[ -f /etc/redhat-release ]]; then
        echo "redhat"
    elif [[ -f /etc/centos-release ]]; then
        echo "centos"
    elif [[ -f /etc/debian_version ]]; then
        echo "debian"
    elif [[ -f /etc/lsb-release ]]; then
        . /etc/lsb-release
        case "$DISTRIB_ID" in
            Ubuntu|Debian)
                echo "debian"
                ;;
            *)
                echo "unknown"
                ;;
        esac
    else
        echo "unknown"
    fi
}

# Get OS version for RedHat/CentOS
get_redhat_version() {
    if [[ -f /etc/redhat-release ]]; then
        grep -oE '[0-9]+' /etc/redhat-release | head -1
    elif [[ -f /etc/centos-release ]]; then
        grep -oE '[0-9]+' /etc/centos-release | head -1
    else
        echo "0"
    fi
}

# Check if running as root on RedHat-based system
is_root_on_redhat() {
    [[ $EUID -eq 0 ]] && [[ "$(detect_os)" == "redhat" || "$(detect_os)" == "centos" ]]
}

# Get latest release version
get_latest_version() {
    curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep -o '"tag_name":.*"' | cut -d'"' -f4 | sed 's/v//'
}

# Detect architecture
get_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7l)
            echo "armv7"
            ;;
        i386|i686)
            echo "386"
            ;;
        *)
            echo "amd64"
            ;;
    esac
}

# Detect OS for binary name
get_os() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux)
            echo "linux"
            ;;
        darwin)
            echo "darwin"
            ;;
        *)
            echo "linux"
            ;;
    esac
}

# Download and install binary
install_binary() {
    local version=$1
    local os=$(get_os)
    local arch=$(get_arch)
    local download_url="https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${os}-${arch}.tar.gz"
    local temp_dir=$(mktemp -d)
    
    echo -e "${YELLOW}Downloading KWCLI v${version} for ${os}-${arch}...${NC}"
    
    # Try to download tar.gz first, fallback to single binary
    if curl -sL "${download_url}" -o "${temp_dir}/kwcli.tar.gz" 2>/dev/null; then
        tar -xzf "${temp_dir}/kwcli.tar.gz" -C "${temp_dir}"
        mv "${temp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        # Install TSBS binaries if present
        if [[ -d "${temp_dir}/bin" ]]; then
            echo -e "${YELLOW}Installing TSBS binaries...${NC}"
            for f in "${temp_dir}/bin"/*; do
                [[ -f "$f" ]] && mv "$f" "${INSTALL_DIR}/"
            done
        fi
    else
        # Fallback: try direct binary download
        local direct_url="https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${os}-${arch}"
        echo -e "${YELLOW}Trying direct binary download...${NC}"
        curl -sL "${direct_url}" -o "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || true
    rm -rf "${temp_dir}"
}

# Add to PATH for RedHat/CentOS 7+
add_to_path() {
    local path_line="export PATH=\${HOME}/.kwcli/bin:\$PATH"
    
    if [[ -f "${BASHRC}" ]]; then
        if ! grep -q "${INSTALL_DIR}" "${BASHRC}" 2>/dev/null; then
            echo "" >> "${BASHRC}"
            echo "# KWCLI" >> "${BASHRC}"
            echo "${path_line}" >> "${BASHRC}"
            echo -e "${GREEN}Added PATH to ${BASHRC}${NC}"
        else
            echo -e "${YELLOW}PATH already configured in ${BASHRC}${NC}"
        fi
    else
        echo "${path_line}" > "${BASHRC}"
        echo -e "${GREEN}Created ${BASHRC} with PATH configuration${NC}"
    fi
}

# Main installation
main() {
    echo -e "${GREEN}KWCLI Installer${NC}"
    echo "====================="
    
    # Check if already installed
    if command -v kwcli &> /dev/null; then
        echo -e "${YELLOW}KWCLI is already installed:${NC} $(kwcli --version)"
        read -p "Do you want to upgrade? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 0
        fi
    fi
    
    # Get latest version
    echo -e "${YELLOW}Checking for latest version...${NC}"
    VERSION=$(get_latest_version)
    if [[ -z "${VERSION}" ]]; then
        echo -e "${RED}Failed to get latest version. Using default: 0.1.2${NC}"
        VERSION="0.1.2"
    fi
    echo -e "${GREEN}Latest version: ${VERSION}${NC}"
    
    # Detect OS
    OS=$(detect_os)
    echo -e "Detected OS: ${OS}"
    
    # Create install directory
    mkdir -p "${INSTALL_DIR}"
    
    # Download and install
    install_binary "${VERSION}"
    
    # Configure for RedHat/CentOS 7+
    if [[ "${OS}" == "redhat" ]] || [[ "${OS}" == "centos" ]]; then
        RH_VERSION=$(get_redhat_version)
        if [[ ${RH_VERSION} -ge 7 ]]; then
            echo -e "${GREEN}Detected RedHat/CentOS ${RH_VERSION}, configuring PATH...${NC}"
            add_to_path
            echo ""
            echo -e "${GREEN}========================================${NC}"
            echo -e "${GREEN}Installation completed!${NC}"
            echo -e "${YELLOW}Please run: source ~/.bashrc${NC}"
            echo -e "${GREEN}========================================${NC}"
        else
            echo -e "${YELLOW}RedHat/CentOS ${RH_VERSION} detected, installing to /usr/local/bin${NC}"
            sudo cp "${INSTALL_DIR}/${BINARY_NAME}" "/usr/local/bin/${BINARY_NAME}"
            sudo chmod +x "/usr/local/bin/${BINARY_NAME}"
        fi
    else
        # For other systems, offer to install to /usr/local/bin
        read -p "Install to /usr/local/bin? (requires sudo, y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            sudo cp "${INSTALL_DIR}/${BINARY_NAME}" "/usr/local/bin/${BINARY_NAME}"
            sudo chmod +x "/usr/local/bin/${BINARY_NAME}"
        fi
    fi
    
    # Verify installation
    if [[ -x "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        echo -e "${GREEN}Installation successful!${NC}"
        echo ""
        "${INSTALL_DIR}/${BINARY_NAME}" --version || true
    else
        echo -e "${RED}Installation failed!${NC}"
        exit 1
    fi
}

main "$@"