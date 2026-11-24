.PHONY: all build build-all clean test install help

# Configuracao
BINARY_NAME := rtx
MAIN_PATH := ./cmd/rtx
VERSION := 1.0.0
BUILD_DIR := build

# O projeto nao usa cgo; desabilita-lo mantem o cross-compile portatil.
CGO_ENABLED := 0

# Flags de build
LDFLAGS := -s -w -X main.version=$(VERSION)

# Plataformas suportadas
PLATFORMS := \
    linux/amd64 \
    linux/arm64 \
    linux/arm \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64 \
    android/arm64

all: help

help:
	@echo "RTX - Rotacionador de IP"
	@echo ""
	@echo "Comandos disponiveis:"
	@echo "  make build        - Build para a plataforma atual"
	@echo "  make build-all    - Build para todas as plataformas (cross-compile)"
	@echo "  make test         - Rodar testes"
	@echo "  make clean        - Limpar artefatos de build"
	@echo "  make install      - Instalar no GOPATH"

build:
	@echo ">> Build para plataforma atual"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

build-all: clean
	@echo ">> Build cross-plataforma para todas as plataformas"
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output="$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch"; \
		if [ "$$os" = "windows" ]; then \
			output="$$output.exe"; \
		fi; \
		echo "   -> $$os/$$arch"; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o $$output $(MAIN_PATH) || exit 1; \
	done
	@echo ""
	@echo ">> Build concluido em $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

test:
	@echo ">> Rodando testes"
	@go test ./...

clean:
	@echo ">> Limpando artefatos"
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)

install:
	@echo ">> Instalando rtx"
	@go install $(MAIN_PATH)
