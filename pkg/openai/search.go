package openai

import (
	"log/slog"
	"net/http"
)

func (s *Server) handleWebSearch(w http.ResponseWriter, r *http.Request) {
	slog.Warn("web search not supported in Agnes API")
	writeError(w, 501, "web search is not supported with Agnes API")
}

func (s *Server) handleRerank(w http.ResponseWriter, r *http.Request) {
	slog.Warn("rerank not supported in Agnes API")
	writeError(w, 501, "rerank is not supported with Agnes API")
}