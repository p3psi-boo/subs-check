// Package config 解析配置文件
package config

import (
	_ "embed"
)

type Config struct {
	PrintProgress        bool     `yaml:"print-progress"`
	ProgressMode         string   `yaml:"progress-mode"`
	Concurrent           int      `yaml:"concurrent"`
	AliveConcurrent      int      `yaml:"alive-concurrent"`
	SpeedConcurrent      int      `yaml:"speed-concurrent"`
	MediaConcurrent      int      `yaml:"media-concurrent"`
	CheckInterval        int      `yaml:"check-interval"`
	CronExpression       string   `yaml:"cron-expression"`
	SpeedTestURL         string   `yaml:"speed-test-url"`
	DownloadTimeout      int      `yaml:"download-timeout"`
	DownloadMB           int      `yaml:"download-mb"`
	TotalSpeedLimit      int      `yaml:"total-speed-limit"`
	Threshold            float32  `yaml:"threshold"`
	MinSpeed             int      `yaml:"min-speed"`
	Timeout              int      `yaml:"timeout"`
	FilterRegex          string   `yaml:"filter-regex"`
	SubUrlsReTry         int      `yaml:"sub-urls-retry"`
	SubUrlsRetryInterval int      `yaml:"sub-urls-retry-interval"`
	SubUrlsTimeout       int      `yaml:"sub-urls-timeout"`
	SubUrlsRemote        []string `yaml:"sub-urls-remote"`
	SubUrls              []string `yaml:"sub-urls"`
	SubURLsStats         bool     `yaml:"sub-urls-stats"`
	SuccessRate          float32  `yaml:"success-rate"`
	MihomoAPIURL         string   `yaml:"mihomo-api-url"`
	MihomoAPISecret      string   `yaml:"mihomo-api-secret"`
	ListenPort           string   `yaml:"listen-port"`
	RenameNode           bool     `yaml:"rename-node"`
	KeepSuccessProxies   bool     `yaml:"keep-success-proxies"`
	OutputDir            string   `yaml:"output-dir"`
	ISPCheck             bool     `yaml:"isp-check"`
	MediaCheck           bool     `yaml:"media-check"`
	Platforms            []string `yaml:"platforms"`
	MaxMindDBPath        string   `yaml:"maxmind-db-path"`
	DropBadCfNodes       bool     `yaml:"drop-bad-cf-nodes"`
	EnhancedTag          bool     `yaml:"enhanced-tag"`
	SuccessLimit         int32    `yaml:"success-limit"`
	NodePrefix           string   `yaml:"node-prefix"`
	NodeType             []string `yaml:"node-type"`
	EnableWebUI          bool     `yaml:"enable-web-ui"`
	APIKey               string   `yaml:"api-key"`
	CallbackScript       string   `yaml:"callback-script"`
	SystemProxy          string   `yaml:"system-proxy"`
	GithubProxy          string   `yaml:"github-proxy"`
	GithubProxyGroup     []string `yaml:"ghproxy-group"`
}

var OriginDefaultConfig = &Config{
	// 新增配置，给未更改配置文件的用户一个默认值
	ListenPort: ":8199",
	Platforms: []string{
		"iprisk",
		"openai",
		"gemini",
		"youtube",
		// "netflix",
		// "disney",
	},
	DownloadMB: 20,
	// ISPCheck:    true,
}

// GlobalConfig 指向当前生效配置
var GlobalConfig = &Config{} // 初始化为空，首次加载后赋值

//go:embed config.yaml.example
var DefaultConfigTemplate []byte
