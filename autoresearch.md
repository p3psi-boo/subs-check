# Autoresearch Rules — subs-check Resource Optimization

## Primary Metric
- **Name**: `peak_rss_mb` — Peak RSS memory during a full detection cycle (subscription fetch + alive check + speed test + media check)
- **Unit**: MB
- **Direction**: lower is better

## Secondary Metrics
- `detect_duration_s`: Total wall-clock time for one detection cycle (lower is better)
- `peak_heap_alloc_mb`: Peak Go heap allocation during cycle (lower is better)
- `yaml_alloc_mb`: Cumulative YAML parser allocations (lower is better)
- `total_goroutines`: Peak goroutine count during detection (lower is better)

## Benchmark Protocol
1. Use `config/config.yaml` with the full 50+ subscription list (realistic load)
2. Enable `speed-test-url` + `media-check` to exercise all code paths
3. Use `check-interval: 720` to prevent automatic re-runs interfering with measurement
4. Run `SUB_CHECK_MEM_MONITOR=1` to emit memory stats to logs
5. Parse memory stats from log output to extract peak values
6. Run `SUB_CHECK_PPROF=1` and collect heap profile at end of detection for alloc_space
7. Run 3 times and report median to reduce noise

## Constraints
- **DO NOT** modify `github.com/metacubex/mihomo` or any upstream dependency
- **DO NOT** modify subscription URLs or network behavior to fake faster results
- **DO NOT** reduce detection accuracy (e.g., skipping platforms, lowering timeout) to win metric
- Only optimize within this repo: `proxy/`, `check/`, `app/`, `save/`, `utils/`, `config/`

## Valid Optimizations
- Memory allocation reduction (buffer reuse, pre-allocation, object pooling)
- Goroutine count reduction (worker pool tuning, connection sharing)
- GC pressure reduction (fewer temporary allocations, reference cutting)
- I/O batching (progress display throttling, log batching)
- Connection pool tuning (idle timeout, max conns)
- Regex/string operation efficiency
- Map/slice capacity planning

## Invalid "Optimizations" (Cheating)
- Disabling media-check or speed-test to reduce work
- Reducing timeout so nodes fail faster
- Removing subscriptions to reduce node count
- Hiding memory leaks by calling FreeOSMemory more often without fixing root cause
