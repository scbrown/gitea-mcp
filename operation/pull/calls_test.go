package pull

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

func methodText(result *mcp.CallToolResult) string {
	var text strings.Builder
	for _, content := range result.Content {
		if c, ok := content.(*mcp.TextContent); ok {
			text.WriteString(c.Text)
		}
	}
	return text.String()
}

func TestPullAliasesDoNotGuessNumbersOrSendIncompleteWrites(t *testing.T) {
	var calls atomic.Int32
	var lists atomic.Int32
	var edits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/version":
			fmt.Fprint(w, `{"version":"1.24.6"}`)
		case "/api/v1/repos/example/demo":
			fmt.Fprint(w, `{"private":false}`)
		case "/api/v1/repos/example/demo/pulls":
			if r.Method != http.MethodGet {
				t.Errorf("listing method=%s", r.Method)
			}
			lists.Add(1)
			fmt.Fprint(w, `[{"number":7,"title":"fixture","state":"open"}]`)
		case "/api/v1/repos/example/demo/pulls/7":
			if r.Method != http.MethodPatch {
				t.Errorf("edit method=%s", r.Method)
			}
			edits.Add(1)
			fmt.Fprint(w, `{"number":7,"title":"updated","state":"open"}`)
		default:
			t.Errorf("unexpected HTTP %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	oldHost := flag.Host
	flag.Host = server.URL
	t.Cleanup(func() { flag.Host = oldHost })
	for _, method := range []string{"get", "get_review"} {
		result, err := pullRequestReadFn(context.Background(), map[string]any{"method": method, "owner": "example", "repo": "demo"})
		if err != nil || !result.IsError || !strings.Contains(methodText(result), "pull_number is required") || !strings.Contains(methodText(result), "Valid methods:") || !strings.Contains(methodText(result), "Example:") {
			t.Fatalf("result=%v err=%v", result, err)
		}
	}
	result, err := pullRequestWriteFn(context.Background(), map[string]any{"method": "edit", "owner": "example", "repo": "demo"})
	if err != nil || !result.IsError || !strings.Contains(methodText(result), "pull_number is required") {
		t.Fatalf("incomplete edit result=%v err=%v", result, err)
	}
	if calls.Load() != 0 {
		t.Fatalf("invalid calls sent %d HTTP requests", calls.Load())
	}
	for _, method := range []string{"list", "list_pull_requests"} {
		args := map[string]any{"method": method, "owner": "example", "repo": "demo"}
		result, err = pullRequestReadFn(context.Background(), args)
		if err != nil || result.IsError || !strings.Contains(methodText(result), "fixture") {
			t.Fatalf("listing result=%v err=%v", result, err)
		}
		if args["method"] != method {
			t.Fatal("caller arguments mutated")
		}
	}
	result, err = pullRequestWriteFn(context.Background(), map[string]any{"method": "edit", "owner": "example", "repo": "demo", "pull_number": float64(7), "title": "updated"})
	if err != nil || result.IsError {
		t.Fatalf("edit result=%v err=%v", result, err)
	}
	if lists.Load() != 2 || edits.Load() != 1 {
		t.Fatalf("lists=%d edits=%d", lists.Load(), edits.Load())
	}
}
