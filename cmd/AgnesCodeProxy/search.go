package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "列出可用模型",
	Long:    "通过 AgnesCode API 获取可用模型列表。",
	GroupID: "query",
	Example: `  agnescode-proxy search`,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := resolveClient()
		if err != nil {
			return err
		}
		models, err := client.ListModels()
		if err != nil {
			return err
		}
		if len(models) == 0 {
			fmt.Println("No models found.")
			return nil
		}
		fmt.Printf("Available models (%d):\n\n", len(models))
		for i, m := range models {
			free := "✅"
			if m.IsMemberOnly {
				free = "🔒"
			}
			fmt.Printf("  %d. %s %s (%s)\n", i+1, free, m.ID, m.OwnedBy)
			fmt.Printf("     %s\n", m.Description)
			fmt.Printf("     Input: %d | Output: %d\n", m.MaxInputTokens, m.MaxOutputTokens)
			fmt.Println()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}