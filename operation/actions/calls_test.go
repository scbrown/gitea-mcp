package actions

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"gitea.com/gitea/gitea-mcp/pkg/flag"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func callText(result *mcp.CallToolResult) string {
	var text strings.Builder
	for _, content := range result.Content {
		if c, ok := content.(*mcp.TextContent); ok {
			text.WriteString(c.Text)
		}
	}
	return text.String()
}

func TestForgejoCallShapesAndStatusAlias(t *testing.T) {
	if err := ConfigureProfile("forgejo-15"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ConfigureProfile("gitea") })
	var calls atomic.Int32
	var runs atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/version":
			fmt.Fprint(w, `{"version":"1.24.6"}`)
		case "/api/v1/repos/example/demo/actions/runs":
			if r.URL.Query().Get("status") != "running" {
				t.Errorf("status=%q", r.URL.Query().Get("status"))
			}
			runs.Add(1)
			fmt.Fprint(w, `{"total_count":1,"workflow_runs":[{"id":42,"status":"running"}]}`)
		default:
			t.Errorf("unexpected HTTP %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	oldHost := flag.Host
	flag.Host = server.URL
	t.Cleanup(func() { flag.Host = oldHost })
	for _, args := range []map[string]any{{}, {"method": "get_run", "owner": "example", "repo": "demo"}, {"method": "typo"}, {"method": "jobs"}} {
		result, err := runReadFn(context.Background(), args)
		if err != nil || !result.IsError {
			t.Fatalf("args=%v result=%v err=%v", args, result, err)
		}
		text := callText(result)
		if !strings.Contains(text, "No request sent") {
			t.Fatalf("missing preHTTP advice: %s", text)
		}
		if _, ok := args["method"]; !ok && (!strings.Contains(text, "Valid methods: list_runs, get_run") || !strings.Contains(text, "Example:")) {
			t.Fatal(text)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("invalid/unsupported calls sent %d HTTP requests", calls.Load())
	}
	for _, method := range []string{"list", "list-runs", "list_runs"} {
		args := map[string]any{"method": method, "owner": "example", "repo": "demo", "status": "in_progress"}
		result, err := runReadFn(context.Background(), args)
		if err != nil || result.IsError {
			t.Fatalf("%s result=%v err=%v", method, result, err)
		}
		if args["method"] != method || args["status"] != "in_progress" {
			t.Fatal("caller arguments mutated")
		}
	}
	if runs.Load() != 3 {
		t.Fatalf("positive controls=%d", runs.Load())
	}
	result, err := runWriteFn(context.Background(), map[string]any{"method": "dispatch", "owner": "example", "repo": "demo"})
	if err != nil || !result.IsError || !strings.Contains(callText(result), "workflow_id is required") || !strings.Contains(callText(result), "dispatch_workflow") {
		t.Fatalf("dispatch guidance result=%v err=%v", result, err)
	}
	if runs.Load() != 3 {
		t.Fatal("incomplete dispatch changed fixture")
	}
}
