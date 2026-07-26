package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Short:   "查看当前认证用户信息",
	GroupID: "query",
	Example: `  agnescode-proxy whoami`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := resolveClient()
		if err != nil {
			return err
		}
		user, err := client.GetUserInfo()
		if err != nil {
			return err
		}
		fmt.Printf("  User: %s (%s)\n", user.Username, user.Email)
		fmt.Printf("  ID: %s\n", user.ID)
		fmt.Printf("  Active: %v\n", user.IsActive)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}