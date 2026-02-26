# HWSCAN Makefile
# Herramienta de detección de hardware con compilación estática

# Variables
BINARY_NAME=hwscan
VERSION=1.0.0
BUILD_DIR=bin
DIST_DIR=dist
CMD_DIR=cmd/hwscan

# Flags de compilación para binarios 100% estáticos
GO=go
GOFLAGS=-v -a
# CRÍTICO: -extldflags "-static" asegura que sea completamente estático
LDFLAGS=-ldflags="-s -w -extldflags '-static' -X main.version=$(VERSION)"
CGO_ENABLED=0

# Colores para output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all build build-amd64 build-arm64 build-armv7 build-all clean test help run install verify-static

# Target por defecto
all: build-amd64

## build: Compilar para la arquitectura actual
build:
	@echo "$(GREEN)Compilando $(BINARY_NAME) para la arquitectura actual...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "$(GREEN)✓ Compilación exitosa: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"
	@$(MAKE) verify-static BIN=$(BUILD_DIR)/$(BINARY_NAME) || true

## build-amd64: Compilar para Linux AMD64 (ESTÁTICO)
build-amd64:
	@echo "$(GREEN)Compilando binario ESTÁTICO para linux/amd64...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Compilación exitosa:$(NC)"
	@echo "  - $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"
	@echo "  - $(DIST_DIR)/$(BINARY_NAME) $(BLUE)(para Alpine)$(NC)"
	@$(MAKE) verify-static BIN=$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 || true

## build-arm64: Compilar para Linux ARM64 (ESTÁTICO)
build-arm64:
	@echo "$(GREEN)Compilando binario ESTÁTICO para linux/arm64...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	@echo "$(GREEN)✓ Compilación exitosa: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64$(NC)"
	@$(MAKE) verify-static BIN=$(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 || true

## build-armv7: Compilar para Linux ARMv7 (ESTÁTICO)
build-armv7:
	@echo "$(GREEN)Compilando binario ESTÁTICO para linux/arm/v7...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(GOFLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv7 ./$(CMD_DIR)
	@echo "$(GREEN)✓ Compilación exitosa: $(BUILD_DIR)/$(BINARY_NAME)-linux-armv7$(NC)"
	@$(MAKE) verify-static BIN=$(BUILD_DIR)/$(BINARY_NAME)-linux-armv7 || true

## build-all: Compilar para todas las arquitecturas
build-all: build-amd64 build-arm64 build-armv7
	@echo ""
	@echo "$(GREEN)✓ Todas las compilaciones completadas$(NC)"
	@echo ""
	@echo "Binarios generados:"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)-*
	@echo ""
	@echo "$(BLUE)Binario para Alpine ISO:$(NC) $(DIST_DIR)/$(BINARY_NAME)"

## verify-static: Verificar que el binario sea estático (uso interno)
verify-static:
	@echo "$(YELLOW)Verificando que el binario sea estático...$(NC)"
	@if [ -z "$(BIN)" ]; then \
		echo "$(RED)Error: Variable BIN no especificada$(NC)"; \
		exit 1; \
	fi
	@if [ ! -f "$(BIN)" ]; then \
		echo "$(RED)Error: Binario no encontrado: $(BIN)$(NC)"; \
		exit 1; \
	fi
	@echo "  Archivo: $(BIN)"
	@file $(BIN) | grep -q "statically linked" && \
		echo "$(GREEN)  ✓ Binario es estático$(NC)" || \
		echo "$(YELLOW)  ⚠ Verificación con 'file' no concluyente$(NC)"
	@if command -v ldd >/dev/null 2>&1; then \
		LDD_OUTPUT=$$(ldd $(BIN) 2>&1); \
		if echo "$$LDD_OUTPUT" | grep -qi "not a dynamic\|no es un ejecutable dinámico\|not a dynamic executable"; then \
			echo "$(GREEN)  ✓ Sin dependencias dinámicas (binario estático confirmado)$(NC)"; \
		elif echo "$$LDD_OUTPUT" | grep -qi "statically linked"; then \
			echo "$(GREEN)  ✓ Binario estáticamente enlazado$(NC)"; \
		else \
			echo "$(RED)  ✗ ADVERTENCIA: Binario puede tener dependencias dinámicas$(NC)"; \
			echo "$$LDD_OUTPUT" | head -5; \
			exit 1; \
		fi; \
	fi
	@SIZE=$$(du -h $(BIN) | cut -f1); \
	echo "$(BLUE)  Tamaño: $$SIZE$(NC)"
	@echo "$(GREEN)  ✓ Verificación completada - Binario listo para Alpine$(NC)"

## verify-all: Verificar todos los binarios compilados
verify-all:
	@echo "$(GREEN)Verificando todos los binarios...$(NC)"
	@for bin in $(BUILD_DIR)/$(BINARY_NAME)-*; do \
		if [ -f "$$bin" ]; then \
			echo ""; \
			$(MAKE) verify-static BIN=$$bin || true; \
		fi; \
	done

## clean: Limpiar binarios y archivos generados
clean:
	@echo "$(YELLOW)Limpiando binarios...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f *.json
	@rm -rf build/alpine
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

## test: Ejecutar tests
test:
	@echo "$(GREEN)Ejecutando tests...$(NC)"
	$(GO) test -v ./...

## run: Compilar y ejecutar
run: build
	@echo "$(GREEN)Ejecutando $(BINARY_NAME)...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

## install: Instalar en /usr/local/bin (requiere sudo)
install: build-amd64
	@echo "$(YELLOW)Instalando $(BINARY_NAME) en /usr/local/bin...$(NC)"
	@sudo cp $(DIST_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ $(BINARY_NAME) instalado exitosamente$(NC)"
	@echo "Ejecuta: $(BINARY_NAME)"

## deps: Verificar dependencias
deps:
	@echo "$(GREEN)Verificando dependencias...$(NC)"
	$(GO) mod download
	$(GO) mod verify
	@echo "$(GREEN)✓ Dependencias verificadas$(NC)"

## fmt: Formatear código
fmt:
	@echo "$(GREEN)Formateando código...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)✓ Código formateado$(NC)"

## vet: Ejecutar go vet
vet:
	@echo "$(GREEN)Ejecutando go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)✓ Verificación completada$(NC)"

## check: Ejecutar todas las verificaciones
check: fmt vet test
	@echo "$(GREEN)✓ Todas las verificaciones completadas$(NC)"

## version: Mostrar versión
version:
	@echo "$(BINARY_NAME) version $(VERSION)"

## size: Mostrar tamaños de binarios
size:
	@echo "$(GREEN)Tamaños de binarios:$(NC)"
	@if [ -d "$(BUILD_DIR)" ]; then \
		ls -lh $(BUILD_DIR)/ | grep $(BINARY_NAME); \
	else \
		echo "$(RED)No hay binarios compilados. Ejecute 'make build-amd64' primero.$(NC)"; \
	fi

## alpine-iso: Compilar y preparar para Alpine ISO (target principal)
alpine-iso: build-amd64
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  ✓ Binario listo para Alpine ISO$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "Ubicación: $(BLUE)$(DIST_DIR)/$(BINARY_NAME)$(NC)"
	@echo ""
	@echo "Siguiente paso:"
	@echo "  $(YELLOW)cd build/alpine/$(NC)"
	@echo "  $(YELLOW)./build.sh$(NC)"
	@echo ""

## alpine-prepare: Preparar estructura completa para Alpine (legacy)
alpine-prepare: build-amd64
	@echo "$(GREEN)Preparando integración con Alpine Linux...$(NC)"
	@mkdir -p build/alpine/rootfs/usr/local/bin
	@mkdir -p build/alpine/rootfs/etc/local.d
	@cp $(DIST_DIR)/$(BINARY_NAME) build/alpine/rootfs/usr/local/bin/$(BINARY_NAME)
	@chmod +x build/alpine/rootfs/usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Archivos preparados en build/alpine/$(NC)"

## help: Mostrar esta ayuda
help:
	@echo ""
	@echo "$(GREEN)╔════════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║  HWSCAN - Hardware Detection Tool         ║$(NC)"
	@echo "$(GREEN)║  Compilación Estática para Alpine Linux   ║$(NC)"
	@echo "$(GREEN)╚════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Targets principales:$(NC)"
	@echo "  $(BLUE)make build-amd64$(NC)    - Compilar binario estático para AMD64"
	@echo "  $(BLUE)make alpine-iso$(NC)     - Preparar binario para Alpine ISO"
	@echo "  $(BLUE)make verify-all$(NC)     - Verificar que todos los binarios sean estáticos"
	@echo "  $(BLUE)make install$(NC)        - Instalar en /usr/local/bin"
	@echo "  $(BLUE)make clean$(NC)          - Limpiar binarios"
	@echo ""
	@echo "$(YELLOW)Otros targets:$(NC)"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/  /'
	@echo ""
	@echo "$(YELLOW)Flujo completo para Alpine ISO:$(NC)"
	@echo "  1. $(GREEN)make alpine-iso$(NC)     $(BLUE)# Compila binario estático$(NC)"
	@echo "  2. $(GREEN)cd scripts/alpine-iso$(NC)"
	@echo "  3. $(GREEN)./build.sh$(NC)          $(BLUE)# Construye ISO$(NC)"
	@echo "  4. $(GREEN)./verify-build.sh$(NC)   $(BLUE)# Verifica ISO$(NC)"
	@echo ""
	@echo "$(YELLOW)Ubicaciones de salida:$(NC)"
	@echo "  - $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64  $(BLUE)(archivo completo)$(NC)"
	@echo "  - $(DIST_DIR)/$(BINARY_NAME)               $(BLUE)(usado por Alpine)$(NC)"
	@echo ""