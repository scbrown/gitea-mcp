package tool

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MethodSpec describes parameters required by one multiplexed tool method.
type MethodSpec struct {
	Name   string
	Fields []string
}

// Method defines a method's additional requirements beyond owner and repo.
func Method(name string, fields ...string) MethodSpec {
	return MethodSpec{Name: name, Fields: fields}
}

func methodExample(spec MethodSpec) string {
	args := map[string]any{"method": spec.Name, "owner": "example", "repo": "demo"}
	for _, field := range spec.Fields {
		switch field {
		case "workflow_id":
			args[field] = "ci.yml"
		case "ref", "base":
			args[field] = "main"
		case "head":
			args[field] = "feature"
		case "title":
			args[field] = "Example change"
		default:
			args[field] = 1
		}
	}
	encoded, _ := json.Marshal(args)
	return string(encoded)
}

// MethodDescription adds an example and required parameters for every method.
func MethodDescription(summary string, specs []MethodSpec) string {
	var out strings.Builder
	out.WriteString(summary)
	for _, spec := range specs {
		fmt.Fprintf(&out, "\n%s requires %s; example %s", spec.Name,
			strings.Join(append([]string{"owner", "repo"}, spec.Fields...), ", "), methodExample(spec))
	}
	return out.String()
}

// ResolveMethodAliases copies arguments so compatibility names never mutate a caller's map.
func ResolveMethodAliases(args map[string]any, aliases map[string]string) map[string]any {
	copyArgs := maps.Clone(args)
	if method, ok := args["method"].(string); ok {
		if canonical, found := aliases[method]; found {
			copyArgs["method"] = canonical
		}
	}
	return copyArgs
}

// MethodCallError supplies advice from the current discovery enum, including API profiles.
func MethodCallError(def *mcp.Tool, args map[string]any, specs []MethodSpec, reason string) error {
	schema := def.InputSchema.(map[string]any)
	methods := schema["properties"].(map[string]any)["method"].(map[string]any)["enum"].([]string)
	method, _ := args["method"].(string)
	selected := MethodSpec{}
	for _, spec := range specs {
		if !slices.Contains(methods, spec.Name) {
			continue
		}
		if selected.Name == "" || spec.Name == method {
			selected = spec
		}
		if spec.Name == method {
			break
		}
	}
	return fmt.Errorf("%s: %s. Valid methods: %s. Required for %s: %s. Example: %s. No request sent",
		def.Name, reason, strings.Join(methods, ", "), selected.Name,
		strings.Join(append([]string{"owner", "repo"}, selected.Fields...), ", "), methodExample(selected))
}

// ValidateMethodCall reports actionable errors before a multiplexed tool sends HTTP.
func ValidateMethodCall(def *mcp.Tool, args map[string]any, specs []MethodSpec) (string, error) {
	schema := def.InputSchema.(map[string]any)
	methods := schema["properties"].(map[string]any)["method"].(map[string]any)["enum"].([]string)
	method, _ := args["method"].(string)
	fail := func(reason string) (string, error) { return "", MethodCallError(def, args, specs, reason) }
	if method == "" {
		return fail("method is required")
	}
	if !slices.Contains(methods, method) {
		return fail(fmt.Sprintf("unknown method %q", method))
	}
	var fields []string
	for _, spec := range specs {
		if spec.Name == method {
			fields = spec.Fields
			break
		}
	}
	for _, field := range append([]string{"owner", "repo"}, fields...) {
		value, found := args[field]
		if !found || value == nil || value == "" {
			return fail(field + " is required")
		}
	}
	return method, nil
}
