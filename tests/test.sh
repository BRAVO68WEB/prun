#!/bin/bash
# Integration test for prun

set -e

# Get the script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PRUN="$PROJECT_ROOT/prun"

echo "=== Running prun integration tests ==="
echo ""

# Test 1: Help flag
echo "Test 1: --help flag"
"$PRUN" --help > /dev/null
echo "✓ Help flag works"
echo ""

# Test 2: List tasks
echo "Test 2: --list flag"
"$PRUN" -c "$SCRIPT_DIR/../examples/simple.toml" --list > /dev/null
echo "✓ List flag works"
echo ""

# Test 3: Run all tasks
echo "Test 3: Running all tasks from simple.toml"
"$PRUN" -c "$SCRIPT_DIR/../examples/simple.toml" > /tmp/prun-output.txt 2>&1
if grep -q "\[hello\]" /tmp/prun-output.txt && grep -q "\[world\]" /tmp/prun-output.txt; then
    echo "✓ Tasks ran in parallel with prefixes"
else
    echo "✗ Output missing task prefixes"
    exit 1
fi
echo ""

# Test 4: Run specific task
echo "Test 4: Running specific task only"
"$PRUN" -c "$SCRIPT_DIR/../examples/simple.toml" hello > /tmp/prun-single.txt 2>&1
if grep -q "\[hello\]" /tmp/prun-single.txt && ! grep -q "\[world\]" /tmp/prun-single.txt; then
    echo "✓ Single task selection works"
else
    echo "✗ Task selection failed"
    exit 1
fi
echo ""

# Test 5: Missing config file
echo "Test 5: Missing config file handling"
cd /tmp
if "$PRUN" 2>&1 | grep -q "no prun.toml found"; then
    echo "✓ Missing config handled correctly"
else
    echo "✗ Missing config error message incorrect"
    exit 1
fi
cd - > /dev/null
echo ""

# Test 6: Error handling
echo "Test 6: Error handling (task failure)"
if ! "$PRUN" -c "$SCRIPT_DIR/error-test.toml" > /dev/null 2>&1; then
    echo "✓ Failed task causes non-zero exit"
else
    echo "✗ Failed task did not return error"
    exit 1
fi
echo ""

echo "=== All tests passed! ==="
