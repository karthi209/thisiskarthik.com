#!/bin/bash
# Install Dependencies Script
# Installs all required dependencies for the static site generator

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Installing dependencies for static site generator...${NC}"
echo ""

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo -e "${RED}Error: Cannot detect operating system${NC}"
    exit 1
fi

install_package() {
    local package=$1
    local name=$2
    local command=${3:-$package}  # Use provided command name or default to package name
    
    if command -v "$command" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $name already installed"
        return 0
    fi
    
    echo -e "${YELLOW}Installing $name...${NC}"
    
    case $OS in
        ubuntu|debian)
            sudo apt-get update -qq
            sudo apt-get install -y "$package"
            ;;
        fedora|rhel|centos)
            sudo dnf install -y "$package"
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm "$package"
            ;;
        *)
            echo -e "${RED}Error: Unsupported OS: $OS${NC}"
            echo "Please install $name manually"
            return 1
            ;;
    esac
    
    if command -v "$command" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $name installed successfully"
    else
        echo -e "${RED}✗${NC} Failed to install $name"
        return 1
    fi
}

# 1. Install Go
echo -e "${BLUE}[1/4] Checking Go...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Go already installed: $GO_VERSION"
else
    echo -e "${YELLOW}Go not found. Installing...${NC}"
    
    case $OS in
        ubuntu|debian)
            # Install Go from official source (newer version)
            GO_VERSION="1.21.5"
            ARCH=$(uname -m)
            if [ "$ARCH" = "x86_64" ]; then
                ARCH="amd64"
            elif [ "$ARCH" = "aarch64" ]; then
                ARCH="arm64"
            fi
            
            cd /tmp
            wget -q "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-${ARCH}.tar.gz"
            rm "go${GO_VERSION}.linux-${ARCH}.tar.gz"
            
            # Add to PATH if not already there
            if ! grep -q '/usr/local/go/bin' ~/.bashrc 2>/dev/null; then
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
            fi
            export PATH=$PATH:/usr/local/go/bin
            ;;
        fedora|rhel|centos)
            sudo dnf install -y golang
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm go
            ;;
        *)
            echo -e "${RED}Error: Cannot auto-install Go on $OS${NC}"
            echo "Please install Go manually from https://go.dev/dl/"
            exit 1
            ;;
    esac
    
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        echo -e "${GREEN}✓${NC} Go installed: $GO_VERSION"
    else
        echo -e "${YELLOW}Note: Go installed but not in PATH. Run: export PATH=\$PATH:/usr/local/go/bin${NC}"
        echo -e "${YELLOW}Or restart your terminal.${NC}"
    fi
fi
echo ""

# 2. Install cwebp (WebP tools)
echo -e "${BLUE}[2/4] Checking WebP tools (cwebp)...${NC}"
case $OS in
    ubuntu|debian)
        install_package "webp" "WebP tools (cwebp)" "cwebp"
        ;;
    fedora|rhel|centos)
        install_package "libwebp-tools" "WebP tools (cwebp)" "cwebp"
        ;;
    arch|manjaro)
        install_package "libwebp" "WebP tools (cwebp)" "cwebp"
        ;;
    *)
        echo -e "${YELLOW}Please install webp package manually${NC}"
        ;;
esac
echo ""

# 3. Install ImageMagick (optional but recommended)
echo -e "${BLUE}[3/4] Checking ImageMagick...${NC}"
# Check if either convert or magick command exists
if command -v convert &> /dev/null || command -v magick &> /dev/null; then
    echo -e "${GREEN}✓${NC} ImageMagick already installed"
else
    echo -e "${YELLOW}Installing ImageMagick...${NC}"
    case $OS in
        ubuntu|debian)
            sudo apt-get update -qq
            sudo apt-get install -y imagemagick
            ;;
        fedora|rhel|centos)
            sudo dnf install -y ImageMagick
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm imagemagick
            ;;
        *)
            echo -e "${YELLOW}Please install ImageMagick manually${NC}"
            ;;
    esac
    
    # Check if installation was successful (either command should work)
    if command -v convert &> /dev/null || command -v magick &> /dev/null; then
        echo -e "${GREEN}✓${NC} ImageMagick installed successfully"
    else
        echo -e "${YELLOW}⚠${NC} ImageMagick package installed but commands not found in PATH"
        echo -e "${YELLOW}   This is optional - image optimization will use cwebp as primary tool${NC}"
    fi
fi
echo ""

# 4. Install Go dependencies
echo -e "${BLUE}[4/4] Installing Go module dependencies...${NC}"
if command -v go &> /dev/null; then
    cd "$(dirname "$0")/.."
    if [ -f "go.mod" ]; then
        echo -e "${YELLOW}Downloading Go modules...${NC}"
        go mod download
        echo -e "${GREEN}✓${NC} Go dependencies installed"
    else
        echo -e "${YELLOW}No go.mod found, skipping Go dependencies${NC}"
    fi
else
    echo -e "${YELLOW}Go not available, skipping Go dependencies${NC}"
    echo -e "${YELLOW}Run 'go mod download' after installing Go${NC}"
fi
echo ""

# Summary
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Installation Summary${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

check_tool() {
    local tool=$1
    local name=$2
    if command -v "$tool" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $name"
    else
        echo -e "${RED}✗${NC} $name (not found)"
    fi
}

check_tool "go" "Go (golang)"
check_tool "cwebp" "WebP tools (cwebp)"
# Check ImageMagick - either convert or magick is sufficient
if command -v convert &> /dev/null || command -v magick &> /dev/null; then
    echo -e "${GREEN}✓${NC} ImageMagick"
else
    echo -e "${RED}✗${NC} ImageMagick (not found)"
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo -e "${YELLOW}Note:${NC} If Go was just installed, you may need to:"
echo "  1. Restart your terminal, or"
echo "  2. Run: export PATH=\$PATH:/usr/local/go/bin"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  - Run 'make generate' to build the site"
echo "  - Run 'make optimize' to optimize images"

