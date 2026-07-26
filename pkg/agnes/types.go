package agnescode

import "encoding/json"

type ModelInfo struct {
	ID              string   `json:"id"`
	OwnedBy         string   `json:"owned_by"`
	Description     string   `json:"description"`
	MaxInputTokens  int      `json:"max_input_tokens"`
	MaxOutputTokens int      `json:"max_output_tokens"`
	IsMemberOnly    bool     `json:"is_member_only"`
	IsGray          bool     `json:"is_gray"`
	ModelType       string   `json:"model_type"`
	Object          string   `json:"object"`
	Provider        string   `json:"provider"`
	EndpointTypes   []string `json:"supported_endpoint_types"`
	Created         int64    `json:"created"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int             `json:"index"`
	FinishReason string          `json:"finish_reason"`
	Message      ResponseMessage `json:"message"`
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamEvent struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        StreamDelta  `json:"delta"`
	FinishReason *string      `json:"finish_reason,omitempty"`
}

type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Balance struct {
	Level                int     `json:"level"`
	LevelName            string  `json:"level_name"`
	TotalBalance         float64 `json:"total_balance"`
	TimeSensitiveBalance float64 `json:"time_sensitive_balance"`
	PermanentBalance     float64 `json:"permanent_balance"`
	DailyFreeCredits     float64 `json:"daily_free_credits"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_url"`
	IsActive bool   `json:"is_active"`
}

type Transactions struct {
	List       []Transaction `json:"list"`
	Pagination Pagination    `json:"pagination"`
}

type Transaction struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Direction   int     `json:"direction"`
	Platform    int     `json:"platform"`
	CreatedAt   int64   `json:"created_at"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

func init() {
	// Ensure json encoding is used
	_ = json.Marshal
}