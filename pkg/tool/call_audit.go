package tool

import (
	"crypto/rand"

	"gitea.com/gitea/gitea-mcp/pkg/log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

type callAudit struct {
	id string
}

func beginCallAudit(name string) *callAudit {
	if name != "actions_run_read" {
		return nil
	}
	audit := &callAudit{id: rand.Text()}
	log.Info("tool_call_started", zap.String("tool", name), zap.String("call_id", audit.id))
	return audit
}

func (a *callAudit) complete(result *mcp.CallToolResult, err error) {
	if a == nil {
		return
	}
	outcome := "success"
	if err != nil || result == nil || result.IsError {
		outcome = "error"
	}
	log.Info("tool_call_completed", zap.String("tool", "actions_run_read"), zap.String("call_id", a.id), zap.String("outcome", outcome))
}
