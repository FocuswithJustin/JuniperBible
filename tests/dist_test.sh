#!/usr/bin/env bash
# Distribution Build Tests
# Verifies that make dist-local produces a correctly structured distribution
#
# Usage: ./tests/dist_test.sh [--skip-build]
#   --skip-build: Skip rebuilding dist, use existing dist/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_ROOT/dist"
TEST_DIR=""
SKIP_BUILD=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
    esac
done

# Test counters
PASSED=0
FAILED=0

# Test helper functions
pass() {
    echo "  PASS: $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo "  FAIL: $1"
    FAILED=$((FAILED + 1))
}

section() {
    echo ""
    echo "=== $1 ==="
}

# Cleanup function
cleanup() {
    if [ -n "$TEST_DIR" ] && [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
    fi
}
trap cleanup EXIT

# Build distribution if needed
if [ "$SKIP_BUILD" = false ]; then
    section "Building Distribution"
    echo "Running: make dist-clean && make dist-local"
    cd "$PROJECT_ROOT"
    make dist-clean >/dev/null 2>&1
    make dist-local >/dev/null 2>&1
    echo "Build complete."
fi

# Find the distribution tarball
section "Finding Distribution"
TARBALL=$(ls "$DIST_DIR"/*.tar.gz 2>/dev/null | head -1)
if [ -z "$TARBALL" ]; then
    echo "ERROR: No distribution tarball found in $DIST_DIR"
    echo "Run 'make dist-local' first or remove --skip-build flag"
    exit 1
fi
echo "Found: $(basename "$TARBALL")"

# Extract to temp directory
TEST_DIR=$(mktemp -d)
echo "Extracting to: $TEST_DIR"
tar -xzf "$TARBALL" -C "$TEST_DIR"

# Find extracted directory
DIST_ROOT=$(ls -d "$TEST_DIR"/juniper-bible-* 2>/dev/null | head -1)
if [ -z "$DIST_ROOT" ]; then
    echo "ERROR: Could not find extracted distribution directory"
    exit 1
fi
echo "Distribution root: $DIST_ROOT"

# =============================================================================
# Test: Required top-level files and directories
# =============================================================================
section "Top-Level Structure"

for dir in bin plugins capsules; do
    if [ -d "$DIST_ROOT/$dir" ]; then
        pass "Directory exists: $dir/"
    else
        fail "Directory missing: $dir/"
    fi
done

for file in README LICENSE; do
    if [ -f "$DIST_ROOT/$file" ]; then
        pass "File exists: $file"
    else
        fail "File missing: $file"
    fi
done

# =============================================================================
# Test: Main binaries
# =============================================================================
section "Main Binaries"

for binary in capsule capsule-web capsule-api juniper.sword; do
    if [ -x "$DIST_ROOT/bin/$binary" ]; then
        pass "Binary exists and executable: bin/$binary"
    else
        fail "Binary missing or not executable: bin/$binary"
    fi
done

# =============================================================================
# Test: Plugin directory structure
# =============================================================================
section "Plugin Directory Structure"

for kind in format tool; do
    if [ -d "$DIST_ROOT/plugins/$kind" ]; then
        pass "Plugin kind directory exists: plugins/$kind/"
    else
        fail "Plugin kind directory missing: plugins/$kind/"
    fi
done

# Check that format plugins are in subdirectories (not flat binaries)
FLAT_FORMAT=$(find "$DIST_ROOT/plugins/format" -maxdepth 1 -type f -name "format-*" 2>/dev/null | wc -l)
if [ "$FLAT_FORMAT" -eq 0 ]; then
    pass "No flat binaries in plugins/format/ (correct structure)"
else
    fail "Found $FLAT_FORMAT flat binaries in plugins/format/ (should be in subdirectories)"
fi

FLAT_TOOL=$(find "$DIST_ROOT/plugins/tool" -maxdepth 1 -type f -name "tool-*" 2>/dev/null | wc -l)
if [ "$FLAT_TOOL" -eq 0 ]; then
    pass "No flat binaries in plugins/tool/ (correct structure)"
else
    fail "Found $FLAT_TOOL flat binaries in plugins/tool/ (should be in subdirectories)"
fi

# =============================================================================
# Test: Format plugins have proper structure
# =============================================================================
section "Format Plugin Structure"

for plugin in osis usfm sword sword-pure esword json sqlite; do
    plugin_dir="$DIST_ROOT/plugins/format/$plugin"
    if [ -d "$plugin_dir" ]; then
        if [ -f "$plugin_dir/plugin.json" ]; then
            pass "Plugin $plugin has plugin.json"
        else
            fail "Plugin $plugin missing plugin.json"
        fi

        if [ -x "$plugin_dir/format-$plugin" ]; then
            pass "Plugin $plugin has executable binary"
        else
            fail "Plugin $plugin missing executable binary"
        fi
    else
        fail "Plugin directory missing: plugins/format/$plugin/"
    fi
done

# =============================================================================
# Test: Tool plugins have proper structure
# =============================================================================
section "Tool Plugin Structure"

for plugin in sqlite repoman hugo libsword pandoc; do
    plugin_dir="$DIST_ROOT/plugins/tool/$plugin"
    if [ -d "$plugin_dir" ]; then
        if [ -f "$plugin_dir/plugin.json" ]; then
            pass "Tool plugin $plugin has plugin.json"
        else
            fail "Tool plugin $plugin missing plugin.json"
        fi

        if [ -x "$plugin_dir/tool-$plugin" ]; then
            pass "Tool plugin $plugin has executable binary"
        else
            fail "Tool plugin $plugin missing executable binary"
        fi
    else
        fail "Tool plugin directory missing: plugins/tool/$plugin/"
    fi
done

# =============================================================================
# Test: Plugin.json file counts
# =============================================================================
section "Plugin.json Counts"

FORMAT_JSON_COUNT=$(find "$DIST_ROOT/plugins/format" -name "plugin.json" | wc -l)
TOOL_JSON_COUNT=$(find "$DIST_ROOT/plugins/tool" -name "plugin.json" | wc -l)
TOTAL_JSON_COUNT=$((FORMAT_JSON_COUNT + TOOL_JSON_COUNT))

if [ "$FORMAT_JSON_COUNT" -ge 10 ]; then
    pass "Found $FORMAT_JSON_COUNT format plugin.json files (expected >= 10)"
else
    fail "Found only $FORMAT_JSON_COUNT format plugin.json files (expected >= 10)"
fi

if [ "$TOOL_JSON_COUNT" -ge 5 ]; then
    pass "Found $TOOL_JSON_COUNT tool plugin.json files (expected >= 5)"
else
    fail "Found only $TOOL_JSON_COUNT tool plugin.json files (expected >= 5)"
fi

echo "Total plugin.json files: $TOTAL_JSON_COUNT"

# =============================================================================
# Test: Capsules directory
# =============================================================================
section "Capsules Directory"

CAPSULE_COUNT=$(find "$DIST_ROOT/capsules" -name "*.tar.gz" -o -name "*.tar.xz" 2>/dev/null | wc -l)
if [ "$CAPSULE_COUNT" -gt 0 ]; then
    pass "Found $CAPSULE_COUNT sample capsules"
else
    fail "No sample capsules found in capsules/"
fi

# =============================================================================
# Test: Binary functionality (quick smoke tests)
# =============================================================================
section "Binary Smoke Tests"

if "$DIST_ROOT/bin/capsule" --help >/dev/null 2>&1; then
    pass "capsule --help works"
else
    fail "capsule --help failed"
fi

if "$DIST_ROOT/bin/capsule-web" --help 2>&1 | grep -q "port"; then
    pass "capsule-web shows help with port option"
else
    fail "capsule-web help doesn't show expected options"
fi

if "$DIST_ROOT/bin/juniper.sword" help 2>&1 | grep -q "list"; then
    pass "juniper.sword help shows list command"
else
    fail "juniper.sword help doesn't show expected commands"
fi

# =============================================================================
# Test: Plugin discovery simulation
# =============================================================================
section "Plugin Discovery Simulation"

DISCOVERABLE=$(find "$DIST_ROOT/plugins" -name "plugin.json" -exec dirname {} \; | wc -l)
if [ "$DISCOVERABLE" -ge 15 ]; then
    pass "At least 15 plugins would be discovered ($DISCOVERABLE found)"
else
    fail "Only $DISCOVERABLE plugins would be discovered (expected >= 15)"
fi

# =============================================================================
# Summary
# =============================================================================
section "Summary"
echo ""
echo "Tests passed: $PASSED"
echo "Tests failed: $FAILED"
TOTAL=$((PASSED + FAILED))
echo "Total tests:  $TOTAL"
echo ""

if [ "$FAILED" -eq 0 ]; then
    echo "All distribution tests passed!"
    exit 0
else
    echo "$FAILED test(s) failed!"
    exit 1
fi
