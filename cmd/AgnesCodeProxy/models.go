package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/agnes"
)

var modelsCmd = &cobra.Command{
	Use:     "models",
	Short:   "列出可用的 AI 模型",
	GroupID: "query",
	Example: `  agnescode-proxy models`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := resolveClient()
		if err != nil {
			return err
		}
		models, err := client.ListModels()
		if err != nil {
			return err
		}
		for _, m := range models {
			pref := ""
			if m.ID == agnescode.DefaultModel {
				pref = " *"
			}
			fmt.Printf("  %s (%s) ctx=%d out=%d%s\n",
				m.ID, m.ID, m.MaxInputTokens, m.MaxOutputTokens, pref)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
}
