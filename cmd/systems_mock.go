package main
//
//import "nore/types"
//
//func buildSystem() types.System {
//	return types.System{
//		Metadata: types.Metadata{ID: "welcome-system", Name: "Customer Welcome Email", Version: "1"},
//		Specification: types.SystemSpec{
//			Services: []types.Service{
//				{
//					Metadata: types.Metadata{ID: "customer-data", Name: "Build Customer Data"}, Type: "set",
//					Inputs: []types.Port{
//						{Name: "name", Type: types.ValueString, Required: true},
//						{Name: "email", Type: types.ValueString, Required: true},
//						{Name: "vip", Type: types.ValueBool, Required: true},
//					},
//					ServiceConfigurations: types.ServiceConfigurations{
//						"data": map[string]any{
//							"name":  `{{ input.name }}`,
//							"email": `{{ input.email }}`,
//							"vip":   `{{ input.vip }}`,
//						},
//					},
//				},
//				{
//					Metadata: types.Metadata{ID: "write-email", Name: "Write Welcome Email"}, Type: "ai",
//					Inputs: []types.Port{{Name: "data", Type: types.ValueObject, Required: true}},
//					ServiceConfigurations: types.ServiceConfigurations{
//						"model":       "mock-model",
//						"temperature": `{{ input.data.vip ? 0.3 : 0.7 }}`,
//						"prompt":      `Write a welcoming email to {{ input.data.name }} at {{ input.data.email }}. {{ input.data.vip ? "Treat this customer as VIP." : "Use the standard welcome tone." }}`,
//					},
//				},
//				{Metadata: types.Metadata{ID: "log-email", Name: "Log Email"}, Type: "log", Inputs: []types.Port{{Name: "content", Type: types.ValueString, Required: true}}},
//			},
//			Connectors: []types.Connector{
//				{
//					Metadata: types.Metadata{ID: "customer-to-ai"},
//					From:     types.Endpoint{ServiceID: "customer-data"}, To: types.Endpoint{ServiceID: "write-email"},
//					Validations: []types.ValidationRule{
//						{Expression: `"data" in source.output`, Message: "customer data is missing"},
//						{Expression: `source.output.data.name != null && source.output.data.email != null`, Message: "customer name and email are required"},
//					},
//					Mappings: []types.MappingRule{{TargetPath: "data", Expression: `source.output.data`}},
//				},
//				{
//					Metadata: types.Metadata{ID: "ai-to-log"},
//					From:     types.Endpoint{ServiceID: "write-email"}, To: types.Endpoint{ServiceID: "log-email"},
//					Mappings: []types.MappingRule{{TargetPath: "content", Expression: `source.output.content`}},
//				},
//			},
//		},
//	}
//}
//
//// BuildMonitoringSystem creates a workflow that pings a URL, uses AI to analyze
//// the response, runs a local command to trigger an alert, and logs the result.
//func buildMonitoringSystem() types.System {
//	return types.System{
//		Metadata: types.Metadata{
//			ID:      "api-monitor-system",
//			Name:    "API Health Monitor & AI Alert",
//			Version: "1",
//		},
//		Specification: types.SystemSpec{
//			Services: []types.Service{
//				// 1. HTTP Executor: Fetches the target API
//				{
//					Metadata: types.Metadata{ID: "fetch-api", Name: "Ping Target API"},
//					Type:     "http",
//					Inputs: []types.Port{
//						{Name: "target_url", Type: types.ValueString, Required: true},
//					},
//					ServiceConfigurations: types.ServiceConfigurations{
//						"url":    `{{ input.target_url }}`,
//						"method": "GET",
//						"headers": map[string]any{
//							"User-Agent": "Nore-Health-Monitor",
//						},
//					},
//				},
//
//				// 2. AI Executor: Analyzes the HTTP status and response body
//				{
//					Metadata: types.Metadata{ID: "analyze-response", Name: "AI Response Analysis"},
//					Type:     "ai",
//					Inputs: []types.Port{
//						{Name: "api_status", Type: types.ValueNumber, Required: true},
//						{Name: "api_body", Type: types.ValueString, Required: false},
//					},
//					ServiceConfigurations: types.ServiceConfigurations{
//						"prompt": `The monitored API returned HTTP status {{ input.api_status }}. Here is the response body: {{ input.api_body }}. Provide a brief, 1-sentence diagnostic summary.`,
//					},
//				},
//
//				// 3. Command Executor: Executes a local command/script to trigger an alert
//				{
//					Metadata: types.Metadata{ID: "trigger-alert", Name: "Run Local Alert Command"},
//					Type:     "command",
//					Inputs: []types.Port{
//						{Name: "ai_message", Type: types.ValueString, Required: true},
//					},
//					ServiceConfigurations: types.ServiceConfigurations{
//						"command": "echo",
//						"args": []any{
//							"[SYSTEM ALERT] AI Diagnostic Report:",
//							`{{ input.ai_message }}`,
//						},
//					},
//				},
//
//				// 4. Log Executor: Logs the final command output
//				{
//					Metadata: types.Metadata{ID: "log-output", Name: "Log Terminal Output"},
//					Type:     "log",
//					Inputs: []types.Port{
//						{Name: "command_stdout", Type: types.ValueString, Required: true},
//						{Name: "exit_code", Type: types.ValueNumber, Required: true},
//					},
//				},
//			},
//
//			Connectors: []types.Connector{
//				// Map HTTP output to AI inputs
//				{
//					Metadata: types.Metadata{ID: "http-to-ai"},
//					From:     types.Endpoint{ServiceID: "fetch-api"},
//					To:       types.Endpoint{ServiceID: "analyze-response"},
//					Validations: []types.ValidationRule{
//						{Expression: `source.output.status_code != null`, Message: "HTTP status code is missing"},
//					},
//					Mappings: []types.MappingRule{
//						{TargetPath: "api_status", Expression: `source.output.status_code`},
//						{TargetPath: "api_body", Expression: `source.output.body`}, // Maps the parsed JSON object
//					},
//				},
//
//				// Map AI output to Command inputs
//				{
//					Metadata: types.Metadata{ID: "ai-to-command"},
//					From:     types.Endpoint{ServiceID: "analyze-response"},
//					To:       types.Endpoint{ServiceID: "trigger-alert"},
//					Validations: []types.ValidationRule{
//						{Expression: `"content" in source.output`, Message: "AI content generation failed"},
//					},
//					Mappings: []types.MappingRule{
//						{TargetPath: "ai_message", Expression: `source.output.content`},
//					},
//				},
//
//				// Map Command output to Log inputs
//				{
//					Metadata: types.Metadata{ID: "command-to-log"},
//					From:     types.Endpoint{ServiceID: "trigger-alert"},
//					To:       types.Endpoint{ServiceID: "log-output"},
//					Mappings: []types.MappingRule{
//						{TargetPath: "command_stdout", Expression: `source.output.stdout`},
//						{TargetPath: "exit_code", Expression: `source.output.exit_code`},
//					},
//				},
//			},
//		},
//	}
//}