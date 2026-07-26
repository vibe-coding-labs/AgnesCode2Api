package agnescode

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	IssueCodeURL = "https://api.agnes-ai.com/api/v1/user/issue-authorization-code"
	BFFBaseURL   = "https://api-agnes-code.agnes-ai.com"
	DefaultModel = "agnes-2.0-flash"
	ClientID     = "agnes-code"
	RedirectURI  = "agnes://auth/callback"
)

type Client struct {
	JWTToken    string
	OAuthToken  string
	httpClient  *http.Client
	mu          sync.RWMutex
	modelsCache []ModelInfo
	cacheTime   time.Time
}

func NewClient(jwtToken string) *Client {
	return &Client{
		JWTToken:   jwtToken,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *Client) SetTimeout(d time.Duration) {
	c.httpClient.Timeout = d
}

func (c *Client) SetHTTPClient(hc *http.Client) {
	c.httpClient = hc
}

func (c *Client) Authenticate() error {
	if c.JWTToken == "" {
		return fmt.Errorf("JWT token is required")
	}
	authCode, err := c.issueAuthCode()
	if err != nil {
		return fmt.Errorf("issue auth code: %w", err)
	}
	token, err := c.exchangeToken(authCode)
	if err != nil {
		return fmt.Errorf("exchange token: %w", err)
	}
	c.mu.Lock()
	c.OAuthToken = token
	c.mu.Unlock()
	return nil
}

func (c *Client) issueAuthCode() (string, error) {
	state := newHexID()
	body := map[string]interface{}{
		"redirect_uri": RedirectURI, "state": state,
		"client_id": ClientID, "ttl_seconds": 60,
	}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", IssueCodeURL, bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+c.JWTToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("origin", "https://app.agnes-ai.com")
	req.Header.Set("x-platform", "1")
	req.Header.Set("x-user-language", "en")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r struct {
		Code string `json:"code"`
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return "", fmt.Errorf("parse: %w, %s", err, string(b))
	}
	if r.Code != "000000" {
		return "", fmt.Errorf("issue auth code failed: code=%s %s", r.Code, string(b))
	}
	return r.Data.Code, nil
}

func (c *Client) exchangeToken(authCode string) (string, error) {
	body := map[string]string{
		"code": authCode, "redirect_uri": RedirectURI,
		"state": newHexID(), "client_id": ClientID,
	}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", BFFBaseURL+"/api/v1/code/auth/exchange-code", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r struct {
		Code string `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return "", fmt.Errorf("parse: %w, %s", err, string(b))
	}
	if r.Code != "000000" && r.Code != "0" {
		return "", fmt.Errorf("exchange failed: code=%s %s", r.Code, string(b))
	}
	return r.Data.AccessToken, nil
}

func (c *Client) ensureAuth() error {
	c.mu.RLock()
	tok := c.OAuthToken
	c.mu.RUnlock()
	if tok != "" {
		return nil
	}
	return c.Authenticate()
}

func (c *Client) bearerHeaders() http.Header {
	c.mu.RLock()
	tok := c.OAuthToken
	c.mu.RUnlock()
	h := http.Header{}
	h.Set("Authorization", "Bearer "+tok)
	h.Set("Content-Type", "application/json")
	h.Set("X-User-Language", "zh-Hans")
	return h
}

func (c *Client) bffGet(path string) (map[string]interface{}, error) {
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("GET", BFFBaseURL+path, nil)
	req.Header = c.bearerHeaders()
	if strings.HasPrefix(path, "/v1/models") {
		req.Header.Set("X-App-Id", "1")
		req.Header.Set("X-Platform", "1")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		c.mu.Lock()
		c.OAuthToken = ""
		c.mu.Unlock()
		return nil, fmt.Errorf("token expired")
	}
	b, _ := io.ReadAll(resp.Body)
	var r map[string]interface{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("parse: %w, %s", err, string(b))
	}
	return r, nil
}

func (c *Client) bffPost(path string, body interface{}) (map[string]interface{}, error) {
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", BFFBaseURL+path, bytes.NewReader(data))
	req.Header = c.bearerHeaders()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		c.mu.Lock()
		c.OAuthToken = ""
		c.mu.Unlock()
		return nil, fmt.Errorf("token expired")
	}
	b, _ := io.ReadAll(resp.Body)
	var r map[string]interface{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("parse: %w, %s", err, string(b))
	}
	return r, nil
}

func (c *Client) ListModels() ([]ModelInfo, error) {
	c.mu.RLock()
	if c.modelsCache != nil && time.Since(c.cacheTime) < 5*time.Minute {
		mm := c.modelsCache
		c.mu.RUnlock()
		return mm, nil
	}
	c.mu.RUnlock()
	resp, err := c.bffGet("/v1/models")
	if err != nil {
		return nil, err
	}
	data, ok := resp["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected models response")
	}
	models := make([]ModelInfo, 0, len(data))
	for _, item := range data {
		b, _ := json.Marshal(item)
		var m ModelInfo
		if json.Unmarshal(b, &m) == nil {
			models = append(models, m)
		}
	}
	c.mu.Lock()
	c.modelsCache = models
	c.cacheTime = time.Now()
	c.mu.Unlock()
	return models, nil
}

func (c *Client) ChatCompletion(req ChatRequest) (*ChatResponse, error) {
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}
	data, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", BFFBaseURL+"/v1/chat/completions", bytes.NewReader(data))
	httpReq.Header = c.bearerHeaders()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		c.mu.Lock()
		c.OAuthToken = ""
		c.mu.Unlock()
		return nil, fmt.Errorf("token expired")
	}
	b, _ := io.ReadAll(resp.Body)
	var r ChatResponse
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("parse: %w, %s", err, string(b))
	}
	return &r, nil
}

func (c *Client) ChatCompletionStream(req ChatRequest) (<-chan StreamEvent, error) {
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}
	req.Stream = true
	data, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", BFFBaseURL+"/v1/chat/completions", bytes.NewReader(data))
	httpReq.Header = c.bearerHeaders()
	httpReq.Header.Set("Accept", "text/event-stream")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("stream error: HTTP %d", resp.StatusCode)
	}
	ch := make(chan StreamEvent, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)
		br := bufio.NewReader(resp.Body)
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					slog.Error("stream read error", "error", err)
				}
				return
			}
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				return
			}
			var event StreamEvent
			if json.Unmarshal([]byte(payload), &event) == nil {
				ch <- event
			}
		}
	}()
	return ch, nil
}

func (c *Client) GetBalance() (*Balance, error) {
	resp, err := c.bffGet("/api/v2/subscription/credits-balance")
	if err != nil {
		return nil, err
	}
	data, _ := resp["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("unexpected balance response")
	}
	b, _ := json.Marshal(data)
	var bal Balance
	json.Unmarshal(b, &bal)
	return &bal, nil
}

func (c *Client) GetUserInfo() (*UserInfo, error) {
	resp, err := c.bffGet("/api/v1/user/profile")
	if err != nil {
		return nil, err
	}
	data, _ := resp["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("unexpected user info response")
	}
	ui, _ := data["user_info"].(map[string]interface{})
	if ui == nil {
		return nil, fmt.Errorf("unexpected user info format")
	}
	b, _ := json.Marshal(ui)
	var info UserInfo
	json.Unmarshal(b, &info)
	return &info, nil
}

func (c *Client) GetTransactions(page, pageSize int) (*Transactions, error) {
	resp, err := c.bffPost("/api/v1/subscription/credits-transactions", map[string]interface{}{
		"page": page, "page_size": pageSize, "filter": 0,
	})
	if err != nil {
		return nil, err
	}
	data, _ := resp["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("unexpected transactions response")
	}
	b, _ := json.Marshal(data)
	var txns Transactions
	json.Unmarshal(b, &txns)
	return &txns, nil
}

func newHexID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}