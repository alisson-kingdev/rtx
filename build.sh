#!/bin/bash
# Cross-platform build script for RTX.
# Supports: Android (Termux), Windows, macOS, Linux.

set -euo pipefail

BINARY_NAME="rtx"
MAIN_PATH="./cmd/rtx"
VERSION="1.0.0"
BUILD_DIR="build"
LDFLAGS="-s -w -X main.version=${VERSION}"

# The tool has no cgo dependencies; disabling cgo keeps cross-compiles portable.
export CGO_ENABLED=0

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "android/arm64"
)

echo "========================================"
echo "  RTX - Build Cross-Plataforma"
echo "========================================"
echo ""

clean_build() {
    rm -f "${BINARY_NAME}"
    rm -rf "${BUILD_DIR}"
    mkdir -p "${BUILD_DIR}"
}

build_current() {
    mkdir -p "${BUILD_DIR}"
    local os
    os="$(go env GOOS)"
    local arch
    arch="$(go env GOARCH)"
    local output="${BUILD_DIR}/${BINARY_NAME}-${os}-${arch}"
    if [ "${os}" = "windows" ]; then
        output="${output}.exe"
    fi
    echo "[*] Build para a plataforma atual (${os}/${arch})..."
    go build -ldflags "${LDFLAGS}" -o "${output}" "${MAIN_PATH}"
    echo "[OK] Binario: ${output}"
}

build_all() {
    clean_build
    echo "[*] Iniciando cross-compile para todas as plataformas..."
    echo ""

    local os arch output
    for platform in "${PLATFORMS[@]}"; do
        os="${platform%%/*}"
        arch="${platform##*/}"

        output="${BUILD_DIR}/${BINARY_NAME}-${os}-${arch}"
        if [ "${os}" = "windows" ]; then
            output="${output}.exe"
        fi

        printf '   [*] %s/%s -> ' "$os" "$arch"
        if GOOS="${os}" GOARCH="${arch}" go build -ldflags "${LDFLAGS}" -o "${output}" "${MAIN_PATH}"; then
            echo "OK"
        else
            echo "FALHA"
            return 1
        fi
    done

    echo ""
    echo "[OK] Build concluido em ${BUILD_DIR}/"
    ls -lh "${BUILD_DIR}/"
}

case "${1:-all}" in
    current)
        build_current
        ;;
    all)
        build_all
        ;;
    clean)
        clean_build
        echo "[OK] Limpeza concluida"
        ;;
    *)
        echo "Uso: $0 {current|all|clean}"
        echo "  current - Build para plataforma atual"
        echo "  all     - Cross-compile para todas as plataformas"
        echo "  clean   - Limpar artefatos"
        exit 1
        ;;
esac
