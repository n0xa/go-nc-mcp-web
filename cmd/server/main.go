package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"go-nc-mcp-web/internal/auth"
	"go-nc-mcp-web/internal/mcp"
	"go-nc-mcp-web/internal/tools"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	loadEnvFile("/etc/ncmcp.env")

	ncHost := os.Getenv("NEXTCLOUD_HOST")
	listen := os.Getenv("MCP_LISTEN")
	mcpHostname := os.Getenv("MCP_HOSTNAME")

	if ncHost == "" {
		fmt.Fprintln(os.Stderr, "error: NEXTCLOUD_HOST must be set")
		os.Exit(1)
	}
	if listen == "" {
		listen = ":8080"
	}
	if mcpHostname == "" {
		mcpHostname = "localhost"
	}

	srv := mcp.NewServer()
	tools.RegisterTools(srv)

	mcpHandler := auth.NextcloudAppToken(ncHost, srv)

	// OAuth server for Claude connector authentication
	issuer := "https://" + mcpHostname
	oauth := auth.NewOAuthServer(issuer)

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", oauth.Metadata)
	mux.HandleFunc("/authorize", oauth.Authorize)
	mux.HandleFunc("/token", oauth.Token)
	mux.Handle("/mcp", corsMiddleware(mcpHandler))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("unhandled: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		http.NotFound(w, r)
	})

	log.Printf("go-nc-mcp-web listening on %s (authenticating against %s)", listen, ncHost)
	if err := http.ListenAndServe(listen, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok {
			os.Setenv(k, v)
		}
	}
}
