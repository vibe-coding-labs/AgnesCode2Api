package openai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/agnes"
)

func TranslateRequest(req *ChatRequest) agnescode.ChatRequest {
	cr := agnescode.ChatRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Temperature: 0.7,
		TopP:        1.0,
	}
	if req.Temperature != nil {
		cr.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		cr.TopP = *req.TopP
	}
	// Convert messages from json.RawMessage to []Message
	if len(req.Messages) > 0 {
		var rawMsgs []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(req.Messages, &rawMsgs); err == nil {
			msgs := make([]agnescode.Message, 0, len(rawMsgs))
			for _, m := range rawMsgs {
				msgs = append(msgs, agnescode.Message{Role: m.Role, Content: m.Content})
			}
			cr.Messages = msgs
		}
	}
	// Convert stop sequences
	if len(req.Stop) > 0 {
		var stops []string
		if err := json.Unmarshal(req.Stop, &stops); err == nil {
			cr.Stop = stops
		}
	}
	return cr
}

func TranslateResponse(resp *agnescode.ChatResponse, model string) map[string]interface{} {
	choices := make([]map[string]interface{}, len(resp.Choices))
	for i, c := range resp.Choices {
		choices[i] = map[string]interface{}{
			"index":         c.Index,
			"finish_reason": c.FinishReason,
			"message": map[string]interface{}{
				"role":    c.Message.Role,
				"content": c.Message.Content,
			},
		}
	}
	return map[string]interface{}{
		"id":                 fmt.Sprintf("chatcmpl-%s", newShortID()),
		"object":             "chat.completion",
		"created":            time.Now().Unix(),
		"model":              model,
		"choices":            choices,
		"usage":              map[string]interface{}{"prompt_tokens": resp.Usage.PromptTokens, "completion_tokens": resp.Usage.CompletionTokens, "total_tokens": resp.Usage.TotalTokens},
		"system_fingerprint": fmt.Sprintf("fp_%s", newShortID()),
	}
}

func TranslateModels(apiModels []agnescode.ModelInfo) map[string]interface{} {
	data := make([]map[string]interface{}, 0, len(apiModels))
	for _, m := range apiModels {
		entry := map[string]interface{}{
			"id": m.ID, "object": "model",
			"created": 1700000000, "owned_by": m.OwnedBy,
			"description":       m.Description,
			"max_input_tokens":  m.MaxInputTokens,
			"max_output_tokens": m.MaxOutputTokens,
			"is_member_only":    m.IsMemberOnly,
			"is_gray":           m.IsGray,
			"model_type":        m.ModelType,
		}
		data = append(data, entry)
	}
	return map[string]interface{}{"object": "list", "data": data}
}

func TranslateStreamChunk(data string, model string) string {
	if data == "[DONE]" {
		return "data: [DONE]\n\n"
	}
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return fmt.Sprintf("data: %s\n\n", data)
	}
	if _, ok := chunk["id"]; !ok {
		chunk["id"] = fmt.Sprintf("chatcmpl-%s", newShortID())
	}
	chunk["model"] = model
	chunk["object"] = "chat.completion.chunk"
	b, _ := json.Marshal(chunk)
	return fmt.Sprintf("data: %s\n\n", b)
}

func newShortID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1e12)
}

func ResolveModel(model string, accountDefault string, systemDefault string) string {
	if model != "" {
		return model
	}
	if accountDefault != "" {
		return accountDefault
	}
	if systemDefault != "" {
		return systemDefault
	}
	return agnescode.DefaultModel
}