package app

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// SharePageData 定义渲染分享页面所需的数据
type SharePageData struct {
	Title       string
	HeaderColor string        // 标题颜色
	HeaderIcon  string        // 图标
	HeaderTitle string        // 标题文字
	Description template.HTML // 描述文本
	PathExample string        // 路径示例
	ExtraHint   template.HTML // 额外提示
	FooterText  string        // 底部文字
}

var sharePageTmpl = template.Must(template.New("share").Parse(sharePageTemplateStr))

// 渲染并发送响应
func renderSharePage(c *gin.Context, statusCode int, data SharePageData) {
	c.Status(statusCode) // 显式设置状态码
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := sharePageTmpl.Execute(c.Writer, data); err != nil {
		slog.Error("渲染分享页面失败", "error", err)
		// 如果模板渲染一半出错了，很难再改状态码，只能记录日志
	}
}

// handleFileShare 返回一个处理文件分享的 Handler
func (app *App) handleFileShare(basePath string, isSecret bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		relPath := c.Param("filepath")

		// 访问根目录：显示欢迎页
		if relPath == "" || relPath == "/" {
			if isSecret {
				renderSharePage(c, http.StatusOK, SharePageData{
					Title:       "Subs-Check 文件分享",
					HeaderColor: "#009768", // 绿色
					HeaderIcon:  "",
					HeaderTitle: "订阅分享",
					Description: template.HTML("您正在访问 <b>/output/</b> 下的订阅文件。"),
					PathExample: "/sub/filename.txt",
					ExtraHint:   template.HTML("请只分享确实需要公开的订阅文件。"),
					FooterText:  "订阅分享未使用密码。",
				})
			} else {
				renderSharePage(c, http.StatusOK, SharePageData{
					Title:       "Subs-Check-PRO 文件分享",
					HeaderColor: "#d9534f", // 红色
					HeaderIcon:  "",
					HeaderTitle: "注意",
					Description: template.HTML("您正在访问 <b>无密码保护的目录</b>。"),
					PathExample: "/more/filename.txt",
					ExtraHint:   template.HTML("请勿在该目录存放敏感文件，以免资源泄露！"),
					FooterText:  "除非文件确实没啥泄露价值。",
				})
			}
			return
		}

		// 安全检查
		absPath := filepath.Join(basePath, filepath.Clean(relPath))

		// 防止路径穿越 (403 Forbidden)
		if !strings.HasPrefix(absPath, basePath) {
			renderSharePage(c, http.StatusForbidden, SharePageData{
				Title:       "非法访问 - Subs-Check-PRO",
				HeaderColor: "#d9534f", // 红色
				HeaderIcon:  "",
				HeaderTitle: "访问被拒绝",
				Description: template.HTML(fmt.Sprintf("检测到非法路径请求：<code>%s</code>", relPath)),
				PathExample: "/",
				ExtraHint:   template.HTML("系统已拦截该请求。<br>请勿尝试访问授权目录之外的文件。"),
				FooterText:  "403 Forbidden",
			})
			return
		}

		// 文件存在检查 (404 Not Found)
		info, err := os.Stat(absPath)
		if err != nil || info.IsDir() {
			// 确定示例路径（方便用户点击回去）
			examplePath := "/more/filename.txt"
			if isSecret {
				examplePath = "/sub/filename.txt"
			}

			// 渲染 404 页面
			renderSharePage(c, http.StatusNotFound, SharePageData{
				Title:       "文件不存在 - Subs-Check-PRO",
				HeaderColor: "#d40000ff", // 表示警告/错误
				HeaderIcon:  "",
				HeaderTitle: "错误！",
				Description: template.HTML(fmt.Sprintf("未找到文件 <code>%s</code>", relPath)),
				PathExample: examplePath, // 显示正确的格式给用户参考
				ExtraHint:   template.HTML("可能是文件名拼写错误，或者该文件已被删除。<br>请检查 URL 是否正确。"),
				FooterText:  "404 Not Found",
			})
			return
		}

		// 返回文件
		c.File(absPath)
	}
}

// 统一 HTML 模板
const sharePageTemplateStr = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <style>
        * { box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; 
            margin: 0; 
            background: #fafafa; 
            display: flex; 
            justify-content: center; 
            align-items: center; 
            min-height: 100vh; 
            padding: 10px; 
        }
        .box { 
            padding: 2em; 
            border: 1px solid #cccccca7; 
            border-radius: 12px; 
            background: #fff; 
            width: 92%; 
            max-width: 650px; 
            box-shadow: 0 10px 25px rgba(0,0,0,0.05); 
        }
        h2 { color: {{ .HeaderColor }}; margin-top: 0; }
        p { margin: 0.5em 0; line-height: 1.6; }
        code {    
            background: #f7f7f5; 
            padding: 3px 8px; 
            border-radius: 6px; 
            font-family: "Menlo", "Monaco", monospace; 
            color: #5d5454; 
            font-size: 0.9em; 
            word-break: break-all; 
            border: 1px solid #eee;  
        }
        @media (max-width: 768px) {
            .box { width: 96%; padding: 1.2em; }
        }  
    </style>
</head>
<body>
    <div class="box">
        <h2>{{ .HeaderIcon }} {{ .HeaderTitle }}</h2>
        <p>{{ .Description }}</p>
        <p>请输入正确的文件名访问，例如：<code>{{ .PathExample }}</code></p>
        <br>
        <b>💡 提示：</b>
        <p>如需保留之前成功的代理节点，仅需开启 <code>keep-success-proxies: true</code></p>
        <br>
        <p>{{ .ExtraHint }}</p>
        <p style="font-size: 0.9em; color: #999;">{{ .FooterText }}</p>
    </div>
</body>
</html>
`
