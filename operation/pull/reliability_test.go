package pull

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitea.com/gitea/gitea-mcp/pkg/flag"
)

func TestUpdateBranchNoOpIsExactAndNeverRetried(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError bool
		fragment  string
	}{
		{"no-op", 500, `{"message":"HeadBranch of PR 7 is up to date"}`, false, `"changed":false`},
		{"updated", 200, `{}`, false, `"changed":true`},
		{"other-pr", 500, `{"message":"HeadBranch of PR 8 is up to date"}`, true, "PR 8"},
		{"embedded-text", 500, `{"message":"fatal: HeadBranch of PR 7 is up to date"}`, true, "fatal:"},
		{"non-json", 500, `HeadBranch of PR 7 is up to date`, true, "status 500"},
		{"conflict", 409, `{"message":"HeadBranch of PR 7 is up to date"}`, true, "status 409"},
		{"git-stderr", 500, `{"message":"exit status 128: fixture lock exists"}`, true, "fixture lock exists"},
		{"gateway", 503, `{"message":"temporarily unavailable"}`, true, "status 503"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/api/v1/repos/example/demo/pulls/7/update" {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			old := flag.Host
			flag.Host = srv.URL
			defer func() { flag.Host = old }()
			result, err := updatePullRequestBranchFn(context.Background(), map[string]any{"owner": "example", "repo": "demo", "pull_number": float64(7)})
			if err != nil || result.IsError != tc.wantError || !strings.Contains(methodText(result), tc.fragment) {
				t.Fatalf("result=%+v text=%s err=%v", result, methodText(result), err)
			}
			if calls != 1 {
				t.Fatalf("writes=%d want1", calls)
			}
		})
	}
}

func TestPullReadRejectsEmptyOrMismatchedObjects(t *testing.T) {
	for _, body := range []string{`null`, `{}`, `{"number":8}`, `{"number":7,"title":"fixture"}`} {
		for _, method := range []string{"get", "get_status"} {
			t.Run(method+body, func(t *testing.T) {
				assets := 0
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					switch r.URL.Path {
					case "/api/v1/version":
						fmt.Fprint(w, `{"version":"1.24.6"}`)
					case "/api/v1/repos/example/demo":
						fmt.Fprint(w, `{"private":false}`)
					case "/api/v1/repos/example/demo/pulls/7":
						fmt.Fprint(w, body)
					case "/api/v1/repos/example/demo/issues/7/assets":
						assets++
						fmt.Fprint(w, `[]`)
					default:
						t.Errorf("unexpected %s", r.URL.Path)
						http.NotFound(w, r)
					}
				}))
				defer srv.Close()
				old := flag.Host
				flag.Host = srv.URL
				defer func() { flag.Host = old }()
				result, err := pullRequestReadFn(context.Background(), map[string]any{"method": method, "owner": "example", "repo": "demo", "pull_number": float64(7)})
				valid := body == `{"number":7,"title":"fixture"}` && method == "get"
				if err != nil || result.IsError == valid {
					t.Fatalf("valid=%v text=%s err=%v", valid, methodText(result), err)
				}
				if valid && assets != 1 {
					t.Fatalf("positive control assets=%d", assets)
				}
				if !valid && assets != 0 {
					t.Fatalf("invalid PR fetched assets %d", assets)
				}
			})
		}
	}
}
