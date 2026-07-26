# CLI UX 精细化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: `superpowers:subagent-driven-development`
> Steps use checkbox (`- [ ]`) syntax.

**Goal:** 全面优化 CLI 的 help 文本、Example 示例、命令分组、参数描述，使 AgnesCodeProxy 成为一个专业级的命令行工具。

**Architecture:** 利用 cobra 的 GroupID 特性将命令分为 core/service/query 三组，为每个命令添加 Example 区域和中文 Short 描述，root 命令开启 SilenceUsage/SilenceErrors 减少噪音输出。所有改动集中在 cmd/AgnesCodeProxy/ 目录下，不涉及 pkg/ 层。

**Tech Stack:** Go 1.23, spf13/cobra v1.10.2

**Risks:**
- cobra GroupID 在 help 输出中可能排列不理想 → 缓解：使用 `AddGroup` 控制顺序
- 中文描述在某些终端可能显示乱码 → 缓解：Go 默认 UTF-8，现代终端均支持

---

### Task 1: 优化 root 命令和 main 入口 — 添加命令分组、增强 help、SilenceErrors

**Depends on:** None
**Files:**
- Modify: `cmd/AgnesCodeProxy/main.go:1-10`
- Modify: `cmd/AgnesCodeProxy/root.go:1-73`

- [ ] **Step 1: 修改 main.go — 添加命令分组定义**
文件: `cmd/AgnesCodeProxy/main.go:1-10`（替换整个文件）

```go
package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd.AddGroup(
		&cobra.Group{ID: "core", Title: "Core Commands:"},
		&cobra.Group{ID: "service", Title: "Service Management:"},
		&cobra.Group{ID: "query", Title: "Query & Info:"},
	)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: 修改 root.go — 增强 Long 描述、SilenceUsage、中文 flag 描述**
文件: `cmd/AgnesCodeProxy/root.go:1-73`（替换整个文件）

```go
package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/vibe-coding-labs/AgnesCodeProxy/pkg/auth"
	"github.com/vibe-coding-labs/AgnesCodeProxy/pkg/agnescode"
)

var (
	ptKey          string
	userID         string
	skipValidation bool
	verbose        bool
)

var rootCmd = &cobra.Command{
	Use:   "agnescode-proxy",
	Short: "AgnesCode API Proxy — 将 AgnesCode API 转换为 OpenAI/Anthropic 兼容格式",
	Long: `AgnesCode API Proxy — 将 AgnesCode 内部 API 转换为 OpenAI / Anthropic 兼容格式。

让 Claude Code、Codex 等 AI 编程工具可以直接使用 AgnesCode 的模型服务。

快速开始:
  agnescode-proxy serve                  # 启动代理服务器（默认端口 34891）
  agnescode-proxy service install        # 安装为 macOS 服务（开机自启、崩溃重启）
  agnescode-proxy check                  # 检查代理是否运行

配置 Claude Code:
  export ANTHROPIC_BASE_URL=http://localhost:34891
  export ANTHROPIC_API_KEY=agnescode
  claude`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ptKey, "ptkey", "k", "", "AgnesCode ptKey（留空则自动从客户端检测）")
	rootCmd.PersistentFlags().StringVarP(&userID, "userid", "u", "", "AgnesCode userID（留空则自动从客户端检测）")
	rootCmd.PersistentFlags().BoolVar(&skipValidation, "skip-validation", false, "跳过凭据验证")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "启用调试日志")
}

func resolveClient() (*agnescode.Client, error) {
	var creds *auth.Credentials
	var source string

	if ptKey != "" && userID != "" {
		creds = &auth.Credentials{PtKey: ptKey, UserID: userID}
		source = "flags"
	} else {
		detected, err := auth.LoadFromSystem()
		if err != nil {
			return nil, fmt.Errorf("cannot auto-detect credentials: %w\n  Please provide --ptkey and --userid flags, or log in to AgnesCode first", err)
		}
		creds = detected
		source = "auto-detected"

		// Partial override: flag value takes precedence
		if ptKey != "" {
			creds.PtKey = ptKey
			source = "flags+auto-detected"
		}
		if userID != "" {
			creds.UserID = userID
			source = "flags+auto-detected"
		}
	}

	log.Printf("Credentials source: %s (userId=%s)", source, creds.UserID)
	client := agnescode.NewClient(creds.PtKey, creds.UserID)

	if skipValidation {
		log.Printf("Credential validation skipped (--skip-validation)")
		return client, nil
	}

	log.Printf("Validating credentials...")
	if err := client.Validate(); err != nil {
		return nil, fmt.Errorf("%w\n  Your credentials may have expired. Try re-logging into AgnesCode or provide fresh --ptkey and --userid", err)
	}
	log.Printf("Credentials validated successfully")
	return client, nil
}
```

- [ ] **Step 3: 验证 root 命令 help 输出**
Run: `go build -o agnescode_proxy_bin ./cmd/AgnesCodeProxy/ && ./agnescode_proxy_bin --help`
Expected:
  - Exit code: 0
  - Output contains: "Core Commands:", "Service Management:", "Query & Info:"
  - Output contains: "将 AgnesCode API 转换为 OpenAI/Anthropic 兼容格式"
  - Output contains: "快速开始"

- [ ] **Step 4: 提交**
Run: `git add cmd/AgnesCodeProxy/main.go cmd/AgnesCodeProxy/root.go && git commit -m "refactor(cli): enhance root command with groups, Chinese help, SilenceErrors"`

---

### Task 2: 增强所有子命令 — 添加 GroupID、Example、中文 Short

**Depends on:** Task 1
**Files:**
- Modify: `cmd/AgnesCodeProxy/serve.go:23-27`
- Modify: `cmd/AgnesCodeProxy/service.go:13-23,25-28,179-184`
- Modify: `cmd/AgnesCodeProxy/chat.go:16-20`
- Modify: `cmd/AgnesCodeProxy/models.go:10-14`
- Modify: `cmd/AgnesCodeProxy/whoami.go:9-12`
- Modify: `cmd/AgnesCodeProxy/version.go:12-18`
- Modify: `cmd/AgnesCodeProxy/config.go:13-17`
- Modify: `cmd/AgnesCodeProxy/check.go:14-18`
- Modify: `cmd/AgnesCodeProxy/search.go:9-13`

- [ ] **Step 1: 修改 serve.go — 添加 GroupID、Example、中文描述**

在 serveCmd 定义中添加 GroupID 和 Example。修改 `cmd/AgnesCodeProxy/serve.go:23-27`：

```go
var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "启动代理服务器",
	Long:    "启动 OpenAI/Anthropic 兼容的 API 代理服务器，将请求转换为 AgnesCode API 格式。",
	GroupID: "core",
	Example: `  # 默认启动（0.0.0.0:34891）
  agnescode-proxy serve

  # 指定端口
  agnescode-proxy serve -p 8080

  # 启用调试日志
  agnescode-proxy -v serve

  # 跳过凭据验证（用于测试）
  agnescode-proxy serve --skip-validation`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

- [ ] **Step 2: 修改 service.go — 添加 GroupID、中文描述**

修改 `cmd/AgnesCodeProxy/service.go:19-23`：

```go
var serviceCmd = &cobra.Command{
	Use:     "service",
	Short:   "管理 macOS 服务（安装/卸载/状态）",
	Long:    "将 AgnesCode Proxy 安装为 macOS launchd 服务，支持开机自启和崩溃自动重启。",
	GroupID: "service",
}
```

修改 `cmd/AgnesCodeProxy/service.go:25-28`：

```go
var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装并启动 launchd 服务",
	Long:  "将代理安装为 macOS launchd 服务。安装后自动启动，支持开机自启和崩溃自动重启。",
	Example: `  # 使用默认端口 34891 安装
  agnescode-proxy service install

  # 指定端口
  agnescode-proxy service install -p 8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

修改 `cmd/AgnesCodeProxy/service.go:120-123`（serviceUninstallCmd）：

```go
var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "停止并移除 launchd 服务",
	Example: `  agnescode-proxy service uninstall`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

修改 `cmd/AgnesCodeProxy/service.go:143-146`（serviceStatusCmd）：

```go
var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看服务运行状态",
	Example: `  agnescode-proxy service status`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

- [ ] **Step 3: 修改 chat.go — 添加 GroupID、Example、中文描述**

修改 `cmd/AgnesCodeProxy/chat.go:16-20`：

```go
var chatCmd = &cobra.Command{
	Use:     "chat [message]",
	Short:   "发送聊天消息",
	Long:    "通过 AgnesCode API 发送一条聊天消息并返回响应。",
	GroupID: "core",
	Example: `  # 发送简单消息
  agnescode-proxy chat "你好"

  # 指定模型
  agnescode-proxy chat -m GLM-5.1 "写一个排序算法"

  # 流式输出
  agnescode-proxy chat -s "解释量子计算"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
```

- [ ] **Step 4: 修改 models.go、whoami.go、version.go — 添加 GroupID、Example**

修改 `cmd/AgnesCodeProxy/models.go:10-14`：

```go
var modelsCmd = &cobra.Command{
	Use:     "models",
	Short:   "列出可用的 AI 模型",
	GroupID: "query",
	Example: `  agnescode-proxy models`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

修改 `cmd/AgnesCodeProxy/whoami.go:9-12`：

```go
var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Short:   "查看当前认证用户信息",
	GroupID: "query",
	Example: `  agnescode-proxy whoami`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

修改 `cmd/AgnesCodeProxy/version.go:12-18`：

```go
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "显示版本信息",
	GroupID: "query",
	Example: `  agnescode-proxy version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AgnesCode Proxy %s\n", Version)
		fmt.Printf("  AgnesCode API: %s\n", agnescode.ClientVersion)
		fmt.Printf("  Go:          %s\n", goVersion())
	},
}
```

在 version.go 末尾添加 goVersion 函数：

```go
func goVersion() string {
	return runtime.Version()
}
```

并在 version.go import 中添加 `"runtime"`：

```go
import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/vibe-coding-labs/AgnesCodeProxy/pkg/agnescode"
)
```

- [ ] **Step 5: 修改 config.go、check.go、search.go — 添加 GroupID、Example**

修改 `cmd/AgnesCodeProxy/config.go:13-17`：

```go
var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "显示当前配置信息",
	Long:    "显示已解析的凭据来源、默认设置和服务安装状态。",
	GroupID: "query",
	Example: `  agnescode-proxy config

  # 查看指定凭据的配置
  agnescode-proxy -k <ptkey> -u <userid> config`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

修改 `cmd/AgnesCodeProxy/check.go:14-18`：

```go
var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "检查代理服务是否运行",
	Long:    "向本地代理发送健康检查请求，验证服务是否正常运行。",
	GroupID: "core",
	Example: `  # 检查默认端口
  agnescode-proxy check

  # 检查指定端口
  agnescode-proxy check -p 8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
```

修改 `cmd/AgnesCodeProxy/search.go:9-13`：

```go
var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "网页搜索",
	Long:    "使用 AgnesCode 内置搜索 API 进行网页搜索。",
	GroupID: "query",
	Example: `  agnescode-proxy search "Go 语言并发编程"`,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
```

- [ ] **Step 6: 验证所有命令 help 输出**
Run: `go build -o agnescode_proxy_bin ./cmd/AgnesCodeProxy/ && ./agnescode_proxy_bin --help && echo "===" && ./agnescode_proxy_bin serve --help && echo "===" && ./agnescode_proxy_bin chat --help`
Expected:
  - Exit code: 0
  - `--help` 输出包含分组标题："Core Commands:", "Service Management:", "Query & Info:"
  - `serve --help` 包含 "Example:" 和中文 Short 描述
  - `chat --help` 包含 "Example:" 和多个示例

- [ ] **Step 7: 提交**
Run: `git add cmd/AgnesCodeProxy/ && git commit -m "refactor(cli): add GroupID, Example, Chinese help to all commands"`

---

### Task 3: 增强 serve 输出和 service install 的 plist 内容 — 输出更专业

**Depends on:** Task 2
**Files:**
- Modify: `cmd/AgnesCodeProxy/serve.go:49-66`（启动 banner）
- Modify: `cmd/AgnesCodeProxy/version.go:12-18`（增强版本输出）

- [ ] **Step 1: 修改 serve.go 启动 banner — 添加版本号和分隔线**

修改 `cmd/AgnesCodeProxy/serve.go:49-66`（go func 中的输出部分）：

```go
		go func() {
			log.Printf("AgnesCode Proxy running on http://%s", addr)
			fmt.Println()
			fmt.Printf("  AgnesCode Proxy %s\n", Version)
			fmt.Println("  ─────────────────────────────────────────────────")
			fmt.Println()
			fmt.Println("  Endpoints:")
			fmt.Println("    POST /v1/chat/completions  — Chat (OpenAI format)")
			fmt.Println("    POST /v1/messages          — Chat (Anthropic/Claude Code format)")
			fmt.Println("    POST /v1/web-search        — Web Search")
			fmt.Println("    POST /v1/rerank            — Rerank documents")
			fmt.Println("    GET  /v1/models            — Model list")
			fmt.Println("    GET  /health               — Health check")
			fmt.Println()
			fmt.Println("  Claude Code setup:")
			fmt.Printf("    export ANTHROPIC_BASE_URL=http://%s\n", addr)
			fmt.Println("    export ANTHROPIC_API_KEY=agnescode")
			if verbose {
				fmt.Println()
				fmt.Println("  Verbose logging: enabled")
			}
			fmt.Println()
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()
```

- [ ] **Step 2: 验证 serve 输出**
Run: `go build -o agnescode_proxy_bin ./cmd/AgnesCodeProxy/ && ./agnescode_proxy_bin --skip-validation serve -p 34892 &
sleep 2
kill %1 2>/dev/null
wait 2>/dev/null`
Expected:
  - Output contains: "AgnesCode Proxy 0.1.0"
  - Output contains: "─────────────────────────────────────────────────"
  - Exit code: 0

- [ ] **Step 3: 提交**
Run: `git add cmd/AgnesCodeProxy/serve.go && git commit -m "feat(cli): add version banner and separator to serve output"`
