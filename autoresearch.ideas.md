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

## Potential Future Work (Outside Current Scope)

- **Replace goccy/go-yaml with a streaming YAML parser** for subscription parsing to eliminate 76% of heap allocations. High effort, high risk.
- **Add a DNS cache layer** in `DialContext` to reduce repeated DNS lookups for common endpoints (google.com, cloudflare.com, etc.). Moderate effort, moderate risk.
- **Share `http.Transport` base instances** across nodes that use the same underlying server, with per-node `DialContext` dispatch. Complex due to `StatsTransport` per-node isolation.
- **Implement object pooling for `strings.Builder` in `GenerateProxyKey`** and `ProxyNode` maps. Low effort, but marginal gain (<1MB).
- **Profile mihomo upstream** and contribute patches to reduce goroutine spawning per proxy connection. Requires deep mihomo expertise.
