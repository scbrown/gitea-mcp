package tool

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"gitea.com/gitea/gitea-mcp/pkg/log"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func observeCallAudit(t *testing.T) *observer.ObservedLogs {
	t.Helper()
	previous := log.Default()
	core, entries := observer.New(zap.InfoLevel)
	log.SetDefault(zap.New(core))
	t.Cleanup(func() { log.SetDefault(previous) })
	return entries
}

func TestActionsReadCallAudit(t *testing.T) {
	entries := observeCallAudit(t)
	success := func(context.Context, map[string]any) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{}, nil
	}
	for _, test := range []struct {
		name      string
		handler   Handler
		arguments string
		want      string
	}{
		{"success", success, `{"token":"fixture-sensitive"}`, "success"},
		{"tool error", func(context.Context, map[string]any) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{IsError: true}, nil
		}, `{}`, "error"},
		{"handler error", func(context.Context, map[string]any) (*mcp.CallToolResult, error) {
			return nil, errors.New("fixture-sensitive")
		}, `{}`, "error"},
		{"protocol error", func(context.Context, map[string]any) (*mcp.CallToolResult, error) {
			return nil, &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "fixture-sensitive"}
		}, `{}`, "error"},
		{"panic", func(context.Context, map[string]any) (*mcp.CallToolResult, error) { panic("fixture-sensitive") }, `{}`, "error"},
		{"decode error", success, `[]`, "error"},
		{"nil result", func(context.Context, map[string]any) (*mcp.CallToolResult, error) {
			return nil, nil //nolint:nilnil // Deliberately malformed handler result must be audited as an error.
		}, `{}`, "error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			entries.TakeAll()
			st := ServerTool{Tool: &mcp.Tool{Name: "actions_run_read"}, Handler: test.handler}
			_, _ = st.MCPHandler()(context.Background(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: json.RawMessage(test.arguments)}})
			records := entries.Filter(func(entry observer.LoggedEntry) bool {
				return entry.Message == "tool_call_started" || entry.Message == "tool_call_completed"
			}).All()
			entries.TakeAll()
			if len(records) != 2 || records[0].Message != "tool_call_started" || records[1].Message != "tool_call_completed" {
				t.Fatalf("want one ordered audit pair, got %#v", records)
			}
			start, end := records[0].ContextMap(), records[1].ContextMap()
			if start["call_id"] == "" || start["call_id"] != end["call_id"] || end["outcome"] != test.want {
				t.Fatalf("incorrect correlation/outcome: %#v %#v", start, end)
			}
			if len(start) != 2 || len(end) != 3 || start["tool"] != "actions_run_read" || end["tool"] != "actions_run_read" {
				t.Fatalf("audit must contain only bounded tool, opaque ID and outcome: %#v %#v", start, end)
			}
		})
	}
	other := ServerTool{Tool: &mcp.Tool{Name: "other"}, Handler: success}
	_, _ = other.MCPHandler()(context.Background(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{}})
	if entries.Len() != 0 {
		t.Fatal("unrelated tool was audited")
	}
}

func TestActionsReadAuditTracksUnfinishedAndConcurrentCalls(t *testing.T) {
	entries := observeCallAudit(t)
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	st := ServerTool{Tool: &mcp.Tool{Name: "actions_run_read"}, Handler: func(context.Context, map[string]any) (*mcp.CallToolResult, error) {
		close(entered)
		<-release
		return &mcp.CallToolResult{}, nil
	}}
	go func() {
		defer close(done)
		_, _ = st.MCPHandler()(context.Background(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{}})
	}()
	<-entered
	if entries.Len() != 1 || entries.All()[0].Message != "tool_call_started" {
		t.Error("unfinished call must retain its start without a completion")
	}
	close(release)
	<-done
	entries.TakeAll()
	st.Handler = func(context.Context, map[string]any) (*mcp.CallToolResult, error) { return &mcp.CallToolResult{}, nil }
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			_, _ = st.MCPHandler()(context.Background(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{}})
		})
	}
	wg.Wait()
	pairs := map[string]int{}
	for _, record := range entries.All() {
		pairs[record.ContextMap()["call_id"].(string)]++
	}
	if len(pairs) != 16 {
		t.Fatalf("want 16 unique concurrent calls, got %d", len(pairs))
	}
	for id, count := range pairs {
		if count != 2 {
			t.Errorf("call %s has %d records, want 2", id, count)
		}
	}
}
