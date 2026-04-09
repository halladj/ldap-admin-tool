#!/bin/bash
# Functional smoke test for ldap-admin-tool
# Run this on the LDAP server after deploying the binary.
# Requires: ldap-admin-tool in PATH,
#           config.yaml readable (or /etc/ldap-admin-tool/config.yaml)

set -euo pipefail

BIN="${1:-ldap-admin-tool}"
TEST_UID="testuser99"
TEST_GROUP="test-group-99"
PASS="0k!gZ2025"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass_count=0
fail_count=0

ok()   { echo -e "${GREEN}[PASS]${NC} $*"; pass_count=$((pass_count + 1)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; fail_count=$((fail_count + 1)); }
info() { echo -e "${YELLOW}[INFO]${NC} $*"; }

run() {
    local desc="$1"; shift
    if output=$("$BIN" "$@" 2>&1); then
        ok "$desc"
        echo "$output" | sed 's/^/       /' | head -8
    else
        fail "$desc"
        echo "$output" | sed 's/^/       /'
    fi
}

run_expect_fail() {
    local desc="$1"; shift
    if output=$("$BIN" "$@" 2>&1); then
        fail "$desc (expected failure, got success)"
        echo "$output" | sed 's/^/       /'
    else
        ok "$desc (correctly rejected)"
    fi
}

echo
echo "================================================"
echo "  ldap-admin-tool functional smoke test"
echo "================================================"
echo

# ── Groups ───────────────────────────────────────────
info "--- Group operations ---"

run "groups create: $TEST_GROUP" \
    groups create "$TEST_GROUP"

run_expect_fail "groups create: duplicate should fail" \
    groups create "$TEST_GROUP"

run "groups query: all groups" \
    groups query

run "groups query: $TEST_GROUP (single)" \
    groups query "$TEST_GROUP"

# ── User create ───────────────────────────────────────
info "--- User create ---"

run "user create: $TEST_UID (PDF + email to halladj00@gmail.com)" \
    user create \
    --first "Test" --last "User" \
    --uid "$TEST_UID" \
    --email "halladj00@gmail.com" \
    --password "$PASS" \
    --groups "$TEST_GROUP"

run_expect_fail "user create: duplicate uid should fail" \
    user create \
    --first "Test" --last "User" \
    --uid "$TEST_UID" \
    --email "halladj00@gmail.com" \
    --no-email --no-pdf

run "user query: all users" \
    user query

run "user query: $TEST_UID (single)" \
    user query --uid "$TEST_UID"

# ── Modify password ───────────────────────────────────
info "--- Modify password ---"

run "user modify password: explicit" \
    user modify password --uid "$TEST_UID" "N3wP@ss2025!"

run "user modify password: auto-generate" \
    user modify password --uid "$TEST_UID"

# ── Modify email ──────────────────────────────────────
info "--- Modify email ---"

run "user modify email" \
    user modify email --uid "$TEST_UID" "changed99@misc-lab.org"

# ── Group membership ──────────────────────────────────
info "--- Group membership ---"

run "user modify remove-group" \
    user modify remove-group --uid "$TEST_UID" "$TEST_GROUP"

run "groups query: $TEST_GROUP (single, should have no members)" \
    groups query "$TEST_GROUP"

run "user modify add-group" \
    user modify add-group --uid "$TEST_UID" "$TEST_GROUP"

run "groups add-users (bulk)" \
    groups add-users "$TEST_GROUP" "$TEST_UID"

run "groups remove-users (bulk)" \
    groups remove-users "$TEST_GROUP" "$TEST_UID"

run "user query: $TEST_UID (single, verify email + group changes)" \
    user query --uid "$TEST_UID"

# ── Error handling ────────────────────────────────────
info "--- Error handling ---"

run_expect_fail "user query: non-existent user (single)" \
    user query --uid "ghost_user_xyz"

run_expect_fail "user modify on non-existent user" \
    user modify password --uid "ghost_user_xyz"

run_expect_fail "groups query: non-existent group (single)" \
    groups query "ghost-group-xyz"

run_expect_fail "groups remove: non-existent group" \
    groups remove "ghost-group-xyz"

# ── Cleanup ───────────────────────────────────────────
info "--- Cleanup ---"

"$BIN" user modify remove-group --uid "$TEST_UID" "$TEST_GROUP" 2>/dev/null || true

run "groups remove: $TEST_GROUP" \
    groups remove "$TEST_GROUP"

run "user delete: $TEST_UID" \
    user delete --uid "$TEST_UID"

run_expect_fail "user query: deleted user should not exist" \
    user query --uid "$TEST_UID"

# ── Summary ───────────────────────────────────────────
echo
echo "================================================"
echo -e "  ${GREEN}PASSED: $pass_count${NC}   ${RED}FAILED: $fail_count${NC}"
echo "================================================"
echo

[ "$fail_count" -eq 0 ]
