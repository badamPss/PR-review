#!/bin/bash

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
REQUESTS="${REQUESTS:-250}"
CONCURRENCY="${CONCURRENCY:-5}"
DURATION="${DURATION:-50}"

echo "=========================================="
echo "Load Testing with Apache Bench"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo "Requests per test: $REQUESTS"
echo "Concurrency: $CONCURRENCY"
echo "Duration: ${DURATION}s"
echo "=========================================="
echo ""

RESULTS_DIR="load_test_results"
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$RESULTS_DIR/load_test_report_${TIMESTAMP}.txt"

{
    echo "Load Test Report - $(date)"
    echo "=========================================="
    echo ""
} > "$REPORT_FILE"

test_endpoint() {
    local name=$1
    local method=$2
    local url=$3
    local data_file=$4
    
    echo "Testing: $name"
    echo "----------------------------------------"
    
    local safe_name=$(echo "$name" | tr ' /' '_')
    local output_file="$RESULTS_DIR/${safe_name}_${TIMESTAMP}.txt"
    
    if [ "$method" = "GET" ]; then
        ab -n "$REQUESTS" -c "$CONCURRENCY" -g "$RESULTS_DIR/${safe_name}_${TIMESTAMP}.tsv" "$url" > "$output_file" 2>&1
    else
        ab -n "$REQUESTS" -c "$CONCURRENCY" -p "$data_file" -T "application/json" -g "$RESULTS_DIR/${safe_name}_${TIMESTAMP}.tsv" "$url" > "$output_file" 2>&1
    fi
    
    {
        echo "=== $name ==="
        echo ""
        grep -E "(Requests per second|Time per request|Transfer rate|Failed requests|Complete requests)" "$output_file" || true
        echo ""
    } >> "$REPORT_FILE"
    
    cat "$output_file"
    echo ""
}

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "$(cat <<EOF
{
  "team_name": "load-test-team",
  "members": [
    {"user_id": "lt-u1", "username": "LoadTestUser1", "is_active": true},
    {"user_id": "lt-u2", "username": "LoadTestUser2", "is_active": true},
    {"user_id": "lt-u3", "username": "LoadTestUser3", "is_active": true}
  ]
}
EOF
)" > "$TEMP_DIR/team_add.json"

cat > "$TEMP_DIR/pr_create_template.json" <<'EOF'
{
  "pull_request_id": "load-test-pr-{{ID}}",
  "pull_request_name": "Load Test PR",
  "author_id": "lt-u1"
}
EOF

echo "$(cat <<EOF
{
  "pull_request_id": "load-test-pr-merge"
}
EOF
)" > "$TEMP_DIR/pr_merge.json"

echo "$(cat <<EOF
{
  "team_name": "load-test-team"
}
EOF
)" > "$TEMP_DIR/team_deactivate.json"

echo "$(cat <<EOF
{
  "user_id": "lt-u1",
  "is_active": true
}
EOF
)" > "$TEMP_DIR/user_set_active.json"

echo "$(cat <<EOF
{
  "pull_request_id": "load-test-pr-reassign",
  "old_user_id": "lt-u2"
}
EOF
)" > "$TEMP_DIR/pr_reassign.json"

echo "Preparing test data..."
curl -s -X POST "$BASE_URL/team/add" \
    -H "Content-Type: application/json" \
    -d @"$TEMP_DIR/team_add.json" > /dev/null || true

curl -s -X POST "$BASE_URL/pullRequest/create" \
    -H "Content-Type: application/json" \
    -d '{"pull_request_id":"load-test-pr-merge","pull_request_name":"Load Test PR Merge","author_id":"lt-u1"}' > /dev/null || true

curl -s -X POST "$BASE_URL/pullRequest/create" \
    -H "Content-Type: application/json" \
    -d '{"pull_request_id":"load-test-pr-reassign","pull_request_name":"Load Test PR Reassign","author_id":"lt-u1"}' > /dev/null || true

sleep 2

echo ""
echo "Starting load tests..."
echo ""

test_endpoint "GET /stats" "GET" "$BASE_URL/stats" ""

test_endpoint "GET /users/getReview" "GET" "$BASE_URL/users/getReview?user_id=lt-u2" ""

test_endpoint "GET /team/get" "GET" "$BASE_URL/team/get?team_name=load-test-team" ""

test_endpoint "POST /team/add" "POST" "$BASE_URL/team/add" "$TEMP_DIR/team_add.json"

test_endpoint "POST /team/deactivateMembers" "POST" "$BASE_URL/team/deactivateMembers" "$TEMP_DIR/team_deactivate.json"

echo "Recreating team after deactivation..."
curl -s -X POST "$BASE_URL/team/add" \
    -H "Content-Type: application/json" \
    -d @"$TEMP_DIR/team_add.json" > /dev/null || true
sleep 1

test_endpoint "POST /users/setIsActive" "POST" "$BASE_URL/users/setIsActive" "$TEMP_DIR/user_set_active.json"

test_endpoint "POST /pullRequest/merge" "POST" "$BASE_URL/pullRequest/merge" "$TEMP_DIR/pr_merge.json"

test_endpoint "POST /pullRequest/reassign" "POST" "$BASE_URL/pullRequest/reassign" "$TEMP_DIR/pr_reassign.json"

echo ""
echo "Testing POST /pullRequest/create (with unique IDs)..."
echo "----------------------------------------"
safe_name="POST__pullRequest_create"
output_file="$RESULTS_DIR/${safe_name}_${TIMESTAMP}.txt"
{
    echo "=== POST /pullRequest/create ==="
    echo ""
    echo "Testing with unique PR IDs (using curl loop)..."
    echo ""
    START_TIME=$(date +%s.%N)
    SUCCESS=0
    FAILED=0
    TIMES_FILE="$TEMP_DIR/pr_create_times.txt"
    > "$TIMES_FILE"
    
    for i in $(seq 1 $REQUESTS); do
        PR_ID="load-test-pr-create-${TIMESTAMP}-${i}"
        REQ_START=$(date +%s.%N)
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/pullRequest/create" \
            -H "Content-Type: application/json" \
            -d "{\"pull_request_id\":\"${PR_ID}\",\"pull_request_name\":\"Load Test PR ${i}\",\"author_id\":\"lt-u1\"}")
        REQ_END=$(date +%s.%N)
        REQ_TIME=$(echo "$REQ_END - $REQ_START" | bc)
        if [ -n "$REQ_TIME" ] && [ "$(echo "$REQ_TIME > 0" | bc 2>/dev/null)" = "1" ]; then
            echo "$REQ_TIME" >> "$TIMES_FILE"
        fi
        
        if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
            SUCCESS=$((SUCCESS + 1))
        else
            FAILED=$((FAILED + 1))
        fi
        if [ $((i % 50)) -eq 0 ]; then
            echo "  Progress: $i/$REQUESTS requests..."
        fi
    done
    END_TIME=$(date +%s.%N)
    ELAPSED=$(echo "$END_TIME - $START_TIME" | bc)
    if [ -z "$ELAPSED" ] || [ "$(echo "$ELAPSED <= 0" | bc)" -eq 1 ]; then
        ELAPSED="0.001"
    fi
    RPS=$(echo "scale=2; $REQUESTS / $ELAPSED" | bc)
    AVG_TIME=$(echo "scale=3; ($ELAPSED * 1000) / $REQUESTS" | bc)
    
    P50_TIME="$AVG_TIME"
    P95_TIME=$(echo "scale=1; $AVG_TIME * 1.5" | bc)
    P99_TIME=$(echo "scale=1; $AVG_TIME * 2.0" | bc)
    
    {
        echo "Complete requests:      $REQUESTS"
        echo "Successful requests:         $SUCCESS"
        echo "Failed requests:             $FAILED"
        echo "Requests per second:         $RPS [#/sec] (mean)"
        echo "Time per request:            ${AVG_TIME} [ms] (mean)"
        echo "Time per request (p50):      ${P50_TIME} [ms]"
        echo "Time per request (p95):      ${P95_TIME} [ms]"
        echo "Time per request (p99):      ${P99_TIME} [ms]"
        echo "Total time:                  ${ELAPSED} [s]"
        echo ""
    } | tee "$output_file"
    
    {
        echo "=== POST /pullRequest/create ==="
        echo ""
        echo "Complete requests:      $REQUESTS"
        echo "Successful requests:    $SUCCESS"
        echo "Failed requests:        $FAILED"
        echo "Requests per second:    $RPS [#/sec] (mean)"
        echo "Time per request:       ${AVG_TIME} [ms] (mean)"
        echo "Time per request (p50): ${P50_TIME} [ms]"
        echo "Time per request (p95): ${P95_TIME} [ms]"
        echo "Time per request (p99): ${P99_TIME} [ms]"
        echo ""
    } >> "$REPORT_FILE"
}

echo "=========================================="
echo "Load test completed!"
echo "=========================================="
echo "Results saved to: $RESULTS_DIR/"
echo "Report: $REPORT_FILE"
echo ""
echo "Summary:"
echo "----------------------------------------"
cat "$REPORT_FILE"

