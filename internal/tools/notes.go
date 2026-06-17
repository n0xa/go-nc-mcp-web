package tools

import (
	"context"

	"go-nc-mcp-web/internal/mcp"
)

func registerNotesTools(srv *mcp.Server) {
	srv.RegisterTool(mcp.Tool{
		Name:        "nc_notes_search",
		Description: "Search notes by query string (returns titles, not content)",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"query": prop("query", "string", "Search query (empty lists all)"),
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		notes, err := client(ctx).NotesSearch(optString(args, "query"))
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(notes)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_notes_get",
		Description: "Get a note by ID (includes full content)",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"id": prop("id", "integer", "Note ID")},
			Required:   []string{"id"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		id, err := requireInt(args, "id")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		note, err := client(ctx).NotesGet(id)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(note)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_notes_create",
		Description: "Create a new note",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"title":    prop("title", "string", "Note title"),
				"content":  prop("content", "string", "Note content (Markdown)"),
				"category": prop("category", "string", "Category/folder"),
			},
			Required: []string{"title", "content"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		title, err := requireString(args, "title")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		content, err := requireString(args, "content")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		note, err := client(ctx).NotesCreate(title, content, optString(args, "category"))
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(note)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_notes_update",
		Description: "Update a note (title, content, category)",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"id":       prop("id", "integer", "Note ID"),
				"title":    prop("title", "string", "New title"),
				"content":  prop("content", "string", "New content"),
				"category": prop("category", "string", "New category"),
			},
			Required: []string{"id"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		id, err := requireInt(args, "id")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		fields := make(map[string]interface{})
		if v, ok := args["title"]; ok {
			fields["title"] = v
		}
		if v, ok := args["content"]; ok {
			fields["content"] = v
		}
		if v, ok := args["category"]; ok {
			fields["category"] = v
		}
		note, err := client(ctx).NotesUpdate(id, fields)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(note)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_notes_delete",
		Description: "Delete a note",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"id": prop("id", "integer", "Note ID")},
			Required:   []string{"id"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		id, err := requireInt(args, "id")
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		if err := client(ctx).NotesDelete(id); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("note deleted")
	})
}
