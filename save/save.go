// Package save 保存检测结果
package save

import (
	"fmt"
	proxyutils "github.com/sinspired/subs-check/proxy"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/sinspired/subs-check/check"
	"github.com/sinspired/subs-check/save/method"
)

// ProxyCategory 定义代理分类
type ProxyCategory struct {
	Name    string
	Proxies []map[string]any
	Filter  func(result check.Result) bool
}

// ConfigSaver 处理配置保存的结构体
type ConfigSaver struct {
	results    []check.Result
	categories []ProxyCategory
	saveMethod func([]byte, string) error
}

// NewConfigSaver 创建新的配置保存器
func NewConfigSaver(results []check.Result) *ConfigSaver {
	return &ConfigSaver{
		results:    results,
		saveMethod: chooseSaveMethod(),
		categories: []ProxyCategory{
			{
				Name:    "all.yaml",
				Proxies: make([]map[string]any, 0),
				Filter:  func(result check.Result) bool { return true },
			},
			{
				Name:    "history.yaml", // 新增
				Proxies: make([]map[string]any, 0),
				Filter:  func(result check.Result) bool { return true }, // 这里可加条件
			},
		},
	}
}

// SaveConfig 保存配置的入口函数
func SaveConfig(results []check.Result) {
	saver := NewConfigSaver(results)
	if err := saver.Save(); err != nil {
		slog.Error(fmt.Sprintf("保存配置失败: %v", err))
	}
}

// Save 执行保存操作
func (cs *ConfigSaver) Save() error {
	// 分类处理代理
	cs.categorizeProxies()

	// 保存各个类别的代理
	for _, category := range cs.categories {
		if err := cs.saveCategory(category); err != nil {
			slog.Error(fmt.Sprintf("保存到本地失败: %v", err))
			continue
		}
	}

	return nil
}

// categorizeProxies 将代理按类别分类
func (cs *ConfigSaver) categorizeProxies() {
	for _, result := range cs.results {
		for i := range cs.categories {
			if cs.categories[i].Filter(result) {
				cs.categories[i].Proxies = append(cs.categories[i].Proxies, result.Proxy)
			}
		}
	}
}

// saveCategory 保存单个类别的代理
func (cs *ConfigSaver) saveCategory(category ProxyCategory) error {
	if len(category.Proxies) == 0 {
		slog.Warn(fmt.Sprintf("yaml节点为空，跳过保存: %s", category.Name))
		return nil
	}
	if category.Name == "history.yaml" {
		saver, err := method.NewLocalSaver()
		if err != nil {
			return fmt.Errorf("本地存储初始化失败，无法启用历史记录功能: %w", err)
		}
		if !filepath.IsAbs(saver.OutputPath) {
			// 处理用户写相对路径的问题
			saver.OutputPath = filepath.Join(saver.BasePath, saver.OutputPath)
		}

		// 读取已有文件
		existing := make([]map[string]any, 0)

		outputPath := saver.OutputPath
		filepath := filepath.Join(outputPath, category.Name)
		// 读取原有历史记录
		data, err := ReadFileIfExists(filepath)
		if err == nil && len(data) > 0 {
			var parsed map[string][]map[string]any
			if err := yaml.Unmarshal(data, &parsed); err == nil {
				existing = parsed["proxies"]
			}
		}

		// 合并去重
		merged := mergeUniqueProxies(existing, category.Proxies)

		// 序列化
		yamlData, err := yaml.Marshal(map[string]any{
			"proxies": merged,
		})
		if err != nil {
			return fmt.Errorf("序列化yaml %s 失败: %w", category.Name, err)
		}

		// 保存（这里直接覆盖写入，因为 merged 已经包含旧数据，相当于逻辑上的“追加”）
		if err := cs.saveMethod(yamlData, category.Name); err != nil {
			return fmt.Errorf("保存 %s 失败: %w", category.Name, err)
		}
		return nil
	}
	if category.Name == "all.yaml" {
		yamlData, err := yaml.Marshal(map[string]any{
			"proxies": category.Proxies,
		})
		if err != nil {
			return fmt.Errorf("序列化yaml %s 失败: %w", category.Name, err)
		}
		if err := cs.saveMethod(yamlData, category.Name); err != nil {
			return fmt.Errorf("保存 %s 失败: %w", category.Name, err)
		}
		return nil
	}

	return nil
}

// chooseSaveMethod 固定使用本地保存。
func chooseSaveMethod() func([]byte, string) error {
	return method.SaveToLocal
}

func mergeUniqueProxies(existing, newProxies []map[string]any) []map[string]any {
	seen := make(map[string]bool)
	result := make([]map[string]any, 0, len(existing)+len(newProxies))

	// 先加旧的
	for _, p := range existing {
		delete(p, "sub_was_succeed")  // 删除旧的标记
		delete(p, "sub_from_history") // 删除旧的标记
		key := proxyutils.GenerateProxyKey(p)
		if !seen[key] {
			seen[key] = true
			result = append(result, p)
		}
	}

	// 再加新的
	for _, p := range newProxies {
		delete(p, "sub_was_succeed")  // 删除旧的标记
		delete(p, "sub_from_history") // 删除旧的标记
		key := proxyutils.GenerateProxyKey(p)
		if !seen[key] {
			seen[key] = true
			result = append(result, p)
		}
	}

	return result
}

func ReadFileIfExists(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	return os.ReadFile(path)
}
