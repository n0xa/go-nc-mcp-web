package nextcloud_test

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"go-nc-mcp-web/internal/nextcloud"
)

func loadTestEnv(t *testing.T) *nextcloud.Client {
	t.Helper()

	// Try test.env in project root
	envPath := "../../test.env"
	f, err := os.Open(envPath)
	if err != nil {
		t.Skipf("skipping integration test: %v (create test.env with credentials)", err)
	}
	defer f.Close()

	env := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok {
			env[k] = v
		}
	}

	host := env["NEXTCLOUD_HOST"]
	user := env["NEXTCLOUD_USERNAME"]
	pass := env["NEXTCLOUD_PASSWORD"]

	if host == "" || user == "" || pass == "" {
		t.Skip("skipping: test.env missing required NEXTCLOUD_HOST, NEXTCLOUD_USERNAME, or NEXTCLOUD_PASSWORD")
	}

	return nextcloud.NewClient(host, user, pass)
}
