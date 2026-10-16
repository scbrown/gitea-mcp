package actions

import (
	"fmt"
	"slices"
	"strings"
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
		ActionsRunReadTool.Description = "Read Forgejo 15 Actions runs (list_runs, get_run). Job/log/artifact methods are not available in this API profile."
		ActionsRunWriteTool.Description = "Dispatch Forgejo 15 Actions workflows. Cancel and rerun methods are not available in this API profile."
	}
	return nil
}

func checkProfileMethod(method string) error {
	if profileName == "forgejo-15" && !slices.Contains(forgejo15Methods, method) {
		return fmt.Errorf("actions method %q is unavailable in the forgejo-15 API profile; supported methods: %s. No request sent; consult the instance's supported Actions UI or task/log interface", method, strings.Join(forgejo15Methods, ", "))
	}
	return nil
}
