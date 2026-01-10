# Juniper Bible Makefile
# Build, test, and documentation targets

.PHONY: all build build-cgo test test-full test-cgo test-sqlite-divergence clean check-stray docs docs-verify plugins help web api juniper juniper-meta repoman format-cmd tools-cmd standalone juniper-test juniper-test-base juniper-test-samples juniper-test-comprehensive juniper-test-all juniper-legacy juniper-legacy-extract juniper-legacy-all nix-test-shell

# Output directories
BIN_DIR := bin
DIST_DIR := dist
BUILD_DIR := build

# Default target
# NOTE: web and api standalone binaries temporarily disabled pending decision to deprecate or improve
# The main capsule binary now has integrated `capsule web` and `capsule api` commands
all: build plugins test juniper juniper-legacy docs

# Create output directories
.PHONY: dirs
dirs:
	@mkdir -p $(BIN_DIR) $(DIST_DIR) $(BUILD_DIR)

# Build the main CLI (pure Go, no CGO)
build: dirs
	CGO_ENABLED=0 go build -o $(BIN_DIR)/capsule ./cmd/capsule

# Build with CGO SQLite
build-cgo: dirs
	CGO_ENABLED=1 go build -tags cgo_sqlite -o $(BIN_DIR)/capsule ./cmd/capsule

# Build juniper plugin suite (standalone, pure Go, no CGO)
juniper: dirs
	@echo "Building juniper plugins..."
	cd plugins/format/sword-pure && CGO_ENABLED=0 go build -o $(CURDIR)/$(BIN_DIR)/juniper.sword .
	@echo "Built: $(BIN_DIR)/juniper.sword"
	@echo ""
	@echo "Usage:"
	@echo "  ./$(BIN_DIR)/juniper.sword list              List Bible modules in ~/.sword"
	@echo "  ./$(BIN_DIR)/juniper.sword ingest            Interactive module ingestion"
	@echo "  ./$(BIN_DIR)/juniper.sword ingest KJV DRC    Ingest specific modules"
	@echo "  ./$(BIN_DIR)/juniper.sword help              Show all options"

# Build juniper meta plugin (unified CLI delegating to format/tool plugins)
juniper-meta: dirs
	@echo "Building juniper meta plugin..."
	CGO_ENABLED=0 go build -o $(BIN_DIR)/capsule-juniper ./plugins/meta/juniper
	@echo "Built: $(BIN_DIR)/capsule-juniper"

# Build repoman tool plugin (SWORD repository management)
repoman: dirs
	@echo "Building repoman tool plugin..."
	CGO_ENABLED=0 go build -o $(BIN_DIR)/capsule-repoman ./plugins/tool/repoman
	@echo "Built: $(BIN_DIR)/capsule-repoman"

# Build format command (unified format plugin CLI)
# Note: Avoids conflict with existing 'format' target for code formatting
format-cmd: dirs plugins-format
	@echo "Format plugins are built in $(BIN_DIR)/plugins/format/"
	@echo "Use individual format plugins or the main 'capsule' CLI"

# Build tools command (unified tool plugin CLI)
# Note: Avoids conflict with potential 'tools' target
tools-cmd: dirs plugins-tool
	@echo "Tool plugins are built in $(BIN_DIR)/plugins/tool/"
	@echo "Use individual tool plugins or the main 'capsule' CLI"

# Build all standalone binaries
standalone: juniper juniper-meta repoman
	@echo ""
	@echo "All standalone binaries built in $(BIN_DIR)/"
	@echo "  - juniper.sword (SWORD format plugin)"
	@echo "  - capsule-juniper (meta plugin CLI)"
	@echo "  - capsule-repoman (repository manager)"

# Test juniper with 11 sample Bibles (quick, parallel execution)
juniper-test-base:
	@echo "Running juniper base tests (11 sample Bibles)..."
	cd plugins/format/sword-pure && CGO_ENABLED=0 go test -v -run "TestBase" .

# Test 11 sample Bibles IR extraction and round-trip
juniper-test-samples:
	@echo "Running juniper sample Bible tests (11 Bibles, parallel)..."
	cd plugins/format/sword-pure && CGO_ENABLED=0 go test -v -timeout 5m -run "TestExtractIRSampleBibles|TestZTextRoundTripAllSampleBibles" .

# Test juniper with ALL Bibles from ~/.sword (comprehensive, requires SWORD modules)
# Warning: This can take a long time depending on how many modules are installed
juniper-test-comprehensive:
	@echo "Running juniper comprehensive tests (ALL ~/.sword modules)..."
	@echo "Warning: This tests ALL installed modules and may take a long time."
	cd plugins/format/sword-pure && SWORD_TEST_ALL=1 CGO_ENABLED=0 go test -v -timeout 30m -run "TestComprehensive|TestZTextRoundTripAllSampleBibles" .

# Run all juniper tests (samples only, use juniper-test-all for comprehensive)
juniper-test: juniper-test-base juniper-test-samples
	@echo ""
	@echo "Sample tests complete. Run 'make juniper-test-comprehensive' for full ~/.sword testing."

# Run comprehensive juniper tests (all installed modules)
juniper-test-all: juniper-test-base
	@echo "Running comprehensive juniper tests (all modules)..."
	cd plugins/format/sword-pure && SWORD_TEST_ALL=1 CGO_ENABLED=0 go test -v -timeout 60m ./...

# Build legacy juniper from contrib reference implementation (for comparison testing)
# Note: Requires CGO (uses mattn/go-sqlite3)
# Source: contrib/tool/juniper/src/
JUNIPER_LEGACY_SRC := contrib/tool/juniper/src

juniper-legacy: dirs
	@echo "Building capsule-juniper-legacy from contrib reference..."
	@echo "Note: This builds the OLD buggy implementation for comparison testing"
	@if [ ! -d "$(JUNIPER_LEGACY_SRC)" ]; then \
		echo "ERROR: Reference juniper source not found at $(JUNIPER_LEGACY_SRC)"; \
		exit 1; \
	fi
	cd $(JUNIPER_LEGACY_SRC) && CGO_ENABLED=1 go build -o $(CURDIR)/$(BIN_DIR)/capsule-juniper-legacy ./cmd/juniper
	@echo "Built: $(BIN_DIR)/capsule-juniper-legacy"

juniper-legacy-extract: dirs
	@echo "Building capsule-juniper-legacy-extract from contrib reference..."
	@if [ ! -d "$(JUNIPER_LEGACY_SRC)" ]; then \
		echo "ERROR: Reference juniper source not found at $(JUNIPER_LEGACY_SRC)"; \
		exit 1; \
	fi
	cd $(JUNIPER_LEGACY_SRC) && CGO_ENABLED=1 go build -o $(CURDIR)/$(BIN_DIR)/capsule-juniper-legacy-extract ./cmd/extract
	@echo "Built: $(BIN_DIR)/capsule-juniper-legacy-extract"

juniper-legacy-all: juniper-legacy juniper-legacy-extract
	@echo "All legacy juniper binaries built in $(BIN_DIR)/"

# Build all plugins (to bin/plugins/, not source directories)
plugins: plugins-format plugins-tool

plugins-format: dirs
	@echo "Building format plugins to $(BIN_DIR)/plugins/format/..."
	@mkdir -p $(BIN_DIR)/plugins/format
	@for plugin in file zip dir tar sword sword-pure osis usfm usx zefania theword json sqlite markdown html epub esword mysword dbl accordance flex gobible logos morphgnt odf onlinebible oshb pdb rtf sblgnt sfm tei txt xml; do \
		if [ -d "plugins/format/$$plugin" ]; then \
			mkdir -p $(BIN_DIR)/plugins/format/$$plugin; \
			if [ -f "plugins/format/$$plugin/go.mod" ]; then \
				(cd plugins/format/$$plugin && go build -o $(CURDIR)/$(BIN_DIR)/plugins/format/$$plugin/format-$$plugin . 2>/dev/null) || true; \
			else \
				go build -o $(BIN_DIR)/plugins/format/$$plugin/format-$$plugin ./plugins/format/$$plugin 2>/dev/null || true; \
			fi; \
			if [ -f "plugins/format/$$plugin/plugin.json" ]; then \
				cp plugins/format/$$plugin/plugin.json $(BIN_DIR)/plugins/format/$$plugin/; \
			fi; \
		fi; \
	done

plugins-tool: dirs
	@echo "Building tool plugins to $(BIN_DIR)/plugins/tool/..."
	@mkdir -p $(BIN_DIR)/plugins/tool
	@for plugin in libsword pandoc calibre usfm2osis sqlite libxml2 unrtf gobible-creator repoman hugo; do \
		if [ -d "plugins/tool/$$plugin" ]; then \
			mkdir -p $(BIN_DIR)/plugins/tool/$$plugin; \
			if [ -f "plugins/tool/$$plugin/go.mod" ]; then \
				(cd plugins/tool/$$plugin && go build -o $(CURDIR)/$(BIN_DIR)/plugins/tool/$$plugin/tool-$$plugin . 2>/dev/null) || true; \
			else \
				go build -o $(BIN_DIR)/plugins/tool/$$plugin/tool-$$plugin ./plugins/tool/$$plugin 2>/dev/null || true; \
			fi; \
			if [ -f "plugins/tool/$$plugin/plugin.json" ]; then \
				cp plugins/tool/$$plugin/plugin.json $(BIN_DIR)/plugins/tool/$$plugin/; \
			fi; \
		fi; \
	done

# Run all tests (pure Go, short mode)
# Use 'make test-full' for comprehensive tests including SWORD round-trips
test:
	CGO_ENABLED=0 go test -short ./...

# Run full tests including slow round-trip tests
test-full:
	CGO_ENABLED=0 go test -timeout 10m ./...

# Run integration tests (requires external tools installed)
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/... -timeout 10m

# Run tests with CGO SQLite
test-cgo:
	CGO_ENABLED=1 go test -tags cgo_sqlite ./...

# Run SQLite divergence tests with both drivers
test-sqlite-divergence:
	@echo "Testing SQLite divergence (pure Go)..."
	@CGO_ENABLED=0 go test ./core/sqlite/... -v -run Divergence 2>&1 | tee /tmp/sqlite-purego.txt
	@echo ""
	@echo "Testing SQLite divergence (CGO)..."
	@CGO_ENABLED=1 go test -tags cgo_sqlite ./core/sqlite/... -v -run Divergence 2>&1 | tee /tmp/sqlite-cgo.txt
	@echo ""
	@echo "Comparing hashes..."
	@grep "Divergence hash:" /tmp/sqlite-purego.txt > /tmp/hash1.txt
	@grep "Divergence hash:" /tmp/sqlite-cgo.txt > /tmp/hash2.txt
	@if diff -q /tmp/hash1.txt /tmp/hash2.txt > /dev/null; then \
		echo "SUCCESS: Both drivers produce identical results"; \
		cat /tmp/hash1.txt; \
	else \
		echo "FAILURE: Drivers have diverged!"; \
		echo "Pure Go:"; cat /tmp/hash1.txt; \
		echo "CGO:"; cat /tmp/hash2.txt; \
		exit 1; \
	fi

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run functional tests (CLI and Web UI workflows)
test-functional:
	./tests/functional_test.sh

# Run quick functional tests
test-functional-quick:
	./tests/functional_test.sh --quick

# Run distribution build tests (verifies dist structure)
test-dist:
	./tests/dist_test.sh

# Run distribution tests without rebuilding (use existing dist/)
test-dist-quick:
	./tests/dist_test.sh --skip-build

# Enter Nix shell with all integration test tools
nix-test-shell:
	@echo "Entering Nix shell with integration test tools..."
	@echo "Available tools: diatheke, ebook-convert, pandoc, xmllint, sqlite3, unrtf"
	nix-shell tests/integration/shell.nix

# Clean build artifacts
# ALL binaries should be in bin/, build/, or dist/ - never in project root
clean:
	rm -rf $(BIN_DIR) $(DIST_DIR) $(BUILD_DIR)
	@# Remove any stray binaries from project root (should never exist)
	rm -f capsule capsule-web juniper.sword
	rm -f capsule-juniper capsule-repoman
	rm -f capsule-juniper-legacy capsule-juniper-legacy-extract
	rm -f sword-pure juniper-sword juniper-esword juniper-hugo juniper-repoman
	rm -f libsword pandoc calibre usfm2osis sqlite libxml2 unrtf gobible-creator zip tar
	@# Remove test artifacts
	rm -f *.out *.test coverage.html
	rm -f tools/docgen/docgen tools/migrate-ipc/migrate-ipc
	find . -name "*.out" -type f -delete 2>/dev/null || true
	find . -name "*.test" -type f -delete 2>/dev/null || true
	find . -name "coverage.html" -type f -delete 2>/dev/null || true
	@# Remove plugin binaries from source directories
	find plugins -name "format-*" -type f -executable -delete 2>/dev/null || true
	find plugins -name "tool-*" -type f -executable -delete 2>/dev/null || true
	find plugins -name "juniper-*" -type f -executable -delete 2>/dev/null || true
	find plugins -name "sword-pure" -type f -executable -delete 2>/dev/null || true

# Check for stray binaries in project root (should be empty)
.PHONY: check-stray
check-stray:
	@STRAY=$$(find . -maxdepth 1 -type f -executable ! -name "*.sh" ! -name "*.py" 2>/dev/null); \
	if [ -n "$$STRAY" ]; then \
		echo "ERROR: Found stray binaries in project root:"; \
		echo "$$STRAY"; \
		echo ""; \
		echo "All binaries should be in bin/, build/, or dist/"; \
		echo "Run 'make clean' to remove them"; \
		exit 1; \
	else \
		echo "OK: No stray binaries in project root"; \
	fi

# Generate documentation
docs: build plugins
	@echo "Generating documentation..."
	@mkdir -p docs/generated
	./$(BIN_DIR)/capsule dev docgen all --output docs/generated
	@echo "Documentation generated in docs/generated/"

# Verify documentation is up to date
docs-verify: docs
	@echo "Checking if documentation is up to date..."
	@if git diff --quiet docs/generated/; then \
		echo "Documentation is up to date."; \
	else \
		echo "ERROR: Documentation is out of date. Run 'make docs' and commit."; \
		git diff --stat docs/generated/; \
		exit 1; \
	fi

# =============================================================================
# Sample Data Generation
# =============================================================================

# Sample Bible modules to include (must be in ~/.sword)
SAMPLE_MODULES := KJV ASV DRC LXX OEB OSMHB SBLGNT Tyndale Vulgate WEB Geneva1599

# Generate sample capsules with IR from ~/.sword modules
.PHONY: sample-data sample-data-clean sample-data-test
sample-data: juniper
	@echo "Generating sample capsules with IR..."
	@mkdir -p contrib/sample-data/capsules
	@for mod in $(SAMPLE_MODULES); do \
		if [ -d "$$HOME/.sword/modules" ]; then \
			./$(BIN_DIR)/juniper.sword ingest $$mod --output contrib/sample-data/capsules/ 2>/dev/null || \
			echo "  Skipping $$mod (not found in ~/.sword)"; \
		fi; \
	done
	@echo "Sample capsules generated in contrib/sample-data/capsules/"
	@ls -la contrib/sample-data/capsules/*.tar.gz 2>/dev/null || echo "No capsules found"

# Clean sample data
sample-data-clean:
	rm -rf contrib/sample-data/capsules/*.tar.gz

# Test sample data workflow
sample-data-test: build juniper
	@echo "Testing sample data workflow..."
	@CAPSULE=$$(ls contrib/sample-data/capsules/*.tar.gz 2>/dev/null | head -1); \
	if [ -z "$$CAPSULE" ]; then \
		echo "ERROR: No sample capsules found. Run 'make sample-data' first."; \
		exit 1; \
	fi; \
	echo "Testing with: $$CAPSULE"; \
	echo "  1. Verify capsule..."; \
	./$(BIN_DIR)/capsule capsule verify $$CAPSULE || exit 1; \
	echo "  2. Check IR exists..."; \
	tar -tzf $$CAPSULE | grep -q "ir/" && echo "     IR found" || (echo "     ERROR: No IR in capsule"; exit 1); \
	echo "  3. Selfcheck..."; \
	./$(BIN_DIR)/capsule capsule selfcheck $$CAPSULE || exit 1; \
	echo "  Sample data workflow test PASSED"

# Full workflow test on all sample capsules
.PHONY: test-workflow
test-workflow: build juniper
	@echo "Running full workflow tests on sample capsules..."
	@PASSED=0; FAILED=0; TOTAL=0; \
	for capsule in contrib/sample-data/capsules/*.tar.gz; do \
		if [ -f "$$capsule" ]; then \
			TOTAL=$$((TOTAL + 1)); \
			NAME=$$(basename $$capsule); \
			echo ""; \
			echo "Testing: $$NAME"; \
			if ./$(BIN_DIR)/capsule capsule verify $$capsule >/dev/null 2>&1 && \
			   tar -tzf $$capsule 2>/dev/null | grep -q "ir/" && \
			   ./$(BIN_DIR)/capsule capsule selfcheck $$capsule >/dev/null 2>&1; then \
				echo "  PASSED"; \
				PASSED=$$((PASSED + 1)); \
			else \
				echo "  FAILED"; \
				FAILED=$$((FAILED + 1)); \
			fi; \
		fi; \
	done; \
	echo ""; \
	echo "==========================================="; \
	echo "Workflow test results: $$PASSED/$$TOTAL passed"; \
	if [ $$FAILED -gt 0 ]; then \
		echo "$$FAILED tests FAILED"; \
		exit 1; \
	fi

# Format Go code
fmt:
	gofmt -w .

# Run linting
# Note: core/ir uses participle parser grammar tags that go vet doesn't understand
lint:
	@echo "Running go vet (excluding core/ir participle grammar files)..."
	@go vet $$(go list ./... | grep -v '/core/ir$$')
	@echo "Running go vet on core/ir without structtag check..."
	@go vet -structtag=false ./core/ir/... 2>/dev/null || true
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

# Install development dependencies
dev-deps:
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Consider installing golangci-lint for additional linting"

# Run the self-check on example capsule
selfcheck: build
	./$(BIN_DIR)/capsule selfcheck testdata/examples/sample.capsule.tar.xz

# Verify example capsule
verify: build
	./$(BIN_DIR)/capsule verify testdata/examples/sample.capsule.tar.xz

# List available plugins
list-plugins: build plugins
	./$(BIN_DIR)/capsule plugins

# Run integration tests
integration: build plugins
	./$(BIN_DIR)/capsule test testdata/fixtures

# Generate API documentation (placeholder)
docs-api:
	@echo "API documentation generation not yet implemented"
	@echo "See docs/IR_IMPLEMENTATION.md for IR API"
	@echo "See docs/PLUGIN_DEVELOPMENT.md for plugin API"

# =============================================================================
# Distribution Build Targets
# =============================================================================

# Version for distribution
VERSION := 0.1.0

# All supported platforms (pure Go builds for all)
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# CGO-capable platforms (get both CGO and pure-Go builds)
# These platforms have reliable CGO cross-compilation or native builds
# CGO builds use mattn/go-sqlite3 (native SQLite, ~10% faster for large DBs)
# Pure Go builds use modernc.org/sqlite (portable, no C compiler needed)
CGO_PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Distribution contents
DIST_BINARIES := capsule juniper.sword
FORMAT_PLUGINS := file zip dir tar sword sword-pure osis usfm usx zefania theword json sqlite markdown html epub esword mysword dbl accordance flex gobible logos morphgnt odf onlinebible oshb pdb rtf sblgnt sfm tei txt xml
TOOL_PLUGINS := libsword pandoc calibre usfm2osis sqlite libxml2 unrtf gobible-creator repoman hugo

# Build all distributions
# - Pure Go builds for ALL platforms
# - CGO builds for CGO_PLATFORMS (linux, darwin)
.PHONY: dist dist-all dist-clean dist-purego dist-cgo

dist: dist-clean dist-purego dist-cgo
	@echo ""
	@echo "All distributions built in $(DIST_DIR)/"
	@echo ""
	@echo "Pure Go builds (all platforms):"
	@ls -1 $(DIST_DIR)/*-purego.tar.gz 2>/dev/null || echo "  (none)"
	@echo ""
	@echo "CGO builds (linux, darwin):"
	@ls -1 $(DIST_DIR)/*-cgo.tar.gz 2>/dev/null || echo "  (none)"

# Build pure-Go distributions for all platforms
dist-purego:
	@echo "Building pure-Go distributions for all platforms..."
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		$(MAKE) dist-platform-purego GOOS=$$GOOS GOARCH=$$GOARCH; \
	done

# Build CGO distributions for supported platforms (native build only)
# Note: Cross-compiling CGO requires platform-specific C toolchains
dist-cgo:
	@echo "Building CGO distribution for current platform..."
	@CURRENT_OS=$$(go env GOOS); \
	CURRENT_ARCH=$$(go env GOARCH); \
	PLATFORM="$$CURRENT_OS/$$CURRENT_ARCH"; \
	if echo "$(CGO_PLATFORMS)" | grep -q "$$PLATFORM"; then \
		$(MAKE) dist-platform-cgo GOOS=$$CURRENT_OS GOARCH=$$CURRENT_ARCH; \
	else \
		echo "  Skipping CGO build for $$PLATFORM (not in CGO_PLATFORMS)"; \
	fi

# Build pure-Go distribution for a single platform
.PHONY: dist-platform-purego
dist-platform-purego:
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "ERROR: GOOS and GOARCH must be set"; \
		exit 1; \
	fi
	@echo ""
	@echo "=========================================="
	@echo "Building pure-Go for $(GOOS)/$(GOARCH)..."
	@echo "=========================================="
	@$(MAKE) dist-build-purego GOOS=$(GOOS) GOARCH=$(GOARCH)
	@$(MAKE) dist-package GOOS=$(GOOS) GOARCH=$(GOARCH) VARIANT=purego

# Build CGO distribution for a single platform
.PHONY: dist-platform-cgo
dist-platform-cgo:
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "ERROR: GOOS and GOARCH must be set"; \
		exit 1; \
	fi
	@echo ""
	@echo "=========================================="
	@echo "Building CGO for $(GOOS)/$(GOARCH)..."
	@echo "=========================================="
	@$(MAKE) dist-build-cgo GOOS=$(GOOS) GOARCH=$(GOARCH)
	@$(MAKE) dist-package GOOS=$(GOOS) GOARCH=$(GOARCH) VARIANT=cgo

# Build pure-Go binaries for a platform
# Uses modernc.org/sqlite (pure Go, portable, ~30 extra dependencies)
.PHONY: dist-build-purego
dist-build-purego:
	@DIST_BUILD=$(BUILD_DIR)/$(GOOS)-$(GOARCH)-purego; \
	mkdir -p $$DIST_BUILD/bin $$DIST_BUILD/plugins/example/noop; \
	EXT=""; \
	if [ "$(GOOS)" = "windows" ]; then EXT=".exe"; fi; \
	echo "  Building main binaries (pure Go, modernc.org/sqlite)..."; \
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $$DIST_BUILD/bin/capsule$$EXT ./cmd/capsule; \
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $$DIST_BUILD/bin/juniper.sword$$EXT ./plugins/format/sword-pure; \
	echo "  Building noop placeholder plugin..."; \
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $$DIST_BUILD/plugins/example/noop/example-noop$$EXT ./plugins/example/noop; \
	cp plugins/example/noop/plugin.json $$DIST_BUILD/plugins/example/noop/; \
	cp plugins/example/noop/README.md $$DIST_BUILD/plugins/example/noop/

# Build CGO binaries for a platform
# Uses mattn/go-sqlite3 (native SQLite via CGO, faster, fewer Go dependencies)
.PHONY: dist-build-cgo
dist-build-cgo:
	@DIST_BUILD=$(BUILD_DIR)/$(GOOS)-$(GOARCH)-cgo; \
	mkdir -p $$DIST_BUILD/bin $$DIST_BUILD/plugins/example/noop; \
	EXT=""; \
	if [ "$(GOOS)" = "windows" ]; then EXT=".exe"; fi; \
	echo "  Building main binaries (CGO, mattn/go-sqlite3)..."; \
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags cgo_sqlite -o $$DIST_BUILD/bin/capsule$$EXT ./cmd/capsule; \
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags cgo_sqlite -o $$DIST_BUILD/bin/juniper.sword$$EXT ./plugins/format/sword-pure; \
	echo "  Building noop placeholder plugin..."; \
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $$DIST_BUILD/plugins/example/noop/example-noop$$EXT ./plugins/example/noop; \
	cp plugins/example/noop/plugin.json $$DIST_BUILD/plugins/example/noop/; \
	cp plugins/example/noop/README.md $$DIST_BUILD/plugins/example/noop/

# Package distribution for a platform (used by both purego and cgo)
# VARIANT must be set to 'purego' or 'cgo'
.PHONY: dist-package
dist-package:
	@if [ -z "$(VARIANT)" ]; then \
		echo "ERROR: VARIANT must be set (purego or cgo)"; \
		exit 1; \
	fi
	@DIST_BUILD=$(BUILD_DIR)/$(GOOS)-$(GOARCH)-$(VARIANT); \
	DIST_NAME=juniper-bible-$(VERSION)-$(GOOS)-$(GOARCH)-$(VARIANT); \
	DIST_PKG=$(DIST_DIR)/$$DIST_NAME; \
	mkdir -p $$DIST_PKG; \
	echo "  Packaging $$DIST_NAME..."; \
	cp -r $$DIST_BUILD/bin $$DIST_PKG/; \
	cp -r $$DIST_BUILD/plugins $$DIST_PKG/; \
	mkdir -p $$DIST_PKG/capsules; \
	if [ -d "contrib/sample-data/capsules" ]; then \
		cp -r contrib/sample-data/capsules/* $$DIST_PKG/capsules/ 2>/dev/null || true; \
	fi; \
	cp README $$DIST_PKG/ 2>/dev/null || true; \
	cp LICENSE $$DIST_PKG/ 2>/dev/null || true; \
	cp CHANGELOG $$DIST_PKG/ 2>/dev/null || true; \
	echo "  Creating archives..."; \
	(cd $(DIST_DIR) && tar -czf $$DIST_NAME.tar.gz $$DIST_NAME); \
	(cd $(DIST_DIR) && tar -cJf $$DIST_NAME.tar.xz $$DIST_NAME); \
	(cd $(DIST_DIR) && zip -rq $$DIST_NAME.zip $$DIST_NAME); \
	rm -rf $$DIST_PKG; \
	echo "  Created: $$DIST_NAME.tar.gz, $$DIST_NAME.tar.xz, $$DIST_NAME.zip"

# Build distribution for current platform only (both variants if CGO supported)
.PHONY: dist-local
dist-local:
	@echo "Building distributions for current platform..."
	@CURRENT_OS=$$(go env GOOS); \
	CURRENT_ARCH=$$(go env GOARCH); \
	$(MAKE) dist-platform-purego GOOS=$$CURRENT_OS GOARCH=$$CURRENT_ARCH; \
	PLATFORM="$$CURRENT_OS/$$CURRENT_ARCH"; \
	if echo "$(CGO_PLATFORMS)" | grep -q "$$PLATFORM"; then \
		$(MAKE) dist-platform-cgo GOOS=$$CURRENT_OS GOARCH=$$CURRENT_ARCH; \
	fi

# Clean distribution artifacts
dist-clean:
	rm -rf $(DIST_DIR)
	rm -rf $(BUILD_DIR)
	mkdir -p $(DIST_DIR) $(BUILD_DIR)

# List distribution files
.PHONY: dist-list
dist-list:
	@echo "Distribution files in $(DIST_DIR)/:"
	@ls -lah $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.tar.xz $(DIST_DIR)/*.zip 2>/dev/null || echo "No distribution files found. Run 'make dist' first."

# =============================================================================
# Help
# =============================================================================

# Show help
help:
	@echo "Juniper Bible Makefile"
	@echo ""
	@echo "Build outputs go to: $(BIN_DIR)/"
	@echo ""
	@echo "Build targets:"
	@echo "  make build                    - Build the capsule CLI to $(BIN_DIR)/"
	@echo "  make web                      - Build the Web UI server to $(BIN_DIR)/"
	@echo "  make api                      - Build the REST API server to $(BIN_DIR)/"
	@echo "  make plugins                  - Build all plugins (in plugins/)"
	@echo "  make juniper                  - Build juniper.sword to $(BIN_DIR)/"
	@echo "  make juniper-meta             - Build capsule-juniper meta plugin to $(BIN_DIR)/"
	@echo "  make repoman                  - Build capsule-repoman to $(BIN_DIR)/"
	@echo "  make format-cmd               - Build format plugins to $(BIN_DIR)/plugins/format/"
	@echo "  make tools-cmd                - Build tool plugins to $(BIN_DIR)/plugins/tool/"
	@echo "  make standalone               - Build all standalone binaries"
	@echo "  make juniper-legacy           - Build legacy juniper (CGO) to $(BIN_DIR)/"
	@echo "  make juniper-legacy-all       - Build all legacy juniper binaries"
	@echo "  make all                      - Build everything and run tests"
	@echo ""
	@echo "Test targets:"
	@echo "  make test                     - Run quick tests (short mode)"
	@echo "  make test-full                - Run ALL tests (may take 10+ minutes)"
	@echo "  make test-v                   - Run tests with verbose output"
	@echo "  make test-coverage            - Run tests with coverage report"
	@echo "  make integration              - Run integration tests"
	@echo "  make juniper-test             - Run juniper sample tests (11 Bibles)"
	@echo "  make juniper-test-base        - Run juniper base tests only"
	@echo "  make juniper-test-samples     - Run 11 sample Bible IR + round-trip tests"
	@echo "  make juniper-test-comprehensive - Run tests with all ~/.sword modules"
	@echo "  make juniper-test-all         - Run ALL juniper tests (comprehensive)"
	@echo ""
	@echo "Documentation targets:"
	@echo "  make docs          - Generate documentation"
	@echo "  make docs-verify   - Verify docs are up to date"
	@echo ""
	@echo "Distribution targets:"
	@echo "  make dist          - Build all distributions (pure-Go + CGO where supported)"
	@echo "  make dist-purego   - Build pure-Go distributions for all platforms"
	@echo "  make dist-cgo      - Build CGO distribution for current platform"
	@echo "  make dist-local    - Build distributions for current platform only"
	@echo "  make dist-list     - List distribution files"
	@echo "  make dist-clean    - Clean distribution artifacts"
	@echo ""
	@echo "Utility targets:"
	@echo "  make clean         - Remove build artifacts ($(BIN_DIR)/, $(DIST_DIR)/, $(BUILD_DIR)/)"
	@echo "  make fmt           - Format Go code"
	@echo "  make lint          - Run linting"
	@echo "  make selfcheck     - Run self-check on example"
	@echo "  make verify        - Verify example capsule"
	@echo "  make list-plugins  - List available plugins"
	@echo "  make help          - Show this help"
	@echo ""
	@echo "Distribution variants:"
	@echo "  -purego: Uses modernc.org/sqlite (pure Go, cross-compiles everywhere)"
	@echo "  -cgo:    Uses mattn/go-sqlite3 (native SQLite, ~10% faster)"
