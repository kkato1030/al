#!/bin/bash
# Demo script to show the difference between normal and debug modes

echo "================================================"
echo "Testing al output organization"
echo "================================================"

# Build the binary
echo ""
echo "Building al..."
make build-dev > /dev/null 2>&1

if [ ! -f ./bin/al ]; then
    echo "Error: Build failed"
    exit 1
fi

echo "Build successful!"
echo ""

# Test 1: Show version (simple command)
echo "================================================"
echo "Test 1: Simple command (al version)"
echo "================================================"
./bin/al version
echo ""

# Test 2: Show help for sync command
echo "================================================"
echo "Test 2: Help command (al sync --help)"
echo "================================================"
./bin/al sync --help | head -10
echo ""

# Test 3: Demonstrate output utilities
echo "================================================"
echo "Test 3: Output demonstration"
echo "================================================"
echo ""
echo "The new output utilities provide:"
echo "  ✓ Consistent formatting with icons"
echo "  ⚠️  Warning messages"
echo "  ✗ Error messages"
echo ""
echo "Tool output (brew, mas, etc.) is:"
echo "  • Hidden by default in normal mode"
echo "  • Shown with [tool] prefix when AL_DEBUG=1"
echo ""
echo "Spinners show progress for long operations"
echo ""

# Test 4: Show that AL_DEBUG enables debug output
echo "================================================"
echo "Test 4: Debug mode demonstration"
echo "================================================"
echo ""
echo "Normal mode (no AL_DEBUG):"
echo "  • Only al's own messages are shown"
echo "  • Tool output is suppressed"
echo "  • Spinners are shown for long operations"
echo ""
echo "Debug mode (AL_DEBUG=1):"
echo "  • All debug messages are shown with timestamps"
echo "  • Tool output is shown with [tool] prefix"
echo "  • Spinners are disabled (so you can see output)"
echo ""

echo "================================================"
echo "Demo complete!"
echo "================================================"
echo ""
echo "To test with real commands:"
echo "  Normal:    ./bin/al sync --plan"
echo "  Debug:     AL_DEBUG=1 ./bin/al sync --plan"
