package tools

import (
	"context"

	"go-nc-mcp-web/internal/mcp"
)

func registerWebDAVTools(srv *mcp.Server) {
	srv.RegisterTool(mcp.Tool{
		Name:        "nc_files_list",
		Description: "List files and directories at a path",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"path": prop("path", "string", "Directory path (e.g. '/' or '/Documents')"),
			},
			Required: []string{"path"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		path, _ := requireString(args, "path")
		files, err := client(ctx).FilesList(path)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(files)
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_files_read",
		Description: "Read the content of a file",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"path": prop("path", "string", "File path")},
			Required:   []string{"path"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		path, _ := requireString(args, "path")
		data, err := client(ctx).FilesRead(path)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult(string(data))
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_files_write",
		Description: "Write content to a file (creates or overwrites)",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"path":    prop("path", "string", "File path"),
				"content": prop("content", "string", "File content"),
			},
			Required: []string{"path", "content"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		path, _ := requireString(args, "path")
		content, _ := requireString(args, "content")
		if err := client(ctx).FilesWrite(path, []byte(content)); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("file written")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_files_mkdir",
		Description: "Create a directory",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"path": prop("path", "string", "Directory path to create")},
			Required:   []string{"path"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		path, _ := requireString(args, "path")
		if err := client(ctx).FilesMkdir(path); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("directory created")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_files_delete",
		Description: "Delete a file or directory",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"path": prop("path", "string", "Path to delete")},
			Required:   []string{"path"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		path, _ := requireString(args, "path")
		if err := client(ctx).FilesDelete(path); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("deleted")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_files_move",
		Description: "Move or rename a file/directory",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"source":      prop("source", "string", "Source path"),
				"destination": prop("destination", "string", "Destination path"),
			},
			Required: []string{"source", "destination"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		src, _ := requireString(args, "source")
		dst, _ := requireString(args, "destination")
		if err := client(ctx).FilesMove(src, dst); err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return mcp.TextResult("moved")
	})

	srv.RegisterTool(mcp.Tool{
		Name:        "nc_files_search",
		Description: "Search for files by name",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.PropertySchema{"pattern": prop("pattern", "string", "Search term (substring match)")},
			Required:   []string{"pattern"},
		},
	}, func(ctx context.Context, args map[string]interface{}) *mcp.ToolResult {
		pattern, _ := requireString(args, "pattern")
		files, err := client(ctx).FilesSearch(pattern)
		if err != nil {
			return mcp.ErrorResult(err.Error())
		}
		return jsonResult(files)
	})
}
