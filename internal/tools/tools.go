package tools

import "go-nc-mcp-web/internal/mcp"

func RegisterTools(srv *mcp.Server) {
	registerDeckTools(srv)
	registerNotesTools(srv)
	registerCalendarTools(srv)
	registerContactsTools(srv)
	registerWebDAVTools(srv)
}
