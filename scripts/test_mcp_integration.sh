#!/bin/bash
# Test MCP Integration - Kubex Ecosystem
# Tests GoBE MCP hub with Grompt and Analyzer tools

set -e

GOBE_URL="${GOBE_URL:-http://localhost:3666}"
GROMPT_URL="${GROMPT_URL:-http://localhost:8080}"
GROMPT_APIKEY="${GROMPT_APIKEY:-}"

echo "=========================================="
echo "  MCP Integration Test Suite"
echo "=========================================="
echo ""
echo "Configuration:"
echo "  GoBE URL:    $GOBE_URL"
echo "  Grompt URL:  $GROMPT_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run test
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    echo -e "${YELLOW}► Testing:${NC} $test_name"

    if eval "$test_cmd" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAILED${NC}"
        ((TESTS_FAILED++))
    fi
    echo ""
}

# Check if services are running
echo "=== Pre-flight Checks ==="
echo ""

if ! curl -s -f "$GOBE_URL/health" > /dev/null; then
    echo -e "${RED}✗ GoBE not running at $GOBE_URL${NC}"
    echo "  Start with: cd /projects/kubex/gobe && ./gobe start -p 3666"
    exit 1
fi
echo -e "${GREEN}✓ GoBE is running${NC}"

if ! curl -s -f "$GROMPT_URL/api/health" > /dev/null; then
    echo -e "${YELLOW}⚠ Grompt not running at $GROMPT_URL${NC}"
    echo "  Grompt tools will not be available"
else
    echo -e "${GREEN}✓ Grompt is running${NC}"
fi

echo ""
echo "=== Testing Built-in MCP Tools ==="
echo ""

# Test 1: system.status (basic)
run_test "system.status (basic)" \
    "curl -s -X POST $GOBE_URL/mcp/exec \
        -H 'Content-Type: application/json' \
        -d '{\"tool\":\"system.status\",\"args\":{\"detailed\":false}}' \
        | jq -e '.status == \"ok\"'"

# Test 2: system.status (detailed)
run_test "system.status (detailed)" \
    "curl -s -X POST $GOBE_URL/mcp/exec \
        -H 'Content-Type: application/json' \
        -d '{\"tool\":\"system.status\",\"args\":{\"detailed\":true}}' \
        | jq -e '.runtime.go_version'"

# Test 3: shell.command (ls)
run_test "shell.command (ls)" \
    "curl -s -X POST $GOBE_URL/mcp/exec \
        -H 'Content-Type: application/json' \
        -d '{\"tool\":\"shell.command\",\"args\":{\"command\":\"ls\",\"args\":[\"-la\"]}}' \
        | jq -e '.status == \"success\"'"

# Test 4: shell.command (uname)
run_test "shell.command (uname)" \
    "curl -s -X POST $GOBE_URL/mcp/exec \
        -H 'Content-Type: application/json' \
        -d '{\"tool\":\"shell.command\",\"args\":{\"command\":\"uname\",\"args\":[\"-a\"]}}' \
        | jq -e '.output'"

# Test 5: shell.command (blocked command)
run_test "shell.command (security check - rm blocked)" \
    "curl -s -X POST $GOBE_URL/mcp/exec \
        -H 'Content-Type: application/json' \
        -d '{\"tool\":\"shell.command\",\"args\":{\"command\":\"rm\"}}' \
        | jq -e '.status == \"error\"'"

echo "=== Testing External Kubex Tools ==="
echo ""

# Test 6: List all MCP tools
run_test "List all MCP tools" \
    "curl -s -X GET $GOBE_URL/mcp/tools \
        | jq -e 'length > 0'"

# Test 7: Verify grompt.generate is registered
run_test "Verify grompt.generate registered" \
    "curl -s -X GET $GOBE_URL/mcp/tools \
        | jq -e 'map(select(.name == \"grompt.generate\")) | length == 1'"

# Test 8: Verify grompt.direct is registered
run_test "Verify grompt.direct registered" \
    "curl -s -X GET $GOBE_URL/mcp/tools \
        | jq -e 'map(select(.name == \"grompt.direct\")) | length == 1'"

# Test 9: Verify analyzer.project is registered
run_test "Verify analyzer.project registered" \
    "curl -s -X GET $GOBE_URL/mcp/tools \
        | jq -e 'map(select(.name == \"analyzer.project\")) | length == 1'"

# Test 10: Verify analyzer.security is registered
run_test "Verify analyzer.security registered" \
    "curl -s -X GET $GOBE_URL/mcp/tools \
        | jq -e 'map(select(.name == \"analyzer.security\")) | length == 1'"

# Only run live Grompt tests if API key is provided
if [ -n "$GROMPT_APIKEY" ]; then
    echo "=== Testing Grompt Integration (Live API) ==="
    echo ""

    # Test 11: grompt.direct with BYOK
    run_test "grompt.direct (BYOK)" \
        "curl -s -X POST $GOBE_URL/mcp/exec \
            -H 'Content-Type: application/json' \
            -d '{
                \"tool\":\"grompt.direct\",
                \"args\":{
                    \"prompt\":\"Say hello from MCP test\",
                    \"provider\":\"gemini\",
                    \"api_key\":\"'$GROMPT_APIKEY'\",
                    \"max_tokens\":50
                }
            }' | jq -e '.response'"

    # Test 12: grompt.generate with BYOK
    run_test "grompt.generate (BYOK)" \
        "curl -s -X POST $GOBE_URL/mcp/exec \
            -H 'Content-Type: application/json' \
            -d '{
                \"tool\":\"grompt.generate\",
                \"args\":{
                    \"ideas\":[\"test prompt\",\"simple example\"],
                    \"purpose\":\"Testing\",
                    \"provider\":\"gemini\",
                    \"api_key\":\"'$GROMPT_APIKEY'\",
                    \"max_tokens\":100
                }
            }' | jq -e '.response'"
else
    echo -e "${YELLOW}⚠ Skipping live Grompt tests (no API key)${NC}"
    echo "  Set GROMPT_APIKEY to run live integration tests"
    echo ""
fi

# Summary
echo "=========================================="
echo "  Test Results Summary"
echo "=========================================="
echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
echo -e "${RED}Failed:${NC} $TESTS_FAILED"
echo "=========================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
