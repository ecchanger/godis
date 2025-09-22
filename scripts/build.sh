#!/bin/bash

# Build script for Godis Web Todo Server
# Supports multiple platforms and build modes

set -e

# Default values
BINARY_NAME="godis-webtodo"
OUTPUT_DIR="./bin"
BUILD_MODE="release"
PLATFORMS="linux/amd64,darwin/amd64,windows/amd64"
ENABLE_CGO="0"
LDFLAGS=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build Godis Web Todo Server for multiple platforms

OPTIONS:
    -h, --help          Show this help message
    -m, --mode MODE     Build mode: debug|release (default: release)
    -o, --output DIR    Output directory (default: ./bin)
    -p, --platforms LIST Comma-separated list of platforms (default: linux/amd64,darwin/amd64,windows/amd64)
    --single            Build for current platform only
    --docker            Build Docker image
    --clean             Clean build directory before building
    --test              Run tests before building
    --version VERSION   Set version information

EXAMPLES:
    $0                          # Build for all default platforms
    $0 --single                 # Build for current platform only
    $0 --mode debug             # Build debug version
    $0 --docker                 # Build Docker image
    $0 --clean --test           # Clean, test, and build
    $0 --platforms linux/amd64 # Build for Linux AMD64 only

EOF
}

# Function to get version information
get_version() {
    if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
        GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "")
        GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
        BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
        
        if [ -n "$GIT_TAG" ]; then
            VERSION="$GIT_TAG"
        else
            VERSION="dev-$GIT_COMMIT"
        fi
    else
        VERSION="unknown"
        GIT_COMMIT="unknown"
        GIT_BRANCH="unknown"
        BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    fi
}

# Function to set build flags
set_build_flags() {
    get_version
    
    LDFLAGS="-s -w"
    LDFLAGS="$LDFLAGS -X 'main.Version=$VERSION'"
    LDFLAGS="$LDFLAGS -X 'main.GitCommit=$GIT_COMMIT'"
    LDFLAGS="$LDFLAGS -X 'main.GitBranch=$GIT_BRANCH'"
    LDFLAGS="$LDFLAGS -X 'main.BuildTime=$BUILD_TIME'"
    
    if [ "$BUILD_MODE" = "debug" ]; then
        LDFLAGS=""
    fi
}

# Function to clean build directory
clean_build() {
    print_status "Cleaning build directory..."
    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
    print_success "Build directory cleaned"
}

# Function to run tests
run_tests() {
    print_status "Running tests..."
    go test -v ./...
    print_success "All tests passed"
}

# Function to build for single platform
build_single() {
    local goos=$1
    local goarch=$2
    local output_name=$3
    
    print_status "Building for $goos/$goarch..."
    
    env GOOS=$goos GOARCH=$goarch CGO_ENABLED=$ENABLE_CGO \
        go build -ldflags "$LDFLAGS" -o "$output_name" ./cmd/webtodo-server
    
    if [ $? -eq 0 ]; then
        # Get file size
        if command -v du &> /dev/null; then
            size=$(du -h "$output_name" | cut -f1)
            print_success "Built $output_name ($size)"
        else
            print_success "Built $output_name"
        fi
        
        # Copy static files
        local static_dir="$(dirname $output_name)/webtodo"
        mkdir -p "$static_dir"
        cp -r ./webtodo/static "$static_dir/"
        print_status "Copied static files to $static_dir/"
        
        return 0
    else
        print_error "Failed to build for $goos/$goarch"
        return 1
    fi
}

# Function to build for all platforms
build_all() {
    local platforms_array
    IFS=',' read -ra platforms_array <<< "$PLATFORMS"
    
    local failed_builds=0
    
    for platform in "${platforms_array[@]}"; do
        IFS='/' read -ra platform_parts <<< "$platform"
        local goos=${platform_parts[0]}
        local goarch=${platform_parts[1]}
        
        local output_name="$OUTPUT_DIR/${BINARY_NAME}-${goos}-${goarch}"
        if [ "$goos" = "windows" ]; then
            output_name="${output_name}.exe"
        fi
        
        if ! build_single "$goos" "$goarch" "$output_name"; then
            failed_builds=$((failed_builds + 1))
        fi
    done
    
    if [ $failed_builds -eq 0 ]; then
        print_success "All builds completed successfully"
        return 0
    else
        print_error "$failed_builds build(s) failed"
        return 1
    fi
}

# Function to build Docker image
build_docker() {
    print_status "Building Docker image..."
    
    local image_tag="godis-webtodo:latest"
    if [ -n "$CUSTOM_VERSION" ]; then
        image_tag="godis-webtodo:$CUSTOM_VERSION"
    fi
    
    docker build -f Dockerfile.webtodo -t "$image_tag" .
    
    if [ $? -eq 0 ]; then
        print_success "Docker image built: $image_tag"
        return 0
    else
        print_error "Failed to build Docker image"
        return 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -m|--mode)
            BUILD_MODE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -p|--platforms)
            PLATFORMS="$2"
            shift 2
            ;;
        --single)
            PLATFORMS="$(go env GOOS)/$(go env GOARCH)"
            shift
            ;;
        --docker)
            BUILD_DOCKER=true
            shift
            ;;
        --clean)
            CLEAN_BUILD=true
            shift
            ;;
        --test)
            RUN_TESTS=true
            shift
            ;;
        --version)
            CUSTOM_VERSION="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main build process
main() {
    print_status "Starting build process..."
    print_status "Build mode: $BUILD_MODE"
    print_status "Output directory: $OUTPUT_DIR"
    print_status "Platforms: $PLATFORMS"
    
    # Clean build directory if requested
    if [ "$CLEAN_BUILD" = true ]; then
        clean_build
    fi
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    # Run tests if requested
    if [ "$RUN_TESTS" = true ]; then
        run_tests
    fi
    
    # Set build flags
    set_build_flags
    
    # Build Docker image if requested
    if [ "$BUILD_DOCKER" = true ]; then
        build_docker
        return $?
    fi
    
    # Build binaries
    build_all
    return $?
}

# Run main function
main
exit $?