#!/usr/bin/env bash
# Benchmark: measure peak RSS and detection duration for subs-check
set -euo pipefail

CONFIG="config/bench.yaml"
TMPDIR="${TMPDIR:-/home/bubu/subs-check/tmp-build}"
export TMPDIR

if [ ! -f ./subs-check ]; then
    echo "Building..."
    go build -o subs-check . 2>/dev/null || { echo "Build failed"; exit 1; }
fi

mkdir -p output tmp
rm -f tmp/bench.log

# Kill any lingering instance (match only actual binary)
for pid in $(pgrep -f "^./subs-check" 2>/dev/null || true); do
    kill $pid 2>/dev/null || true
done
sleep 1

export SUB_CHECK_MEM_MONITOR=1
export SUB_CHECK_MEM_INTERVAL=2s

./subs-check -f "$CONFIG" > tmp/bench.log 2>&1 &
APP_PID=$!

TIMEOUT=120
START=$(date +%s)
PEAK_RSS_KB=0
DURATION=0

poll_rss() {
    local pid=$1
    if [ -f /proc/$pid/status ]; then
        awk '/VmRSS:/ {print $2}' /proc/$pid/status 2>/dev/null || echo 0
    else
        echo 0
    fi
}

while [ $(( $(date +%s) - START )) -lt $TIMEOUT ]; do
    RSS=$(poll_rss $APP_PID)
    if [ "$RSS" -gt "$PEAK_RSS_KB" ]; then
        PEAK_RSS_KB=$RSS
    fi
    if grep -q "检测完成" tmp/bench.log 2>/dev/null; then
        DURATION=$(( $(date +%s) - START ))
        break
    fi
    sleep 0.5
done

kill $APP_PID 2>/dev/null || true
wait $APP_PID 2>/dev/null || true

if [ "$DURATION" -eq 0 ]; then
    echo "METRIC peak_rss_mb=9999"
    echo "METRIC detect_duration_s=9999"
    echo "TIMEOUT"
    exit 1
fi

# Strip ANSI color codes helper
strip_ansi() { sed 's/\x1b\[[0-9;]*m//g'; }

# Parse peak HeapAlloc from log
HEAP=$(grep "内存使用情况" tmp/bench.log 2>/dev/null | strip_ansi | \
    sed -n 's/.*HeapAlloc="\([0-9.]*\).*/\1/p' | sort -n | tail -1)
[ -z "$HEAP" ] && HEAP=0

# Parse peak HeapInuse
INUSE=$(grep "内存使用情况" tmp/bench.log 2>/dev/null | strip_ansi | \
    sed -n 's/.*HeapInuse="\([0-9.]*\).*/\1/p' | sort -n | tail -1)
[ -z "$INUSE" ] && INUSE=0

# Parse peak Sys
SYS=$(grep "内存使用情况" tmp/bench.log 2>/dev/null | strip_ansi | \
    sed -n 's/.*Sys="\([0-9.]*\).*/\1/p' | sort -n | tail -1)
[ -z "$SYS" ] && SYS=0

PEAK_RSS_MB=$(awk "BEGIN{print $PEAK_RSS_KB/1024}")

printf "METRIC peak_rss_mb=%.2f\n" "$PEAK_RSS_MB"
printf "METRIC detect_duration_s=%d\n" "$DURATION"
printf "METRIC peak_heap_alloc_mb=%.2f\n" "$HEAP"
printf "METRIC peak_heap_inuse_mb=%.2f\n" "$INUSE"
printf "METRIC peak_sys_mb=%.2f\n" "$SYS"

echo ""
echo "--- Summary ---"
printf "Peak RSS:      %.2f MB\n" "$PEAK_RSS_MB"
printf "Duration:      %d s\n" "$DURATION"
printf "Peak Heap:     %.2f MB\n" "$HEAP"
printf "Peak Inuse:    %.2f MB\n" "$INUSE"
printf "Peak Sys:      %.2f MB\n" "$SYS"
grep -E "节点解析|可用节点|测试总消耗" tmp/bench.log 2>/dev/null | tail -4
