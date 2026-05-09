package utils

import (
	"strings"
	"time"

	"github.com/sinspired/subs-check/config"
)

// WarpURL applies date placeholders and the configured GitHub proxy when needed.
func WarpURL(rawURL string, isGhProxyAvailable bool) string {
	rawURL = formatTimePlaceholders(rawURL, time.Now())

	if strings.HasPrefix(rawURL, "https://raw.githubusercontent.com") && isGhProxyAvailable {
		return config.GlobalConfig.GithubProxy + rawURL
	}
	if strings.Contains(rawURL, "/raw") && strings.Contains(rawURL, "github.com/") {
		return config.GlobalConfig.GithubProxy + rawURL
	}
	return rawURL
}

func formatTimePlaceholders(rawURL string, t time.Time) string {
	replacer := strings.NewReplacer(
		"{Y}", t.Format("2006"),
		"{y}", t.Format("06"),
		"{m}", t.Format("01"),
		"{mm}", t.Format("01"),
		"{d}", t.Format("02"),
		"{dd}", t.Format("02"),
		"{Ymd}", t.Format("20060102"),
		"{ymd}", t.Format("060102"),
		"{Y-m-d}", t.Format("2006-01-02"),
		"{y-m-d}", t.Format("06-01-02"),
		"{Y_m_d}", t.Format("2006_01_02"),
		"{y_m_d}", t.Format("06_01_02"),
	)
	return replacer.Replace(rawURL)
}
