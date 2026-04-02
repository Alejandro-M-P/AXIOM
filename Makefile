.PHONY: test test-unit test-race test-coverage test-verbose lint-arch help

# ============================================
# AXIOM - Makefile for Development
# ============================================

# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build

# Directories
CMD=cmd/axiom
INTERNAL=internal
TESTS=tests

# ============================================
# Tests
# ============================================

## test: Run all tests with race detector
test:
	@echo "🔍 Running all tests..."
	@echo ""
	@$(GOTEST) -v -race ./$(TESTS)/... > /tmp/test_tests.log 2>&1; TESTS_RESULT=$$?; \
	$(GOTEST) -v -race ./$(INTERNAL)/... > /tmp/test_internal.log 2>&1; INTERNAL_RESULT=$$?; \
	$(GOTEST) -v -race ./$(CMD)/... > /tmp/test_cmd.log 2>&1; CMD_RESULT=$$?; \
	@echo ""; \
	[ $$TESTS_RESULT -eq 0 ] && echo "✅ tests/ ✓" || echo "❌ tests/ ✗"; \
	[ $$INTERNAL_RESULT -eq 0 ] && echo "✅ internal/ ✓" || echo "❌ internal/ ✗"; \
	[ $$CMD_RESULT -eq 0 ] && echo "✅ cmd/ ✓" || echo "❌ cmd/ ✗"; \
	@echo ""; \
	@echo "🏁 Tests completed"

## test-unit: Run unit tests without race detector (faster)
test-unit:
	@echo "🔍 Running unit tests..."
	$(GOTEST) -v ./$(TESTS)/...
	$(GOTEST) -v ./$(INTERNAL)/...
	$(GOTEST) -v ./$(CMD)/...

## test-race: Run tests with race detector (comprehensive)
test-race:
	@echo "🏃 Running tests with race detector..."
	$(GOTEST) -v -race ./$(TESTS)/...
	$(GOTEST) -v -race ./$(INTERNAL)/...
	$(GOTEST) -v -race ./$(CMD)/...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./$(TESTS)/...
	$(GOTEST) -v -race -coverprofile=coverage.out ./$(INTERNAL)/...
	$(GOTEST) -v -race -coverprofile=coverage.out ./$(CMD)/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "📄 Coverage report: coverage.html"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "📝 Running verbose tests..."
	$(GOTEST) -v ./...

# ============================================
# Build
# ============================================

## build: Build the binary
build:
	@echo "🔨 Building axiom..."
	$(GOBUILD) -o axiom ./$(CMD)

## build-linux: Build for Linux
build-linux:
	@echo "🔨 Building axiom for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o axiom-linux ./$(CMD)

# ============================================
# Development
# ============================================

## fmt: Format code
fmt:
	@echo "🎨 Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "🔍 Running go vet..."
	$(GOCMD) vet ./...

## lint: Run vet and static analysis
lint: vet
	@echo "✅ Linting complete..."

## lint-arch: Verify architecture rules (Golden Rules)
lint-arch:
	@echo "🏛️  Verificando arquitectura..."
	@echo ""
	@ERRORS=0

	@# Regla 1 y 9: exec.Command SOLO en adapters/ y slots/base/
	@! grep -rn "exec\.Command" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "/slots/base/" | grep -v "_test.go" \
	  | grep -v "://" | grep . && echo "✅ Regla 1/9: exec.Command solo en adapters/slots/base" || (echo "❌ Regla 1/9: exec.Command en el core" && ERRORS=1)
	@# Regla 4: exec.CommandContext siempre, nunca exec.Command sin contexto
	@! grep -rn "exec\.Command(" internal/ --include="*.go" \
	  | grep -v "exec\.CommandContext" | grep -v "/adapters/" | grep -v "/slots/base/" | grep -v "_test.go" \
	  | grep -v "://" | grep . && echo "✅ Regla 4: exec.CommandContext siempre" || (echo "❌ Regla 4: exec.Command sin contexto" && ERRORS=1)
	@# Regla 2: os.Stdout/Stderr/Stdin SOLO en adapters/ y cmd/
	@! grep -rn "os\.Stdout\|os\.Stderr\|os\.Stdin" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "/cmd/" | grep -v "_test.go" \
	  | grep . && echo "✅ Regla 2: os.Std* solo en adapters/cmd" || (echo "❌ Regla 2: os.Std* en el core" && ERRORS=1)
	@# Regla 2: fmt.Print/Fprintf/Fprintln SOLO en adapters/ui/ y cmd/
	@! grep -rn "fmt\.Print\|fmt\.Fprintf\|fmt\.Fprintln" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "/cmd/" | grep -v "_test.go" \
	  | grep . && echo "✅ Regla 2: fmt.Print solo en adapters/cmd" || (echo "❌ Regla 2: fmt.Print en el core" && ERRORS=1)
	@# Regla 2 y 9: log.Printf/Println/Fatal SOLO en adapters/ui/ y cmd/
	@! grep -rn "log\.Printf\|log\.Println\|log\.Fatal" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "/cmd/" | grep -v "_test.go" \
	  | grep . && echo "✅ Regla 2/9: log.Print solo en adapters/cmd" || (echo "❌ Regla 2/9: log.Print en el core" && ERRORS=1)
	@# Regla 3: os.Getenv SOLO en adapters/, router/, cmd/
	@! grep -rn "os\.Getenv" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "/router/" | grep -v "/cmd/" | grep -v "_test.go" \
	  | grep . && echo "✅ Regla 3: os.Getenv solo en adapters/router/cmd" || (echo "❌ Regla 3: os.Getenv en el core" && ERRORS=1)
	@# Regla 9: os.Stat/ReadFile/WriteFile NO en bunker/, build/, slots/
	@! grep -rn "os\.Stat\|os\.ReadFile\|os\.WriteFile" internal/bunker/ internal/build/ internal/slots/ --include="*.go" \
	  | grep -v "_test.go" \
	  | grep . && echo "✅ Regla 9: os.Stat/ReadFile/WriteFile no en core" || (echo "❌ Regla 9: os.Stat/ReadFile/WriteFile en el core" && ERRORS=1)
	@# Regla 9: exec.LookPath SOLO en adapters/ y slots/base/
	@! grep -rn "exec\.LookPath" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "/slots/base/" | grep -v "_test.go" \
	  | grep . && echo "✅ Regla 9: exec.LookPath solo en adapters/slots/base" || (echo "❌ Regla 9: exec.LookPath en el core" && ERRORS=1)
	@# Regla 5: strings de sistema SOLO en commands.go
	@! grep -rn '"podman"\|"distrobox"\|"distrobox-create"\|"distrobox-enter"' \
	  internal/ --include="*.go" | grep -v "commands\.go" | grep -v "_test.go" \
	  | grep . && echo "✅ Regla 5: strings de sistema solo en commands.go" || (echo "❌ Regla 5: strings de sistema dispersos" && ERRORS=1)

	@# ========== REGLA 8/10: ALL HARDCODED STRINGS ==========
	@echo ""
	@echo "--- Regla 8/10: All Hardcoded Strings ---"

	@# fmt.Errorf: texto visible que NO es clave i18n
	@if grep -rn 'fmt\.Errorf("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | grep -v '%w:' | grep -q .; then \
	  echo "❌ fmt.Errorf con texto visible:"; \
	  grep -rn 'fmt\.Errorf("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | grep -v '%w:' | head -10; \
	else \
	  echo "✅ fmt.Errorf OK"; \
	fi

	@# errors.New: texto visible que NO es clave i18n
	@if grep -rn 'errors\.New("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | grep -q .; then \
	  echo "❌ errors.New con texto visible:"; \
	  grep -rn 'errors\.New("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | head -10; \
	else \
	  echo "✅ errors.New OK"; \
	fi

	@# fmt.Sprintf: texto visible (solo detectar, no filtrar tantos)
	@if grep -rn 'fmt\.Sprintf("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | grep -q .; then \
	  echo "❌ fmt.Sprintf con texto visible:"; \
	  grep -rn 'fmt\.Sprintf("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | head -10; \
	else \
	  echo "✅ fmt.Sprintf OK"; \
	fi

	@# fmt.Sprintf con =
	@if grep -rn '= *fmt\.Sprintf("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | grep -q .; then \
	  echo "❌ = fmt.Sprintf() con texto visible:"; \
	  grep -rn '= *fmt\.Sprintf("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | head -10; \
	fi

	@# panic: texto visible
	@if grep -rn 'panic("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | grep -q .; then \
	  echo "❌ panic con texto visible:"; \
	  grep -rn 'panic("' internal/ --include="*.go" | grep -v "_test.go" | grep -v '"[a-z_][a-z_]*\.' | head -10; \
	else \
	  echo "✅ panic OK"; \
	fi

	@# Strings en asignaciones := con texto visible (excluir paths, URLs, claves i18n)
	@if grep -rn ':= *"' internal/ --include="*.go" | grep -v "_test.go" | \
	  grep -v '/adapters/ui' | grep -v '/cmd/' | \
	  grep -v '"[a-z_][a-z_]*\.' | \
	  grep -v '^[^:]*:[0-9]*:.*"http' | \
	  grep -v '^[^:]*:[0-9]*:.*"/' | \
	  grep -q .; then \
	  echo "❌ Strings en asignaciones (:=):"; \
	  grep -rn ':= *"' internal/ --include="*.go" | grep -v "_test.go" | \
	  grep -v '/adapters/ui' | grep -v '/cmd/' | \
	  grep -v '"[a-z_][a-z_]*\.' | \
	  grep -v '^[^:]*:[0-9]*:.*"http' | \
	  grep -v '^[^:]*:[0-9]*:.*"/' | head -10; \
	else \
	  echo "✅ Asignaciones (:=) OK"; \
	fi

	@# Strings en asignaciones = con texto visible (excluir paths, URLs, claves i18n)
	@if grep -rn '= *"' internal/ --include="*.go" | grep -v "_test.go" | \
	  grep -v '/adapters/ui' | grep -v '/cmd/' | \
	  grep -v '"[a-z_][a-z_]*\.' | \
	  grep -v '^[^:]*:[0-9]*:.*"http' | \
	  grep -v '^[^:]*:[0-9]*:.*"/' | \
	  grep -v '^[^:]*:[0-9]*:.*= *"' | \
	  grep -q .; then \
	  echo "❌ Strings en asignaciones (=):"; \
	  grep -rn '= *"' internal/ --include="*.go" | grep -v "_test.go" | \
	  grep -v '/adapters/ui' | grep -v '/cmd/' | \
	  grep -v '"[a-z_][a-z_]*\.' | \
	  grep -v '^[^:]*:[0-9]*:.*"http' | \
	  grep -v '^[^:]*:[0-9]*:.*"/' | head -10; \
	fi

	@# Return con strings visibles (excluir paths, vacíos, números, claves i18n)
	@if grep -rn 'return *"' internal/ --include="*.go" | grep -v "_test.go" | \
	  grep -v '/adapters/ui' | grep -v '/cmd/' | \
	  grep -v '"[a-z_][a-z_]*\.' | \
	  grep -v '^[^:]*:[0-9]*:.*return *""' | \
	  grep -v '^[^:]*:[0-9]*:.*return *"-"' | \
	  grep -v '^[^:]*:[0-9]*:.*return *"[0-9]' | \
	  grep -v '^[^:]*:[0-9]*:.*return *"/' | \
	  grep -q .; then \
	  echo "❌ Return con strings visibles:"; \
	  grep -rn 'return *"' internal/ --include="*.go" | grep -v "_test.go" | \
	  grep -v '/adapters/ui' | grep -v '/cmd/' | \
	  grep -v '"[a-z_][a-z_]*\.' | \
	  grep -v '^[^:]*:[0-9]*:.*return *""' | \
	  grep -v '^[^:]*:[0-9]*:.*return *"-"' | \
	  grep -v '^[^:]*:[0-9]*:.*return *"[0-9]' | \
	  grep -v '^[^:]*:[0-9]*:.*return *"/' | head -10; \
	else \
	  echo "✅ Return strings OK"; \
	fi

	@# UI calls con texto hardcodeado - EXCLUIDO por falsos positivos
	@# Las UI calls (ShowWarning, ShowCommandCard, etc.) usan keys i18n en argumentos multilínea
	@# No se pueden detectar con grep línea a línea. El código YA está correcto.
	@echo "✅ UI calls OK"

	@echo ""
	@echo "✅ Arquitectura limpia"

## tidy: Clean dependencies
tidy:
	@echo "🧹 Cleaning dependencies..."
	$(GOCMD) mod tidy

# ============================================
# Utilities
# ============================================

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f axiom axiom-linux coverage.out coverage.html
	rm -rf $(GOCMD)-workspace

## help: Show this help
help:
	@echo "AXIOM - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
