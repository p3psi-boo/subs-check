# Subs-Check

Subs-Check 是一个代理订阅检测与筛选工具。本仓库是基于上游项目 [sinspired/subs-check](https://github.com/sinspired/subs-check) 的 fork，当前维护方向是收窄功能边界、降低运行面复杂度，并保留订阅检测流水线的核心能力。

当前版本的定位很明确：拉取订阅、解析节点、并发检测、输出可用节点文件，并提供一个轻量 Web 控制面板用于配置和运行管理。

这份 README 描述的是当前 fork 的代码状态，不代表上游项目的完整功能集，也不再沿用历史文档中的功能承诺。已经移除的能力包括自动更新、通知推送、Sub-Store 前后端、云端保存后端、多平台内置 mihomo 资源和 Windows/macOS 构建目标。

## 当前定位

这个项目适合下面几类场景：

- 定时从多个订阅源拉取节点，过滤出可用节点。
- 在受控机器上运行一个本地订阅整理服务。
- 对节点做测活、测速、媒体解锁和基础质量标记。
- 生成稳定的本地输出文件，交给 Mihomo/Clash 或其他下游工具消费。
- 通过 WebUI 管理配置、触发检测、查看状态和日志。

当前版本不试图承担完整订阅转换平台、通知中心或云同步系统的职责。这些能力被移出后，运行面更小，部署和排查成本也更低。

## 功能范围

### 保留能力

- 订阅拉取：支持本地 `sub-urls` 和远程订阅清单 `sub-urls-remote`。
- 订阅格式：支持 Clash/Mihomo、V2Ray/Base64 等常见节点格式，并兼容部分非标准格式。
- 远程清单：支持纯文本、YAML/JSON 数组、`sub-urls` 对象，以及 Mihomo `proxy-providers.*.url`。
- 时间占位符：订阅 URL 支持 `{Ymd}`、`{Y-m-d}`、`{Y}`、`{m}`、`{d}` 等日期占位符。
- GitHub 代理：可配置 `github-proxy` 或 `ghproxy-group`，用于拉取 GitHub 上的订阅清单。
- 流水线检测：测活、测速、媒体检测分阶段并发执行。
- 测速保护：支持单节点下载时长、下载大小和总测速下载速率限制。
- 媒体检测：支持 OpenAI、Gemini、YouTube、Netflix、Disney、TikTok、X、IP risk 等平台配置。
- 节点标记：支持位置重命名、增强位置标签、ISP 类型检测、Cloudflare 访问质量过滤。
- 历史保留：可加载上次成功节点和历史成功节点，降低上游订阅波动造成的可用节点丢失。
- 本地保存：检测结果保存到本地 `output` 目录或 `output-dir` 指定目录。
- WebUI：提供配置编辑、状态查看、手动触发检测、强制停止检测和日志查看。
- HTTP 文件服务：提供订阅文件访问入口，便于下游客户端订阅。
- 回调脚本：检测完成后可执行本地脚本做后处理。
- Nix 开发环境：提供 `flake.nix` 和 `.envrc`。

### 已移除能力

- 自动检查更新和自更新。
- Telegram、DingTalk、Apprise 等通知推送。
- Sub-Store 前端、后端和分享转换能力。
- Cloudflare R2、Gist、MinIO/S3、WebDAV 等远程保存后端。
- 内置多平台 mihomo 二进制资源。
- Windows/macOS 构建目标。当前构建目标以 Linux 为准。

## 快速运行

### 源码运行

要求 Go 工具链可用。当前项目按 Linux 环境维护。

```bash
go run . -f ./config/config.yaml
```

首次运行会生成或使用配置文件。建议先从示例配置开始：

```bash
cp config/config.yaml.example config/config.yaml
go run . -f ./config/config.yaml
```

### 构建二进制

```bash
make build
./subs-check -f ./config/config.yaml
```

构建 Linux 平台包：

```bash
make build-all
```

### Nix 开发环境

```bash
nix develop
make build
```

项目的 `.envrc` 会配合 direnv 使用，并将 GOPATH 指向项目内的 `.gopath`。`.direnv/` 和 `.gopath/` 已加入 `.gitignore`。

## 配置文件

完整配置以 [config/config.yaml.example](config/config.yaml.example) 为准。下面只列核心项。

### 调度

| 配置 | 作用 | 建议 |
| --- | --- | --- |
| `check-interval` | 按分钟周期检测 | 简单部署使用 |
| `cron-expression` | 标准 cron 表达式 | 生产环境优先使用 |
| `print-progress` | 终端显示进度 | 批处理环境可关闭 |
| `progress-mode` | 进度显示模式 | 默认 `auto` |

如果设置了 `cron-expression`，会优先使用 cron 调度。当前逻辑下，使用 cron 时首次启动不会立即检测。

### 并发与 IO

| 配置 | 作用 | 建议 |
| --- | --- | --- |
| `concurrent` | 默认并发基准，也影响订阅获取并发 | 常规 20 到 100 |
| `alive-concurrent` | 测活并发 | 100 到 1000，视 CPU 和路由器能力调整 |
| `speed-concurrent` | 测速并发 | 4 到 32，视出口带宽调整 |
| `media-concurrent` | 媒体检测并发 | 20 到 200 |
| `timeout` | 单节点检测超时，毫秒 | 网络差时适当增大 |
| `total-speed-limit` | 总测速下载速率限制，MB/s | 共享限制，只作用于测速下载 |
| `download-timeout` | 单节点测速最长时间 | 建议保留 |
| `download-mb` | 单节点测速最大下载量 | 建议保留 |

当前流水线主要受 IO 影响。测速阶段会主动消耗出口带宽；如果 `speed-concurrent` 偏高，应该同时设置 `total-speed-limit`，避免把本机或上游网络打满。

最近的 IO 调整有三点：

- `total-speed-limit` 只限制测速下载 reader，不再限制所有 HTTP 连接。
- `sub-urls-remote` 使用有界并发拉取，最多并发 8 个远程清单。
- 订阅获取阶段的 channel 和去重 map 容量改为按订阅规模和并发动态计算。

### 订阅输入

```yaml
sub-urls-remote:
  - "https://example.com/sub-list.txt"
  - "https://example.com/sub-list.yaml"

sub-urls:
  - "https://example.com/sub.yaml#source-a"
  - "https://raw.githubusercontent.com/example/repo/main/{Ymd}.yaml"
```

`sub-urls-remote` 用于集中维护订阅清单。它可以是：

- 纯文本，每行一个 URL，支持空行和 `#` 注释。
- YAML/JSON 字符串数组。
- 包含 `sub-urls` 字段的对象。
- Mihomo 配置中的 `proxy-providers.*.url`。

订阅 URL 的 fragment 会作为来源备注参与节点命名，例如 `#source-a`。

### 节点处理

| 配置 | 作用 |
| --- | --- |
| `rename-node` | 根据出口位置重命名节点 |
| `node-prefix` | 为节点名增加统一前缀 |
| `node-type` | 只检测指定协议，例如 `ss`、`vmess`、`vless`、`trojan` |
| `threshold` | 智能乱序相似度阈值，降低相邻同网段节点被连续测速的概率 |
| `keep-success-proxies` | 保留并加载上次成功和历史成功节点 |
| `success-limit` | 限制输出成功节点数量，0 表示不限制 |

`keep-success-proxies` 对稳定性有实际价值。订阅源短期失效或上游文件更新时，历史成功节点可以继续进入下一轮检测。

### 媒体与质量检测

```yaml
media-check: true
platforms:
  - iprisk
  - openai
  - gemini
  - youtube
  - x
```

可选平台包括 `iprisk`、`openai`、`gemini`、`youtube`、`tiktok`、`netflix`、`disney`、`x`。

相关配置：

| 配置 | 作用 |
| --- | --- |
| `drop-bad-cf-nodes` | 丢弃无法访问 Cloudflare 的节点 |
| `isp-check` | 查询 ISP 类型，例如原生、广播、住宅、机房 |
| `enhanced-tag` | 使用增强位置标签表达出口位置和 CDN 识别结果 |
| `maxmind-db-path` | 指定本地 MaxMind 数据库路径 |

`drop-bad-cf-nodes` 可能造成可用节点数量明显下降，也可能误杀只是不适合访问 Cloudflare 的节点。除非下游业务明确依赖 Cloudflare，否则不建议轻易打开。

### 代理环境

```yaml
system-proxy: ""
github-proxy: ""
ghproxy-group:
  - "https://example-gh-proxy/"
```

`system-proxy` 用于拉取订阅、访问 GitHub 代理检测等控制面请求，不用于节点自身测速结果造假。GitHub 订阅源可以使用 `github-proxy` 或 `ghproxy-group` 提高拉取成功率。

也可以使用标准环境变量：

```bash
export HTTP_PROXY=http://127.0.0.1:7890
export HTTPS_PROXY=http://127.0.0.1:7890
```

### WebUI 与 API

| 配置 | 作用 |
| --- | --- |
| `listen-port` | HTTP 服务监听地址，默认 `:8199` |
| `enable-web-ui` | 是否启用 Web 控制面板 |
| `api-key` | WebUI/API 鉴权密钥 |
| `output-dir` | 输出目录 |
| `callback-script` | 检测完成后执行的脚本 |

如果 `api-key` 为空，程序优先读取环境变量 `API_KEY`；仍为空时会生成随机值并打印到日志。

## 输出文件与 HTTP 访问

默认输出目录为程序所在目录下的 `output/`。设置 `output-dir` 后使用指定目录。

核心输出文件：

| 文件 | 含义 |
| --- | --- |
| `all.yaml` | 本轮检测通过的节点 |
| `history.yaml` | 历史成功节点去重合并结果 |
| `stats/` | 开启 `sub-urls-stats` 后生成的订阅统计 |

HTTP 路由：

| 路径 | 说明 | 鉴权 |
| --- | --- | --- |
| `/admin` | Web 控制面板 | 页面登录使用 API key |
| `/api/*` | 配置、状态、触发检测、日志等 API | `X-API-Key` |
| `/all.yaml` | 直接访问最新节点文件 | `X-API-Key` |
| `/history.yaml` | 直接访问历史节点文件 | `X-API-Key` |
| `/sub/<filename>` | 访问 `output/` 下文件 | 无内置密码 |
| `/more/<filename>` | 访问 `output/more/` 下文件 | 无内置密码 |

安全上要按公开服务处理 `/sub/*` 和 `/more/*`。如果部署到公网，建议使用反向代理、访问控制或 Cloudflare Tunnel 保护入口；不要把不准备公开的文件放入 `output/more/`。

## WebUI 操作

启动后访问：

```text
http://127.0.0.1:8199/admin
```

WebUI 当前提供：

- 查看和编辑配置文件。
- 手动触发检测。
- 强制停止当前检测。
- 查看最近检测状态。
- 查看运行日志。
- 复制常用订阅文件地址。

WebUI 不是多用户管理系统。它适合单实例、受控网络环境下的运维入口。公网暴露时需要额外访问控制。

## 与 Mihomo 集成

检测完成后可以让 Mihomo 重新拉取 HTTP provider：

```yaml
mihomo-api-url: "http://127.0.0.1:9090"
mihomo-api-secret: "your-secret"
```

程序会访问 Mihomo API：

- `GET /version`
- `GET /providers/proxies`
- `PUT /providers/proxies/{name}`

只有 `vehicleType` 为 `HTTP` 的 provider 会被更新。

## 运维建议

### 并发设置

先保守设置，再观察瓶颈：

```yaml
alive-concurrent: 200
speed-concurrent: 8
media-concurrent: 100
total-speed-limit: 0
download-timeout: 10
download-mb: 20
```

如果机器 CPU 空闲但检测慢，优先提高 `alive-concurrent`。如果出口带宽被打满，降低 `speed-concurrent` 或设置 `total-speed-limit`。媒体检测依赖外部平台响应，过高并发可能带来更多超时，不一定提升有效吞吐。

### 测速策略

测速会改变节点状态，有些节点会因为高频或高流量下载短期失效。建议：

- 节点少、网络差或只需要可用性时，保持 `speed-test-url: ""` 关闭测速。
- 开启测速时设置 `download-timeout` 和 `download-mb`。
- 在低峰时段运行检测。
- 使用 `threshold` 智能乱序，减少同网段节点连续测速。

### 订阅质量排查

开启统计：

```yaml
sub-urls-stats: true
success-rate: 0
```

程序会在 `output/stats/` 下记录订阅节点数量、可用数量和成功率。成功率持续偏低的订阅源应移除或单独检查。

### 日志和停止

程序会同时输出终端日志和临时日志文件。WebUI 的日志接口读取临时日志。

停止行为：

- 第一次中断信号会请求当前检测尽快停止。
- 第二次中断信号会触发更直接的退出流程。

## 开发命令

```bash
go test ./proxy ./check
go test ./save ./save/method ./utils ./app ./assets ./config
go test ./check/platform -run '^$'
go build -o /tmp/subs-check .
```

`check/platform` 下存在真实外网测速和外部服务测试。完整运行这部分测试会受网络环境影响，也可能生成测速结果文件。常规编译检查建议使用 `-run '^$'`。

常用构建命令：

```bash
make build
make build-all
make clean
```

## 版本和发布边界

当前仓库按 Linux 运行路径维护。发布配置和 Makefile 只构建 Linux 架构：

- linux/amd64
- linux/arm64
- linux/armv7
- linux/386

如果需要 Windows 或 macOS 支持，需要恢复对应的构建文件、标准输出实现和平台资源。这不属于当前简化版本的维护范围。

## 许可证与使用边界

本项目仅用于学习、测试和自有网络环境下的订阅整理。使用者需要自行确认订阅来源、节点使用和网络访问符合当地法律法规及服务条款。
