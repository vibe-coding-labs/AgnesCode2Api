package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/agnes"
)

var Version = "0.5.0"

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "显示版本信息",
	GroupID: "query",
	Example: `  agnescode-proxy version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AgnesCode Proxy %s\n", Version)
		fmt.Printf("  AgnesCode API: %s\n", agnescode.ClientVersion)
		fmt.Printf("  Go:          %s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
