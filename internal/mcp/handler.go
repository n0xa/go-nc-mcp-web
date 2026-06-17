package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// ToolHandler processes a tool call. The context carries the authenticated client.
type ToolHandler func(ctx context.Context, args map[string]interface{}) *ToolResult

type Server struct {
	tools     []Tool
	handlers  map[string]ToolHandler
	mu        sync.RWMutex
	sessionID string
}

func NewServer() *Server {
	id := make([]byte, 16)
	rand.Read(id)
	return &Server{
		handlers:  make(map[string]ToolHandler),
		sessionID: hex.EncodeToString(id),
	}
}

func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, tool)
	s.handlers[tool.Name] = handler
}

// ServeHTTP implements the MCP Streamable HTTP transport.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	switch r.Method {
	case http.MethodGet:
		// SSE stream endpoint per MCP Streamable HTTP spec.
		// Open an SSE stream that stays alive for server-initiated messages.
		// For now, we hold the connection open briefly then close since we
		// don't have server-initiated notifications.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Mcp-Session-Id", s.sessionID)
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// Hold connection until client disconnects
		<-r.Context().Done()

	case http.MethodPost:
		s.handlePost(w, r)

	case http.MethodDelete:
		// Session termination
		w.Header().Set("Mcp-Session-Id", s.sessionID)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, s.sessionID, &Response{
			JSONRPC: "2.0",
			ID:      json.RawMessage("null"),
			Error:   &RPCError{Code: -32700, Message: "parse error"},
		})
		return
	}

	log.Printf("JSON-RPC method: %s", req.Method)

	resp := s.handle(r.Context(), &req)
	if resp == nil {
		w.Header().Set("Mcp-Session-Id", s.sessionID)
		w.WriteHeader(http.StatusAccepted)
		return
	}

	writeJSON(w, http.StatusOK, s.sessionID, resp)
}

func (s *Server) handle(ctx context.Context, req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.respond(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "go-nc-mcp-web",
				"version": "0.1.0",
			},
		})

	case "notifications/initialized":
		return nil

	case "tools/list":
		s.mu.RLock()
		tools := make([]Tool, len(s.tools))
		copy(tools, s.tools)
		s.mu.RUnlock()
		return s.respond(req.ID, map[string]any{"tools": tools})

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.respondError(req.ID, -32602, "invalid params")
		}
		s.mu.RLock()
		handler, ok := s.handlers[params.Name]
		s.mu.RUnlock()
		if !ok {
			return s.respond(req.ID, ErrorResult(fmt.Sprintf("unknown tool: %s", params.Name)))
		}
		result := handler(ctx, params.Arguments)
		return s.respond(req.ID, result)

	default:
		return s.respondError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *Server) respond(id json.RawMessage, result any) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Result: result}
}

func (s *Server) respondError(id json.RawMessage, code int, msg string) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}

func writeJSON(w http.ResponseWriter, status int, sessionID string, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Mcp-Session-Id", sessionID)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}
