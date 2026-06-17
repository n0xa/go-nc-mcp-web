package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"go-nc-mcp-web/internal/auth"
	"go-nc-mcp-web/internal/mcp"
	"go-nc-mcp-web/internal/nextcloud"
)

func client(ctx context.Context) *nextcloud.Client {
	return auth.ClientFromContext(ctx)
}

func jsonResult(v interface{}) *mcp.ToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.ErrorResult(fmt.Sprintf("marshal error: %v", err))
	}
	return mcp.TextResult(string(data))
}

func requireInt(args map[string]interface{}, key string) (int, error) {
	v, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case json.Number:
		i, err := n.Int64()
		return int(i), err
	default:
		return 0, fmt.Errorf("parameter %s must be a number", key)
	}
}

func requireString(args map[string]interface{}, key string) (string, error) {
	v, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	return s, nil
}

func optString(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func optInt(args map[string]interface{}, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return def
	}
}

func prop(name, typ, desc string) mcp.PropertySchema {
	return mcp.PropertySchema{Type: typ, Description: desc}
}
