package openai

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/agnes"
	"github.com/vibe-coding-labs/AgnesCode2Api/pkg/store"
)

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("decode chat request", "error", err)
		writeError(w, 400, fmt.Sprintf("Failed to parse request body: %s", err.Error()))
		return
	}
	systemDefault := ""
	if s.store != nil {
		systemDefault = s.store.GetSetting("default_model")
	}
	model := ResolveModel(req.Model, store.GetAccountDefaultModel(r), systemDefault)
	store.SetModel(r, model)
	client := s.getClient(r)
	jcReq := TranslateRequest(&req)
	if req.Stream {
		s.handleStreamChat(w, r, client, &jcReq, model)
	} else {
		s.handleNonStreamChat(w, r, client, &jcReq, model)
	}
}

func (s *Server) handleNonStreamChat(w http.ResponseWriter, r *http.Request, client *agnescode.Client, req *agnescode.ChatRequest, model string) {
	req.Model = model
	resp, err := client.ChatCompletion(*req)
	if err != nil {
		slog.Error("chat non-stream upstream error", "model", model, "error", err)
		msg := err.Error()
		code := 500
		if isTimeoutError(msg) {
			code = 504
			msg = "Upstream timeout, please retry later. " + msg
		}
		writeError(w, code, msg)
		return
	}
	store.SetTokenUsage(r, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	writeJSON(w, 200, TranslateResponse(resp, model))
}

func (s *Server) handleStreamChat(w http.ResponseWriter, r *http.Request, client *agnescode.Client, req *agnescode.ChatRequest, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.Error("streaming not supported by response writer")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "close")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(200)

	req.Model = model
	ch, err := client.ChatCompletionStream(*req)
	if err != nil {
		slog.Error("chat stream upstream error", "model", model, "error", err)
		msg := err.Error()
		if isTimeoutError(msg) {
			msg = "Upstream timeout, please retry later. " + msg
		}
		fmt.Fprintf(w, "data: {\"error\":{\"message\":\"%s\"}}\n\n", msg)
		flusher.Flush()
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}
	for event := range ch {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func isTimeoutError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "context deadline exceeded") ||
		strings.Contains(lower, "client.timeout exceeded") ||
		strings.Contains(lower, "deadline exceeded") ||
		strings.Contains(lower, "i/o timeout")
}