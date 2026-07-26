package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/auth"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/agnes"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "显示当前配置信息",
	Long:    "显示已解析的凭据来源、默认设置和服务安装状态。",
	GroupID: "query",
	Example: `  agnescode-proxy config

  # 查看指定凭据的配置
  agnescode-proxy -k <ptkey> -u <userid> config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("AgnesCode Proxy Configuration")
		fmt.Println("============================")

		// Credentials
		fmt.Println()
		fmt.Println("  Credentials:")
		if PtKey != "" && userID != "" {
			fmt.Printf("    Source:    flags\n")
			fmt.Printf("    UserID:    %s\n", userID)
			fmt.Printf("    PtKey:     %s...%s\n", PtKey[:8], PtKey[len(PtKey)-4:])
		} else {
			creds, err := auth.LoadFromSystem()
			if err != nil {
				fmt.Printf("    Source:    not available (%s)\n", err)
			} else {
				fmt.Printf("    Source:    auto-detected\n")
				fmt.Printf("    UserID:    %s\n", creds.UserID)
				fmt.Printf("    PtKey:     %s...%s\n", creds.PtKey[:8], creds.PtKey[len(creds.PtKey)-4:])
			}
		}

		// API
		fmt.Println()
		fmt.Println("  API:")
		fmt.Printf("    Base URL:       %s\n", agnescode.BFFBaseURL)
		fmt.Printf("    Default Model:  %s\n", agnescode.DefaultModel)

		// Server
		fmt.Println()
		fmt.Println("  Server:")
		fmt.Printf("    Default Host:  %s\n", "0.0.0.0")
		fmt.Printf("    Default Port:  %d\n", 34891)

		// Service
		fmt.Println()
		fmt.Println("  Service:")
		home, _ := os.UserHomeDir()
		plistPath := filepath.Join(home, "Library", "LaunchAgents", plistName)
		if _, err := os.Stat(plistPath); os.IsNotExist(err) {
			fmt.Println("    Installed: no")
		} else {
			fmt.Println("    Installed: yes")
			fmt.Printf("    Plist:     %s\n", plistPath)
		}

		// Models
		fmt.Println()
		fmt.Println("  Available Models:")
		for _, m := range []string{} {
			suffix := ""
			if m == agnescode.DefaultModel {
				suffix = " (default)"
			}
			fmt.Printf("    - %s%s\n", m, suffix)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
