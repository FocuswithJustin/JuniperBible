#!/usr/bin/env bash
# Juniper Bible Functional Test Suite
# Tests all major CLI and workflow functionality
#
# Usage: ./tests/functional_test.sh [--quick] [--verbose]
#
# Options:
#   --quick    Run only essential tests (skip long-running tests)
#   --verbose  Show detailed output
#
# Exit codes:
#   0 - All tests passed
#   1 - One or more tests failed

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_DIR="/tmp/capsule-functional-test-$$"
CAPSULE_BIN="$PROJECT_ROOT/bin/capsule"
CAPSULE_WEB_BIN="$PROJECT_ROOT/bin/capsule-web"
SAMPLE_DATA="$PROJECT_ROOT/contrib/sample-data"

# Test modules (KJV, DRC, Vulgate, ASV, Geneva as specified)
TEST_MODULES=("kjv" "drc" "vulgate" "asv" "geneva1599")

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Options
QUICK_MODE=false
VERBOSE=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --quick)
            QUICK_MODE=true
            ;;
        --verbose)
            VERBOSE=true
            ;;
    esac
done

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "\n${GREEN}=== TEST: $1 ===${NC}"
}

pass_test() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++)) || true
}

fail_test() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++)) || true
}

skip_test() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((TESTS_SKIPPED++)) || true
}

cleanup() {
    log_info "Cleaning up test directory..."
    rm -rf "$TEST_DIR"
    # Kill any background web server
    pkill -f "capsule-web.*8899" 2>/dev/null || true
}

trap cleanup EXIT

# Setup
setup() {
    log_info "Setting up test environment..."
    mkdir -p "$TEST_DIR"
    cd "$PROJECT_ROOT"

    # Build binaries if needed
    if [[ ! -f "$CAPSULE_BIN" ]]; then
        log_info "Building capsule CLI..."
        CGO_ENABLED=0 go build -o "$CAPSULE_BIN" ./cmd/capsule
    fi

    if [[ ! -f "$CAPSULE_WEB_BIN" ]]; then
        log_info "Building capsule-web..."
        go build -o "$CAPSULE_WEB_BIN" ./cmd/capsule-web
    fi

    # Build sword-pure plugin if needed
    if [[ ! -f "$PROJECT_ROOT/plugins/format/sword-pure/format-sword-pure" ]]; then
        log_info "Building sword-pure plugin..."
        (cd "$PROJECT_ROOT/plugins/format/sword-pure" && go build -o format-sword-pure .)
    fi
}

# Test 1: Juniper List
test_juniper_list() {
    log_test "Juniper List - Sample Modules"

    for mod in "${TEST_MODULES[@]}"; do
        local mod_path="$SAMPLE_DATA/$mod"
        if [[ -d "$mod_path" ]]; then
            local output=$("$CAPSULE_BIN" juniper list "$mod_path" 2>&1)
            if echo "$output" | grep -qi "modules"; then
                pass_test "juniper list $mod"
            else
                fail_test "juniper list $mod - unexpected output"
                $VERBOSE && echo "$output"
            fi
        else
            skip_test "juniper list $mod - directory not found"
        fi
    done
}

# Test 2: Juniper Ingest
test_juniper_ingest() {
    log_test "Juniper Ingest - Create Capsules"

    for mod in "${TEST_MODULES[@]}"; do
        local mod_path="$SAMPLE_DATA/$mod"
        # Map directory name to module ID (case-sensitive)
        case "$mod" in
            kjv) mod_upper="KJV" ;;
            drc) mod_upper="DRC" ;;
            vulgate) mod_upper="Vulgate" ;;
            asv) mod_upper="ASV" ;;
            geneva1599) mod_upper="Geneva1599" ;;
            *) mod_upper=$(echo "$mod" | tr '[:lower:]' '[:upper:]') ;;
        esac

        if [[ -d "$mod_path" ]]; then
            local output=$("$CAPSULE_BIN" juniper ingest --path "$mod_path" -o "$TEST_DIR" "$mod_upper" 2>&1)
            if [[ -f "$TEST_DIR/$mod_upper.capsule.tar.gz" ]]; then
                local size=$(stat -f%z "$TEST_DIR/$mod_upper.capsule.tar.gz" 2>/dev/null || stat -c%s "$TEST_DIR/$mod_upper.capsule.tar.gz" 2>/dev/null)
                if [[ $size -gt 0 ]]; then
                    pass_test "juniper ingest $mod_upper (${size} bytes)"
                else
                    fail_test "juniper ingest $mod_upper - empty capsule"
                fi
            else
                fail_test "juniper ingest $mod_upper - capsule not created"
                $VERBOSE && echo "$output"
            fi
        else
            skip_test "juniper ingest $mod - directory not found"
        fi
    done
}

# Test 3: Capsule Ingest (proper capsule format)
test_capsule_ingest() {
    log_test "Capsule Ingest - Create Proper Capsules"

    local conf_file="$SAMPLE_DATA/kjv/mods.d/kjv.conf"
    if [[ -f "$conf_file" ]]; then
        local output=$("$CAPSULE_BIN" capsule ingest "$conf_file" --out "$TEST_DIR/kjv-proper.capsule.tar.gz" 2>&1)
        if [[ -f "$TEST_DIR/kjv-proper.capsule.tar.gz" ]]; then
            pass_test "capsule ingest kjv.conf"
        else
            fail_test "capsule ingest kjv.conf - capsule not created"
            $VERBOSE && echo "$output"
        fi
    else
        skip_test "capsule ingest - kjv.conf not found"
    fi
}

# Test 4: Capsule Verify
test_capsule_verify() {
    log_test "Capsule Verify"

    if [[ -f "$TEST_DIR/kjv-proper.capsule.tar.gz" ]]; then
        local output=$("$CAPSULE_BIN" capsule verify "$TEST_DIR/kjv-proper.capsule.tar.gz" 2>&1)
        if echo "$output" | grep -qi "passed\|ok"; then
            pass_test "capsule verify"
        else
            fail_test "capsule verify - verification failed"
            $VERBOSE && echo "$output"
        fi
    else
        skip_test "capsule verify - capsule not found"
    fi
}

# Test 5: Capsule Export
test_capsule_export() {
    log_test "Capsule Export"

    if [[ -f "$TEST_DIR/kjv-proper.capsule.tar.gz" ]]; then
        local output=$("$CAPSULE_BIN" capsule export "$TEST_DIR/kjv-proper.capsule.tar.gz" --artifact kjv --out "$TEST_DIR/kjv-exported.conf" 2>&1)
        if [[ -f "$TEST_DIR/kjv-exported.conf" ]]; then
            if grep -q "KJV" "$TEST_DIR/kjv-exported.conf"; then
                pass_test "capsule export"
            else
                fail_test "capsule export - content mismatch"
            fi
        else
            fail_test "capsule export - file not created"
            $VERBOSE && echo "$output"
        fi
    else
        skip_test "capsule export - capsule not found"
    fi
}

# Test 6: Format Detect
test_format_detect() {
    log_test "Format Detect"

    local kjv_path="$SAMPLE_DATA/kjv"
    if [[ -d "$kjv_path" ]]; then
        local output=$("$CAPSULE_BIN" format detect "$kjv_path" 2>&1)
        if echo "$output" | grep -qi "sword\|MATCH"; then
            pass_test "format detect kjv"
        else
            fail_test "format detect kjv - SWORD not detected"
            $VERBOSE && echo "$output"
        fi
    else
        skip_test "format detect - kjv not found"
    fi
}

# Test 7: Plugins List
test_plugins_list() {
    log_test "Plugins List"

    local output=$("$CAPSULE_BIN" plugins list 2>&1)
    local format_count=$(echo "$output" | grep -c "format\." || true)

    if [[ $format_count -ge 30 ]]; then
        pass_test "plugins list ($format_count format plugins)"
    else
        fail_test "plugins list - expected 30+ format plugins, got $format_count"
        $VERBOSE && echo "$output"
    fi
}

# Test 8: Web UI Tests (via go test)
test_webui() {
    log_test "Web UI Tests"

    if $QUICK_MODE; then
        skip_test "Web UI tests - skipped in quick mode"
        return
    fi

    local output=$(go test ./cmd/capsule-web/ -v -run "Test" 2>&1 | tail -20)
    if echo "$output" | grep -q "PASS"; then
        local pass_count=$(echo "$output" | grep -c "PASS" || true)
        pass_test "Web UI tests ($pass_count passed)"
    else
        fail_test "Web UI tests - some tests failed"
        $VERBOSE && echo "$output"
    fi
}

# Test 9: Web UI HTTP endpoints
test_webui_endpoints() {
    log_test "Web UI HTTP Endpoints"

    if $QUICK_MODE; then
        skip_test "Web UI HTTP tests - skipped in quick mode"
        return
    fi

    # Start web server
    "$CAPSULE_WEB_BIN" -capsules "$SAMPLE_DATA/capsules" -port 8899 &
    local web_pid=$!
    sleep 2

    # Test endpoints
    local endpoints=("/" "/plugins" "/convert" "/tools" "/sword")
    for endpoint in "${endpoints[@]}"; do
        local response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8899$endpoint" 2>/dev/null || echo "000")
        if [[ "$response" == "200" ]]; then
            pass_test "HTTP GET $endpoint"
        else
            fail_test "HTTP GET $endpoint - status $response"
        fi
    done

    # Kill web server
    kill $web_pid 2>/dev/null || true
}

# Test 10: Go Unit Tests (quick)
test_go_unit() {
    log_test "Go Unit Tests"

    if $QUICK_MODE; then
        # Run only core tests in quick mode
        local output=$(go test ./core/... -short 2>&1 | tail -10)
        if echo "$output" | grep -q "ok"; then
            pass_test "core package tests (quick)"
        else
            fail_test "core package tests failed"
            $VERBOSE && echo "$output"
        fi
    else
        local output=$(go test ./... -short 2>&1 | tail -30)
        local pass_count=$(echo "$output" | grep -c "^ok" || true)
        local fail_count=$(echo "$output" | grep -c "^FAIL" || true)

        if [[ $fail_count -eq 0 ]]; then
            pass_test "All Go tests ($pass_count packages)"
        else
            fail_test "Go tests - $fail_count packages failed"
            $VERBOSE && echo "$output"
        fi
    fi
}

# Test 11: MySword Plugin
test_mysword_plugin() {
    log_test "MySword Plugin"

    local output=$(go test ./plugins/format/mysword/ -v 2>&1 | tail -15)
    if echo "$output" | grep -q "PASS"; then
        local pass_count=$(echo "$output" | grep -c "PASS" || true)
        pass_test "MySword plugin tests ($pass_count passed)"
    else
        fail_test "MySword plugin tests failed"
        $VERBOSE && echo "$output"
    fi
}

# Test 12: CGO Comparison Tests (long running)
test_cgo_comparison() {
    log_test "CGO Comparison Tests"

    if $QUICK_MODE; then
        skip_test "CGO comparison tests - skipped in quick mode"
        return
    fi

    # Check if native tools are available
    if ! command -v diatheke &> /dev/null; then
        skip_test "CGO comparison tests - diatheke not installed"
        return
    fi

    local output=$(go test ./plugins/format/sword-pure/ -run CGOComparison -v -timeout 10m 2>&1 | tail -20)
    if echo "$output" | grep -q "PASS"; then
        pass_test "CGO comparison tests"
    else
        fail_test "CGO comparison tests failed"
        $VERBOSE && echo "$output"
    fi
}

# Main
main() {
    echo "========================================"
    echo "Juniper Bible Functional Test Suite"
    echo "========================================"
    echo "Project: $PROJECT_ROOT"
    echo "Test Dir: $TEST_DIR"
    echo "Quick Mode: $QUICK_MODE"
    echo "Verbose: $VERBOSE"
    echo ""

    setup

    # Run tests
    test_juniper_list
    test_juniper_ingest
    test_capsule_ingest
    test_capsule_verify
    test_capsule_export
    test_format_detect
    test_plugins_list
    test_webui
    test_webui_endpoints
    test_go_unit
    test_mysword_plugin
    test_cgo_comparison

    # Summary
    echo ""
    echo "========================================"
    echo "Test Summary"
    echo "========================================"
    echo -e "${GREEN}Passed:${NC}  $TESTS_PASSED"
    echo -e "${RED}Failed:${NC}  $TESTS_FAILED"
    echo -e "${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
    echo ""

    if [[ $TESTS_FAILED -gt 0 ]]; then
        log_error "Some tests failed!"
        exit 1
    else
        log_info "All tests passed!"
        exit 0
    fi
}

main "$@"
