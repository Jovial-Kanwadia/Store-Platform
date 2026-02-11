#!/usr/bin/env bash
# e2e.sh — End-to-end smoke test for the store platform.
# Usage: API_URL=http://localhost:8080 ./test/e2e.sh
set -euo pipefail

API="${API_URL:-http://localhost:8080}"
STORE_NAME="e2e-test-$(date +%s)"
TIMEOUT=200
POLL_INTERVAL=5

cleanup() {
  echo "--- Cleanup: deleting store ${STORE_NAME} ---"
  curl -sf -X DELETE "${API}/api/v1/stores/${STORE_NAME}" > /dev/null 2>&1 || true
}
trap cleanup EXIT

echo "=== E2E Test Start ==="

# 1. Create Store
echo "--- Creating store: ${STORE_NAME} ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${API}/api/v1/stores" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${STORE_NAME}\",\"engine\":\"woo\",\"plan\":\"small\"}")

if [ "$HTTP_CODE" != "201" ]; then
  echo "FAIL: Expected 201, got ${HTTP_CODE}"
  exit 1
fi
echo "OK: Store created (201)"

# 2. Poll until Ready
echo "--- Waiting for store to become Ready (timeout: ${TIMEOUT}s) ---"
ELAPSED=0
while [ $ELAPSED -lt $TIMEOUT ]; do
  STATUS=$(curl -sf "${API}/api/v1/stores/${STORE_NAME}" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)

  if [ "$STATUS" = "Ready" ]; then
    echo "OK: Store is Ready after ${ELAPSED}s"
    break
  fi

  echo "  status=${STATUS:-unknown}, waiting ${POLL_INTERVAL}s..."
  sleep $POLL_INTERVAL
  ELAPSED=$((ELAPSED + POLL_INTERVAL))
done

if [ $ELAPSED -ge $TIMEOUT ]; then
  echo "FAIL: Store did not become Ready within ${TIMEOUT}s (last status: ${STATUS:-unknown})"
  exit 1
fi

# 3. Delete Store
echo "--- Deleting store: ${STORE_NAME} ---"
# Disable the trap cleanup since we're doing it manually
trap - EXIT

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "${API}/api/v1/stores/${STORE_NAME}")
if [ "$HTTP_CODE" != "204" ] && [ "$HTTP_CODE" != "200" ]; then
  echo "FAIL: Expected 204, got ${HTTP_CODE}"
  exit 1
fi
echo "OK: Store deleted"

# 4. Verify Gone (poll for 404)
echo "--- Verifying store is gone ---"
ELAPSED=0
while [ $ELAPSED -lt 60 ]; do
  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${API}/api/v1/stores/${STORE_NAME}")
  if [ "$HTTP_CODE" = "404" ]; then
    echo "OK: Store returns 404"
    break
  fi
  sleep 3
  ELAPSED=$((ELAPSED + 3))
done

if [ "$HTTP_CODE" != "404" ]; then
  echo "FAIL: Store still exists after deletion (got ${HTTP_CODE})"
  exit 1
fi

echo "=== E2E Test Passed ==="
