#!/usr/bin/env bash
# =============================================================================
# shokudo-nexus gRPC 契約テスト -- スモークテスト
#
# サービスが契約（proto定義）通りに動作することを検証する。
#
# 前提:
#   - grpcurl がインストール済み
#   - 対象サービスが起動済み
# =============================================================================

set -euo pipefail

SERVICE_HOST="${GRPC_HOST:-localhost}"
SERVICE_PORT="${GRPC_PORT:-9090}"
TARGET="${SERVICE_HOST}:${SERVICE_PORT}"

GRPCURL="grpcurl -plaintext"

PASS=0
FAIL=0

run_test() {
    local name="$1"
    shift
    echo -n "  CONTRACT: ${name} ... "
    if OUTPUT=$("$@" 2>&1); then
        echo "PASS"
        PASS=$((PASS + 1))
    else
        echo "FAIL"
        echo "    Output: ${OUTPUT}"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== 契約テスト: スモーク ==="
echo "Target: ${TARGET}"
echo ""

echo "[FoodInventoryService メソッド確認]"
run_test "FoodInventoryService が公開されている" \
    bash -c "$GRPCURL ${TARGET} list | grep -q 'shokudo.v1.FoodInventoryService'"

run_test "CreateFoodItem メソッドが存在する" \
    bash -c "$GRPCURL ${TARGET} describe shokudo.v1.FoodInventoryService.CreateFoodItem"

run_test "GetFoodItem メソッドが存在する" \
    bash -c "$GRPCURL ${TARGET} describe shokudo.v1.FoodInventoryService.GetFoodItem"

run_test "ListFoodItems メソッドが存在する" \
    bash -c "$GRPCURL ${TARGET} describe shokudo.v1.FoodInventoryService.ListFoodItems"

echo ""
echo "[FusionService メソッド確認]"
run_test "FusionService が公開されている" \
    bash -c "$GRPCURL ${TARGET} list | grep -q 'shokudo.v1.FusionService'"

run_test "CreateFusionRequest メソッドが存在する" \
    bash -c "$GRPCURL ${TARGET} describe shokudo.v1.FusionService.CreateFusionRequest"

echo ""
echo "[エラーコード確認]"
run_test "存在しないIDで NOT_FOUND が返る" \
    bash -c "$GRPCURL -d '{\"id\": \"nonexistent-id\"}' ${TARGET} shokudo.v1.FoodInventoryService/GetFoodItem 2>&1 | grep -q 'NOT_FOUND'"

run_test "空のnameで INVALID_ARGUMENT が返る" \
    bash -c "$GRPCURL -d '{\"name\": \"\"}' ${TARGET} shokudo.v1.FoodInventoryService/CreateFoodItem 2>&1 | grep -q 'INVALID_ARGUMENT'"

echo ""
echo "=== 契約テスト結果 ==="
echo "  PASS: ${PASS}"
echo "  FAIL: ${FAIL}"
echo "  TOTAL: $((PASS + FAIL))"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
