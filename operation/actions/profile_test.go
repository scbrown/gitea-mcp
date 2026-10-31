package actions

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"gitea.com/gitea/gitea-mcp/pkg/flag"
)

func TestForgejo15DiscoveryAndStaleCalls(t *testing.T) {
	if err := ConfigureProfile("forgejo-15"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ConfigureProfile("gitea") })
	if got := profileMethods(ActionsRunReadTool.InputSchema); !reflect.DeepEqual(got, []string{"list_runs", "get_run"}) {
		t.Fatalf("read methods = %v", got)
	}
	if got := profileMethods(ActionsRunWriteTool.InputSchema); !reflect.DeepEqual(got, []string{"dispatch_workflow"}) {
		t.Fatalf("write methods = %v", got)
	}

	var calls atomic.Int32
	wantPage := "1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/version":
			fmt.Fprint(w, `{"version":"1.24.6"}`)
		case "/api/v1/repos/demo/repo/actions/runs":
			t.Logf("Actions list request: %s?%s", r.URL.Path, r.URL.RawQuery)
			if r.URL.Query().Get("page") != wantPage || r.URL.Query().Get("limit") != "30" {
				t.Errorf("pagination = %s, want page=%s and limit=30", r.URL.RawQuery, wantPage)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, `{"total_count":1,"workflow_runs":[{"id":42,"status":"success"}]}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	oldHost := flag.Host
	flag.Host = server.URL
	t.Cleanup(func() { flag.Host = oldHost })
	for _, method := range originalReadMethods {
		if method == "list_runs" || method == "get_run" {
			continue
		}
		result, err := runReadFn(context.Background(), map[string]any{"method": method})
		if err != nil || !result.IsError {
			t.Fatalf("%s result=%v err=%v", method, result, err)
		}
	}
	for _, method := range []string{"rerun_run", "cancel_run"} {
		result, err := runWriteFn(context.Background(), map[string]any{"method": method})
		if err != nil || !result.IsError {
			t.Fatalf("%s result=%v err=%v", method, result, err)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("stale methods sent %d requests", calls.Load())
	}
	result, err := runReadFn(context.Background(), map[string]any{"method": "list_runs", "owner": "demo", "repo": "repo"})
	if err != nil || result.IsError {
		t.Fatalf("supported positive control result=%v err=%v", result, err)
	}
	if calls.Load() == 0 {
		t.Fatal("positive control sent no request")
	}
	wantPage = "2"
	for _, page := range []any{float64(2), "2"} {
		result, err = runReadFn(context.Background(), map[string]any{"method": "list_runs", "owner": "demo", "repo": "repo", "page": page})
		if err != nil || result.IsError {
			t.Fatalf("positive page %v result=%v err=%v", page, result, err)
		}
	}
	for _, page := range []any{float64(0), float64(-1), float64(1.5), "0", "-1", "bad", "1.5", "9223372036854775808", true, nil, math.NaN(), math.Inf(1), math.Ldexp(1, strconv.IntSize-1)} {
		before := calls.Load()
		result, err = runReadFn(context.Background(), map[string]any{"method": "list_runs", "owner": "demo", "repo": "repo", "page": page})
		if err != nil || !result.IsError {
			t.Fatalf("invalid page %v result=%v err=%v", page, result, err)
		}
		if calls.Load() != before {
			t.Fatalf("invalid page %v sent HTTP", page)
		}
		t.Logf("Invalid page %v: rejected with zero HTTP requests", page)
	}
}

func TestActionsProfileValidationAndDefault(t *testing.T) {
	if err := ConfigureProfile("gitea"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(profileMethods(ActionsRunReadTool.InputSchema), originalReadMethods) {
		t.Fatal("default read methods changed")
	}
	if !reflect.DeepEqual(profileMethods(ActionsRunWriteTool.InputSchema), originalWriteMethods) {
		t.Fatal("default write methods changed")
	}
	if err := checkProfileMethod("rerun_run"); err != nil {
		t.Fatal(err)
	}
	if err := ConfigureProfile("typo"); err == nil || !strings.Contains(err.Error(), "unknown Actions API profile") {
		t.Fatalf("invalid profile error = %v", err)
	}
	if !reflect.DeepEqual(profileMethods(ActionsRunReadTool.InputSchema), originalReadMethods) {
		t.Fatal("invalid profile mutated discovery")
	}
}
