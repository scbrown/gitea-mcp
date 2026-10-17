package actions

import (
	"fmt"
	"slices"

	"gitea.com/gitea/gitea-mcp/pkg/tool"
)

// These opt-in capabilities match the published Forgejo 15 Actions API. A
// Gitea-compatible version suffix does not imply Gitea Actions route parity.
var (
	forgejo15Methods = []string{"list_runs", "get_run", "dispatch_workflow"}
	profileName      = "gitea"
)

var (
	originalReadDescription  = ActionsRunReadTool.Description
	originalWriteDescription = ActionsRunWriteTool.Description
	originalReadMethods      = profileMethods(ActionsRunReadTool.InputSchema)
	originalWriteMethods     = profileMethods(ActionsRunWriteTool.InputSchema)
)

func profileMethods(schema any) []string {
	return slices.Clone(schema.(map[string]any)["properties"].(map[string]any)["method"].(map[string]any)["enum"].([]string))
}

// ConfigureProfile selects a capability set before tool registration. Call it
// once at startup, before serving requests. The default retains Gitea behavior.
func ConfigureProfile(name string) error {
	if name != "gitea" && name != "forgejo-15" {
		return fmt.Errorf("unknown Actions API profile %q; choose gitea or forgejo-15", name)
	}
	profileName = name
	for _, entry := range []struct {
		schema  any
		methods []string
	}{{ActionsRunReadTool.InputSchema, originalReadMethods}, {ActionsRunWriteTool.InputSchema, originalWriteMethods}} {
		methods := slices.Clone(entry.methods)
		if name == "forgejo-15" {
			methods = slices.DeleteFunc(methods, func(method string) bool { return !slices.Contains(forgejo15Methods, method) })
		}
		entry.schema.(map[string]any)["properties"].(map[string]any)["method"].(map[string]any)["enum"] = methods
	}
	ActionsRunReadTool.Description = originalReadDescription
	ActionsRunWriteTool.Description = originalWriteDescription
	if name == "forgejo-15" {
		ActionsRunReadTool.Description = tool.MethodDescription("Read Forgejo 15 Actions runs. Status uses running, not in_progress. Job/log/artifact methods are unavailable.", []tool.MethodSpec{tool.Method("list_runs"), tool.Method("get_run", "run_id")})
		ActionsRunWriteTool.Description = tool.MethodDescription("Dispatch Forgejo 15 Actions workflows. Supply workflow-defined inputs; cancel and rerun are unavailable.", []tool.MethodSpec{tool.Method("dispatch_workflow", "workflow_id", "ref")})
	}
	return nil
}

func checkProfileMethod(method string) error {
	if profileName == "forgejo-15" && !slices.Contains(forgejo15Methods, method) {
		return fmt.Errorf("actions method %q is unavailable in the forgejo-15 API profile; consult the instance's supported Actions UI or task/log interface", method)
	}
	return nil
}
