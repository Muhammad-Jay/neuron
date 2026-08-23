package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/project"
)

var variablePattern = regexp.MustCompile(`\$\{([^}]+)\}`)

type VariableResolver struct {
	variables map[string]any
}

func NewVariableResolver(project *project.ProjectFile) *VariableResolver {
	vars := make(map[string]any)

	if project.Metadata.Name != "" {
		vars["project.name"] = project.Metadata.Name
	}
	if project.Metadata.Version != "" {
		vars["project.version"] = project.Metadata.Version
	}
	if project.Metadata.Description != "" {
		vars["project.description"] = project.Metadata.Description
	}

	flattenVariables("", project.Variables, vars)

	if project.Runtime.Execution.DefaultMode != "" {
		vars["runtime.execution.defaultMode"] = project.Runtime.Execution.DefaultMode
	}
	if project.Runtime.Execution.Timeout != "" {
		vars["runtime.execution.timeout"] = project.Runtime.Execution.Timeout
	}
	if project.Runtime.Workers.Min > 0 {
		vars["runtime.workers.min"] = project.Runtime.Workers.Min
	}
	if project.Runtime.Workers.Max > 0 {
		vars["runtime.workers.max"] = project.Runtime.Workers.Max
	}

	if project.Storage.Provider != "" {
		vars["storage.provider"] = project.Storage.Provider
	}
	if project.Storage.Local.Directory != "" {
		vars["storage.local.directory"] = project.Storage.Local.Directory
	}

	for _, src := range project.Executors.Sources {
		if src.Name != "" {
			vars["executors.sources."+src.Name+".type"] = src.Type
			vars["executors.sources."+src.Name+".url"] = src.URL
		}
	}

	if project.Inspector.Enabled {
		vars["inspector.enabled"] = project.Inspector.Enabled
	}
	if project.Inspector.Address != "" {
		vars["inspector.address"] = project.Inspector.Address
	}

	return &VariableResolver{variables: vars}
}

func (v *VariableResolver) Resolve(value any) any {
	switch typed := value.(type) {
	case string:
		return v.resolveString(typed)
	case map[string]any:
		result := make(map[string]any)
		for k, val := range typed {
			result[k] = v.Resolve(val)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, val := range typed {
			result[i] = v.Resolve(val)
		}
		return result
	default:
		return typed
	}
}

func (v *VariableResolver) ResolveString(s string) string {
	return v.resolveString(s)
}

func (v *VariableResolver) resolveString(s string) string {
	return variablePattern.ReplaceAllStringFunc(s, func(match string) string {
		key := match[2 : len(match)-1]
		key = strings.TrimSpace(key)

		if strings.HasPrefix(key, "env.") {
			envKey := key[4:]
			if val, ok := os.LookupEnv(envKey); ok {
				return val
			}
			return ""
		}

		if strings.HasPrefix(key, "secret.") {
			return "${secret." + key[7:] + "}"
		}

		if val, ok := v.variables[key]; ok {
			return toString(val)
		}

		return match
	})
}

func flattenVariables(prefix string, vars map[string]any, out map[string]any) {
	for k, val := range vars {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch typed := val.(type) {
		case map[string]any:
			flattenVariables(key, typed, out)
		default:
			out[key] = val
		}
	}
}

func toString(val any) string {
	switch typed := val.(type) {
	case string:
		return typed
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case float64:
		return fmt.Sprintf("%g", typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", typed)
	}
}