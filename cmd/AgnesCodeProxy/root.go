package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/auth"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/agnes"
)

var (
	JWTToken          string
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
	rootCmd.PersistentFlags().StringVarP(&JWTToken, "jwttoken", "k", "", "AgnesCode JWTToken（留空则自动从客户端检测）")
	rootCmd.PersistentFlags().StringVarP(&userID, "userid", "u", "", "AgnesCode userID（留空则自动从客户端检测）")
	rootCmd.PersistentFlags().BoolVar(&skipValidation, "skip-validation", false, "跳过凭据验证")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "启用调试日志")
}

func resolveClient() (*agnescode.Client, error) {
	var creds *auth.Credentials
	var source string

	if JWTToken != "" && userID != "" {
		creds = &auth.Credentials{JWTToken: JWTToken, UserID: userID}
		source = "flags"
	} else {
		detected, err := auth.LoadFromSystem()
		if err != nil {
			if !skipValidation {
				return nil, fmt.Errorf("cannot auto-detect credentials: %w\n  Please provide --jwttoken and --userid flags, or log in to AgnesCode first", err)
			}
			// With --skip-validation, create a placeholder client; real requests use DB accounts via resolver
			log.Printf("Warning: cannot auto-detect credentials (%v); using placeholder (requests will use DB accounts)", err)
			creds = &auth.Credentials{JWTToken: "placeholder", UserID: "placeholder"}
			source = "placeholder (no local AgnesCode session)"
		} else {
			creds = detected
			source = "auto-detected"

			if JWTToken != "" {
				creds.JWTToken = JWTToken
				source = "flags+auto-detected"
			}
			if userID != "" {
				creds.UserID = userID
				source = "flags+auto-detected"
			}
		}
	}

	log.Printf("Credentials source: %s (userId=%s)", source, creds.UserID)
	client := agnescode.NewClient(creds.JWTToken)

	if skipValidation {
		log.Printf("Credential validation skipped (--skip-validation)")
		return client, nil
	}

	log.Printf("Validating credentials...")
	if err := client.Authenticate(); err != nil {
		return nil, fmt.Errorf("%w\n  Your credentials may have expired. Try re-logging into AgnesCode or provide fresh --jwttoken and --userid", err)
	}
	log.Printf("Credentials validated successfully")
	return client, nil
}
