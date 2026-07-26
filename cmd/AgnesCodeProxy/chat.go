package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/agnes"
)

var (
	chatModel     string
	chatStream    bool
	chatMaxTokens int
)

var chatCmd = &cobra.Command{
	Use:     "chat [message]",
	Short:   "发送聊天消息",
	Long:    "通过 AgnesCode API 发送一条聊天消息并返回响应。",
	GroupID: "core",
	Example: `  agnescode-proxy chat "你好"
  agnescode-proxy chat -m agnes-2.0-flash "写一个排序算法"
  agnescode-proxy chat -s "解释量子计算"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := resolveClient()
		if err != nil {
			return err
		}
		req := agnescode.ChatRequest{
			Model:     chatModel,
			MaxTokens: chatMaxTokens,
			Stream:    chatStream,
			Messages:  []agnescode.Message{{Role: "user", Content: args[0]}},
		}
		if chatStream {
			return streamChat(client, req)
		}
		resp, err := client.ChatCompletion(req)
		if err != nil {
			return err
		}
		if len(resp.Choices) > 0 {
			fmt.Println(resp.Choices[0].Message.Content)
		}
		return nil
	},
}

func streamChat(client *agnescode.Client, req agnescode.ChatRequest) error {
	ch, err := client.ChatCompletionStream(req)
	if err != nil {
		return err
	}
	for event := range ch {
		for _, c := range event.Choices {
			fmt.Print(c.Delta.Content)
		}
	}
	fmt.Println()
	return nil
}

func init() {
	chatCmd.Flags().StringVarP(&chatModel, "model", "m", "agnes-2.0-flash", "模型名称")
	chatCmd.Flags().BoolVarP(&chatStream, "stream", "s", false, "流式输出")
	chatCmd.Flags().IntVar(&chatMaxTokens, "max-tokens", 64000, "最大输出 token 数")
	rootCmd.AddCommand(chatCmd)
}