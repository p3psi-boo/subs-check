# Deferred / Abandoned Optimization Ideas — subs-check

## Successful Optimizations (Committed)

1. **Regex pre-compilation** — `mediaTagRegex` + `datePlaceholderRegexes` package-level vars
2. **Reduced `debug.FreeOSMemory()` calls** — removed per-200K-node async GC in `distributeJobs`
3. **Pre-allocated `results` slice** in `NewProxyChecker` based on `SuccessLimit` or `proxyCount/4`
4. **HTTP connection pool tuning** — `MaxIdleConns` 100→50, `IdleConnTimeout` 90s→15s (fetch), 5s→2s (check)
5. **Progress display adaptive interval** — 100ms→1s for ≥10K nodes
6. **Explicit GC before `GetProxies` returns** — `runtime.GC()` + `debug.FreeOSMemory()` after YAML parsing
7. **Temp log directory moved to `./tmp`** — avoids system `/tmp` pollution
8. **Tags slice pre-allocation** in `updateProxyName` — `make([]string, 0, len(platforms)+3)`
9. **Progress bar stack buffer** — `[41]byte` replaces `strings.Repeat` heap allocation
10. **`fmt.Sprintf` → `strconv` / string concat / structured slog** — Replaces per-node string formatting with faster zero-allocation patterns
11. **mediaChan buffer 2x→1x** — Reduces queued ProxyJobs after speed stage; safe because media stage processes fewer jobs than speed stage
12. **Benchmark RSS sampling 0.5s→0.2s** — More accurate peak capture for fast detection runs
13. **Conditional geoDB loading** — Skip 5-10MB MaxMind DB when `rename-node=false` and no `iprisk` platform
14. **Single-goroutine `distributeJobs` for ≤1000 nodes** — Eliminates worker pool overhead for small workloads; preserves pool for >1000
15. **Package-level regex pre-compilation** — `mediaTagRegex` in `check.go` and `tiktokRegionRegex` in `tiktok.go`; removes per-node compilation hotspots
16. **GOGC tuning (SetGCPercent 10) during `check.Check()`** — Peak RSS ~46-47MB vs ~57-58MB baseline (-18%). Peak heap cut from ~17MB to ~5.5MB (-51%). Triggered by detection-phase massive temporary allocations.
17. **Lazy ProxyClient creation** — Move `CreateClient` from `distributeJobs` to `runAliveStage` workers. aliveChan buffer now holds lightweight jobs (map only) instead of heavyweight mihomo proxies. Peak concurrent proxies cut from ~110 to ~50. Combined with GOGC=10: peak RSS ~46-51MB, best 46.14 MB (-19.5% from baseline). **24.3× confidence.**
18. **Dedup map capacity 2048→512 per sub** — Less empty map overhead during subscription parsing, faster GC scans. Best 46.29 MB. **21.3× confidence.**
19. **Structured slog in `CreateClient` hot path** — Replace `fmt.Sprintf` with `slog.Debug("msg", "key", value)` to avoid string formatting for disabled log levels. Code quality win.
20. **Disable idle connection pool in check Transport** — `IdleConnTimeout` 2s→0, `MaxIdleConnsPerHost` 2→0. Prevents failed proxy connections from accumulating in the Transport idle pool. Combined with all previous: peak RSS ~41.8 MB vs ~57.5MB baseline (-27%). **22.5× confidence.**
21. **`checkCtxDone` select→c.Err()** — Replace `select { case <-c.Done(): }` with direct `c.Err() != nil` check. Avoids per-node select overhead. Micro-optimization.
22. **`needsCF()` cache in ProxyChecker** — Cache `needsCF(config.GlobalConfig.Platforms)` in `ProxyChecker` struct to avoid per-node platform loop. Code quality win.

## Attempted & Reverted (No Benefit or Worse)

| Idea | Result | Reason |
|------|--------|--------|
| Halve `distributeJobs` concurrency (50→16) | Duration +17%, RSS flat | Distribution is not the bottleneck |
| Clash YAML direct struct unmarshal shortcut | RSS +10%, Heap +41% | goccy/go-yaml overhead unchanged; struct approach added complexity |
| Plain-text link shortcut before YAML parse | RSS +1% | Bench subscriptions are YAML-dominant |
| Disable HTTP Keep-Alive | RSS +5% | Forces TLS re-handshake per node, more concurrent connections |
| ForceClose ticker 100ms→500ms + remove per-node slog.Debug | RSS +4% | These allocations are negligible compared to mihomo |
| Map pre-allocation in `mergeUniqueProxies` / `checkSubscriptionSuccessRate` | RSS +8% | Measurement noise; effect below threshold |
| Reduce `GetProxies` map hint (2048→256 per sub) | RSS +1%, Heap +15% | Hint too low causes rehash allocations during growth |
| `StatsTransport.CloseIdleConnections()` proxy method | N/A (reverted with group) | Clean but no measurable resource impact |
| Remove `Connection: close` from google.go + cloudflare.go | RSS +5-8% | Connection reuse increases idle connection retention; no benefit for failed nodes |
| Reduce `GetProxies` map hint (2048→1024 per sub) | RSS +1%, Heap +15% | Rehash cost still exceeds bucket savings at this scale |
| `fmt.Sprintf` → `strconv` in updateProxyName + structured slog in check.go | RSS stable | Micro-optimization; code quality improvement with minor allocation reduction |
| `fmt.Sprintf` → string concat for YT/TK tags | RSS stable | Removes per-node fmt.Sprintf for simple prefix patterns |
| **mediaChan buffer 2x→1x** | RSS ~-0.5MB (confirmed with 3.0× confidence) | Clean reduction; mediaChan rarely fills with failed nodes |
| aliveChan buffer 1.2x→1x | RSS mixed (58.59 vs 57.37 control) | 1x causes more goroutine blocking, increasing stack memory; **1.2x is sweet spot** |
| aliveChan buffer 1.2x→aliveConc/2 | RSS +4MB (61.67) | Too small; goroutine blocking overhead exceeds buffer savings |
| `nodes[i]=nil` in `processSubscription` after send | RSS +3MB (61.58) | Adding nil assignments per-iteration adds overhead without helping GC |
| Benchmark RSS sampling 0.5s→0.2s | Reveals true peak is 1-3MB higher | 0.5s interval missed peaks for fast runs; **0.2s is more accurate** |
| **Conditional geoDB loading** | Saves 5-10MB when rename-node=false & no iprisk | For users who don't need geolocation; no impact when geoDB is needed |
| **Single-goroutine distributeJobs ≤1000 nodes** | RSS stable, heap inuse slightly lower | Eliminates 49 goroutine stacks, atomic contention, WaitGroup overhead; preserves pool for >1000 |
| **mediaTagRegex package-level pre-compile** | Fixes per-node regex compilation in `updateProxyName` | Removes CPU and alloc hotspot for successful nodes; benchmark doesn't exercise it |
| **tiktokRegionRegex package-level pre-compile** | Fixes per-node regex compilation in `CheckTikTok` | Same as above; media checks only run on successful nodes |
| Map capacity 2048→1536 per sub | Within noise (54.51-60.22 MB) | **Revert**: 2048 is near optimal; lower causes rehash overhead for large subscriptions |
| aliveChan 1.2x→1x with single-goroutine distributeJobs | RSS +1.5MB (59.53) | **Revert**: 1.2x remains sweet spot even with single-goroutine distribution |
| GOGC env 50 (whole program) | ~52 MB peak RSS | Code-level equivalent: `debug.SetGCPercent(50)` in `check.Check()` |
| GOGC 50→25 in `check.Check()` | ~48 MB peak RSS | Further improvement; diminishing returns |
| GOGC 25→10 in `check.Check()` | ~46-47 MB peak RSS | **Keep**: best result, stable across 3+ runs, no duration regression |
| GOGC 10→5 in `check.Check()` | ~45-48 MB peak RSS, duration ~29s | **Revert**: marginal RSS improvement, duration regression on slow runs |
| aliveChan 1.2x→1x with lazy creation | RSS +2MB (48.88) | **Revert**: 1.2x remains sweet spot even with lightweight buffered jobs |
| aliveChan 1.2x→aliveConc/2 with lazy creation | Within noise (46.34-52.17) | **Revert**: no clear improvement, keep 1.2x for stability |
| Periodic `FreeOSMemory()` during detection | RSS +5MB (51.80) | **Revert**: adds CPU overhead, no benefit when GOGC=10 keeps heap tight |
| `MaxIdleConnsPerHost` 2→0, `IdleConnTimeout` 2s→0 | **Peak RSS ~41.8 MB** | **Keep**: eliminates idle connection pool accumulation across failed checks |
| aliveChan 1.2x→2.0x with lazy creation | RSS +2MB (48.00) | **Revert**: no benefit; 1.2x is optimal |
| `checkCtxDone` select→c.Err() | RSS stable | **Keep**: avoids select overhead per-node; micro-optimization |
| `needsCF()` cache in ProxyChecker | RSS stable | **Keep**: removes per-node platform loop; code quality win |
| **Lazy ProxyClient creation** | Peak RSS ~46-51 MB (best 46.14) | **Keep**: aliveChan buffer no longer holds heavy mihomo proxies. Peak concurrent proxies cut from ~110 to ~50. 24.3× confidence |
| Map capacity 2048→512 per sub | Peak RSS ~46-47 MB (best 46.29) | **Keep**: less empty map overhead, faster GC scans. 21.3× confidence |
| `fmt.Sprintf` → structured slog in `CreateClient` | RSS stable | **Keep**: removes per-node string formatting in hot path. Code quality win |

## Root-Cause Analysis (From Real Flamegraphs)

### CPU Hotspots
- **TLS/X509 certificate verification: 32.54%** — `crypto/tls.(*Conn).HandshakeContext`, `crypto/x509.(*Certificate).buildChains`
- **GC mark/sweep: 10.53%** — `runtime.gcDrain`, `runtime.scanObject`
- **Scheduler: 9.57%** — `runtime.findRunnable`, `runtime.stealWork`

### Block Hotspots
- **Channel/select wait: 86.37%** — `runtime.chanrecv2`, `runtime.selectgo`
- **WaitGroup synchronization: 41.30%** — `sync.(*WaitGroup).Go.func1`
- **HTTP connection wait: 20.08%** — `net/http.(*Client).Do`
- **TLS handshake wait: 10.36%** — `net/http.(*persistConn).addTLS`

### Memory Allocation Hotspots
- **YAML parser: 76.34% cumulative** — `github.com/goccy/go-yaml/scanner.(*Context).addBuf`
- **V2Ray conversion: 7.59%** — `github.com/metacubex/mihomo/common/convert.ConvertsV2Ray`
- **IO.ReadAll fallback: 3.23%** — `io.ReadAll` (subscription download)

### Goroutine Scale
- **351 goroutines** during peak detection (131 from subs-check code, ~220 from mihomo internals)
- Each mihomo proxy instance spawns 2-3 internal goroutines for connection management

## Why Further Optimization Is Blocked

1. **TLS/X509 32% CPU** — This is intrinsic to proxy detection. Each node must establish a TLS tunnel through mihomo. No way to batch or share TLS handshakes across independent proxies.
2. **YAML parser 76% alloc** — `goccy/go-yaml` is a third-party library. Replacing it with a zero-allocation parser is a massive undertaking and risks breaking subscription parsing for edge-case formats.
3. **351 goroutines** — ~63% come from mihomo's internal connection management (WebSocket framing, QUIC streams, TLS state machines). These are outside project scope.
4. **Channel blocking 86%** — This is natural backpressure in a producer-consumer pipeline. Alive workers block sending to speedChan; speed workers block on download I/O. Not a bug.

## Potential Future Work (Outside Current Scope / Deferred)

- **Replace goccy/go-yaml with a streaming YAML parser** for subscription parsing to eliminate 76% of heap allocations. High effort, high risk. Requires extensive testing across all subscription formats.
- **Share `http.Transport` base instances** across nodes that use the same underlying server, with per-node `DialContext` dispatch. Complex due to `StatsTransport` per-node isolation and mihomo proxy lifecycle.

## Tried and Ruled Out

- **Add a DNS cache layer** — DNS resolution happens inside mihomo's `DialContext`, not subs-check's custom `DialContext`. Cannot implement without modifying mihomo.
- **Object pooling for `strings.Builder` / `ProxyNode` maps** — `GenerateProxyKey` runs during subscription fetch (pre-peak), pooling would save <1MB.
- **Profile mihomo upstream** — Outside project scope per autoresearch.md constraints.
- **Connection: close removal** — Increases idle connection retention; only benefits successful nodes, not measurable in current benchmark.
- **Map hint reduction (2048→1024/256)** — Rehash cost outweighs bucket savings at tested scale.
- **distributeJobs concurrency reduction** — Not the bottleneck; increases duration.
