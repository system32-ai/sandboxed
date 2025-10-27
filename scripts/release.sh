#!/bin/bash

# GitHub Release Script for Sandboxed
# This script automates the process of creating GitHub releases with cross-platform binaries

set -e  # Exit on any error

# Configuration
REPO_OWNER="system32-ai"
REPO_NAME="sandboxed"
BINARY_NAME="sandboxed"
BUILD_DIR="build"
CHANGELOG_FILE="CHANGELOG.md"
HOMEBREW_TAP_DIR="homebrew-tap"
HOMEBREW_FORMULA_DIR="${HOMEBREW_TAP_DIR}/Formula"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if gh CLI is installed
    if ! command_exists gh; then
        log_error "GitHub CLI (gh) is not installed. Please install it from https://cli.github.com/"
        exit 1
    fi
    
    # Check if user is authenticated with GitHub
    if ! gh auth status >/dev/null 2>&1; then
        log_error "Not authenticated with GitHub. Please run 'gh auth login'"
        exit 1
    fi
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Check if we're on the main/master branch
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
        log_warning "Not on main/master branch (current: $current_branch)"
        read -p "Do you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Check if working directory is clean
    if [[ -n $(git status --porcelain) ]]; then
        log_error "Working directory is not clean. Please commit or stash changes."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to get the next version
get_next_version() {
    local current_version
    local version_type="$1"
    
    # Get the latest tag
    current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
    
    # Remove 'v' prefix if present
    current_version=${current_version#v}
    
    # Split version into parts
    IFS='.' read -ra version_parts <<< "$current_version"
    major=${version_parts[0]:-0}
    minor=${version_parts[1]:-0}
    patch=${version_parts[2]:-0}
    
    # Increment based on type
    case $version_type in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch"|*)
            patch=$((patch + 1))
            ;;
    esac
    
    echo "v${major}.${minor}.${patch}"
}

# Function to build cross-platform binaries
build_binaries() {
    local version="$1"
    
    log_info "Building cross-platform binaries for version $version..."
    
    # Clean build directory
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    
    # Define target platforms
    declare -a platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    
    # Build for each platform
    for platform in "${platforms[@]}"; do
        IFS='/' read -ra platform_parts <<< "$platform"
        os="${platform_parts[0]}"
        arch="${platform_parts[1]}"
        
        output_name="$BINARY_NAME"
        if [[ "$os" == "windows" ]]; then
            output_name="${BINARY_NAME}.exe"
        fi
        
        output_path="$BUILD_DIR/${BINARY_NAME}-${version}-${os}-${arch}"
        if [[ "$os" == "windows" ]]; then
            output_path="${output_path}.exe"
        fi
        
        log_info "Building for $os/$arch..."
        
        GOOS="$os" GOARCH="$arch" go build \
            -ldflags "-X main.version=$version -s -w" \
            -o "$output_path" \
            .
        
        # Create compressed archives for non-Windows platforms
        if [[ "$os" != "windows" ]]; then
            tar_name="${BINARY_NAME}-${version}-${os}-${arch}.tar.gz"
            tar -czf "$BUILD_DIR/$tar_name" -C "$BUILD_DIR" "$(basename "$output_path")"
            rm "$output_path"  # Remove the binary, keep the archive
        else
            # Create zip for Windows
            zip_name="${BINARY_NAME}-${version}-${os}-${arch}.zip"
            (cd "$BUILD_DIR" && zip "$zip_name" "$(basename "$output_path")")
            rm "$output_path"  # Remove the binary, keep the archive
        fi
    done
    
    log_success "Binaries built successfully"
    ls -la "$BUILD_DIR"
}

# Function to generate changelog
generate_changelog() {
    local version="$1"
    local previous_tag="$2"
    
    log_info "Generating changelog for $version..."
    
    local changelog_content=""
    local date=$(date +"%Y-%m-%d")
    
    # Get commits since last tag
    local commits
    if [[ "$previous_tag" != "" ]]; then
        commits=$(git log --oneline --no-merges "${previous_tag}..HEAD" | head -20)
    else
        commits=$(git log --oneline --no-merges | head -20)
    fi
    
    # Create changelog entry
    changelog_content="## [$version] - $date\n\n"
    
    if [[ -n "$commits" ]]; then
        changelog_content+="### Changes\n\n"
        while IFS= read -r commit; do
            # Format: hash message
            commit_hash=$(echo "$commit" | cut -d' ' -f1)
            commit_message=$(echo "$commit" | cut -d' ' -f2-)
            changelog_content+="- $commit_message ([$commit_hash](https://github.com/$REPO_OWNER/$REPO_NAME/commit/$commit_hash))\n"
        done <<< "$commits"
    else
        changelog_content+="### Changes\n\n- Initial release\n"
    fi
    
    changelog_content+="\n"
    
    # Update or create CHANGELOG.md
    if [[ -f "$CHANGELOG_FILE" ]]; then
        # Insert new changelog at the top (after title)
        {
            head -n 2 "$CHANGELOG_FILE"
            echo -e "$changelog_content"
            tail -n +3 "$CHANGELOG_FILE"
        } > "${CHANGELOG_FILE}.tmp"
        mv "${CHANGELOG_FILE}.tmp" "$CHANGELOG_FILE"
    else
        # Create new changelog
        echo "# Changelog" > "$CHANGELOG_FILE"
        echo "" >> "$CHANGELOG_FILE"
        echo -e "$changelog_content" >> "$CHANGELOG_FILE"
    fi
    
    log_success "Changelog updated"
}

# Function to calculate SHA256 hash of a file
calculate_sha256() {
    local file="$1"
    
    if command_exists shasum; then
        shasum -a 256 "$file" | cut -d' ' -f1
    elif command_exists sha256sum; then
        sha256sum "$file" | cut -d' ' -f1
    else
        log_error "Neither shasum nor sha256sum found"
        exit 1
    fi
}

# Function to generate Homebrew formula
generate_homebrew_formula() {
    local version="$1"
    
    log_info "Generating Homebrew formula for version $version..."
    
    # Create homebrew-tap directory structure
    mkdir -p "$HOMEBREW_FORMULA_DIR"
    
    # Download URLs for macOS binaries
    local macos_amd64_url="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$version/${BINARY_NAME}-${version}-darwin-amd64.tar.gz"
    local macos_arm64_url="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$version/${BINARY_NAME}-${version}-darwin-arm64.tar.gz"
    
    # Calculate SHA256 hashes for macOS binaries
    local macos_amd64_file="$BUILD_DIR/${BINARY_NAME}-${version}-darwin-amd64.tar.gz"
    local macos_arm64_file="$BUILD_DIR/${BINARY_NAME}-${version}-darwin-arm64.tar.gz"
    
    local macos_amd64_sha256=""
    local macos_arm64_sha256=""
    
    if [[ -f "$macos_amd64_file" ]]; then
        macos_amd64_sha256=$(calculate_sha256 "$macos_amd64_file")
    else
        log_warning "macOS AMD64 binary not found, SHA256 will be empty"
    fi
    
    if [[ -f "$macos_arm64_file" ]]; then
        macos_arm64_sha256=$(calculate_sha256 "$macos_arm64_file")
    else
        log_warning "macOS ARM64 binary not found, SHA256 will be empty"
    fi
    
    # Generate the Homebrew formula
    local formula_file="${HOMEBREW_FORMULA_DIR}/${BINARY_NAME}.rb"
    
    cat > "$formula_file" << EOF
class Sandboxed < Formula
  desc "A comprehensive sandbox platform for secure code execution in Kubernetes environments"
  homepage "https://github.com/$REPO_OWNER/$REPO_NAME"
  version "$version"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "$macos_amd64_url"
      sha256 "$macos_amd64_sha256"

      def install
        bin.install "$BINARY_NAME"
      end
    end
    if Hardware::CPU.arm?
      url "$macos_arm64_url"
      sha256 "$macos_arm64_sha256"

      def install
        bin.install "$BINARY_NAME"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$version/${BINARY_NAME}-${version}-linux-amd64.tar.gz"
      # SHA256 will need to be updated manually for Linux builds
      # sha256 "LINUX_AMD64_SHA256_HERE"

      def install
        bin.install "$BINARY_NAME"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$version/${BINARY_NAME}-${version}-linux-arm64.tar.gz"
      # SHA256 will need to be updated manually for Linux builds
      # sha256 "LINUX_ARM64_SHA256_HERE"

      def install
        bin.install "$BINARY_NAME"
      end
    end
  end

  def caveats
    <<~EOS
      Sandboxed requires access to a Kubernetes cluster to function properly.
      
      Make sure you have kubectl configured and the necessary RBAC permissions:
      - pods: create, delete, get, list, watch
      - pods/exec: create
      
      For more information, visit: https://github.com/$REPO_OWNER/$REPO_NAME
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/$BINARY_NAME version")
  end
end
EOF
    
    log_success "Homebrew formula generated: $formula_file"
    
    # Create a README for the tap
    local readme_file="${HOMEBREW_TAP_DIR}/README.md"
    cat > "$readme_file" << EOF
# Homebrew Tap for Sandboxed

This is the official Homebrew tap for [Sandboxed](https://github.com/$REPO_OWNER/$REPO_NAME).

## Installation

\`\`\`bash
# Add the tap
brew tap $REPO_OWNER/sandboxed

# Install sandboxed
brew install sandboxed
\`\`\`

## Usage

\`\`\`bash
# Show version
sandboxed version

# Start REST API server
sandboxed server

# Start MCP server for AI integration
sandboxed mcp

# Get help
sandboxed --help
\`\`\`

## Requirements

Sandboxed requires access to a Kubernetes cluster. Make sure you have:

1. kubectl installed and configured
2. Proper RBAC permissions for pod management
3. Access to a Kubernetes cluster

For detailed setup instructions, see the [main repository](https://github.com/$REPO_OWNER/$REPO_NAME).

## Updating

\`\`\`bash
# Update the tap
brew update

# Upgrade sandboxed
brew upgrade sandboxed
\`\`\`
EOF
    
    log_success "Homebrew tap README generated: $readme_file"
    
    # Display the generated formula content
    log_info "Generated Homebrew formula content:"
    echo "----------------------------------------"
    cat "$formula_file"
    echo "----------------------------------------"
    
    # Provide instructions for updating the tap repository
    cat << EOF

📋 To publish this Homebrew formula:

1. Create a new repository named 'homebrew-sandboxed' under $REPO_OWNER organization
2. Copy the contents of '$HOMEBREW_TAP_DIR' to the repository root
3. Commit and push the changes:
   
   cd $HOMEBREW_TAP_DIR
   git init
   git add .
   git commit -m "Add sandboxed formula $version"
   git remote add origin https://github.com/$REPO_OWNER/homebrew-sandboxed.git
   git push -u origin main

4. Users can then install with:
   brew tap $REPO_OWNER/sandboxed
   brew install sandboxed

📝 Note: Linux SHA256 hashes need to be calculated and added manually to the formula.

EOF
}

# Function to create GitHub release
create_github_release() {
    local version="$1"
    local is_prerelease="$2"
    
    log_info "Creating GitHub release $version..."
    
    # Get changelog content for this version
    local release_notes=""
    if [[ -f "$CHANGELOG_FILE" ]]; then
        # Extract content for this version from changelog
        release_notes=$(awk "/## \[$version\]/,/## \[/{if(/## \[/ && !/## \[$version\]/) exit; if(!/## \[$version\]/) print}" "$CHANGELOG_FILE")
    fi
    
    if [[ -z "$release_notes" ]]; then
        release_notes="Release $version"
    fi
    
    # Create the release
    local release_args=(
        "release" "create" "$version"
        "--title" "Release $version"
        "--notes" "$release_notes"
        "--repo" "$REPO_OWNER/$REPO_NAME"
    )
    
    if [[ "$is_prerelease" == "true" ]]; then
        release_args+=("--prerelease")
    fi
    
    # Add all build artifacts
    for file in "$BUILD_DIR"/*; do
        if [[ -f "$file" ]]; then
            release_args+=("$file")
        fi
    done
    
    gh "${release_args[@]}"
    
    log_success "GitHub release created: https://github.com/$REPO_OWNER/$REPO_NAME/releases/tag/$version"
}

# Function to update version in code
update_version_in_code() {
    local version="$1"
    
    log_info "Updating version in code..."
    
    # Update version in main.go if it exists
    if [[ -f "main.go" ]]; then
        sed -i.bak "s/version = \".*\"/version = \"$version\"/" main.go || true
        rm -f main.go.bak
    fi
    
    # Update version in cmd/version.go if it exists
    if [[ -f "cmd/version.go" ]]; then
        sed -i.bak "s/var version = \".*\"/var version = \"$version\"/" cmd/version.go || true
        rm -f cmd/version.go.bak
    fi
    
    log_info "Version updated in code"
}

# Main release function
main() {
    local version_type="${1:-patch}"
    local is_prerelease="${2:-false}"
    
    echo "🚀 GitHub Release Script for $REPO_NAME"
    echo "========================================"
    
    # Check prerequisites
    check_prerequisites
    
    # Get current and next version
    local current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    local next_version
    
    if [[ "$version_type" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        # Custom version provided
        next_version="$version_type"
    else
        # Auto-increment version
        next_version=$(get_next_version "$version_type")
    fi
    
    log_info "Current version: $current_version"
    log_info "Next version: $next_version"
    
    # Confirm release
    read -p "Create release $next_version? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Release cancelled"
        exit 0
    fi
    
    # Update version in code
    update_version_in_code "$next_version"
    
    # Generate changelog
    generate_changelog "$next_version" "$current_version"
    
    # Commit changelog and version updates
    if [[ -n $(git status --porcelain) ]]; then
        git add .
        git commit -m "chore: prepare release $next_version"
        git push origin "$(git branch --show-current)"
    fi
    
    # Create git tag
    git tag -a "$next_version" -m "Release $next_version"
    git push origin "$next_version"
    
    # Build binaries
    build_binaries "$next_version"
    
    # Create GitHub release
    create_github_release "$next_version" "$is_prerelease"
    
    # Generate Homebrew formula
    generate_homebrew_formula "$next_version"
    
    # Cleanup
    rm -rf "$BUILD_DIR"
    
    log_success "Release $next_version completed successfully! 🎉"
    log_info "View the release at: https://github.com/$REPO_OWNER/$REPO_NAME/releases/tag/$next_version"
    log_info "Homebrew formula generated in: $HOMEBREW_TAP_DIR"
}

# Script usage
usage() {
    echo "Usage: $0 [version_type|version] [prerelease]"
    echo ""
    echo "This script automates the release process including:"
    echo "  • Cross-platform binary builds (Linux, macOS, Windows)"
    echo "  • Changelog generation"
    echo "  • GitHub release creation"
    echo "  • Homebrew tap formula generation"
    echo ""
    echo "Arguments:"
    echo "  version_type    One of: patch (default), minor, major"
    echo "  version         Custom version in format: v1.2.3"
    echo "  prerelease      Set to 'true' for prerelease (default: false)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Create patch release"
    echo "  $0 minor              # Create minor release"
    echo "  $0 major              # Create major release"
    echo "  $0 v1.5.0             # Create specific version"
    echo "  $0 patch true         # Create patch prerelease"
    echo ""
    echo "Generated files:"
    echo "  • Cross-platform binaries in build/"
    echo "  • Homebrew formula in homebrew-tap/Formula/sandboxed.rb"
    echo "  • Updated CHANGELOG.md"
    echo ""
}

# Check if help is requested
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    usage
    exit 0
fi

# Run main function
main "$@"