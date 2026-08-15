package connectors

import "github.com/Muhammad-Jay/neuron/shared/types/core"

var HttpToAI = core.Connector{
	Metadata: core.Metadata{ID: "http-to-ai"},
	From:     core.Endpoint{ServiceID: "fetch-api"},
	To:       core.Endpoint{ServiceID: "analyze-response"},
	Validations: []core.ValidationRule{
		{Expression: `source.output.status_code != null`, Message: "HTTP status code is missing"},
	},
	Mappings: []core.MappingRule{
		{TargetPath: "api_status", Expression: `source.output.status_code`},
		{TargetPath: "api_body", Expression: `source.output.body`}, // Maps the parsed JSON object
	},
}

var AIToCommand = core.Connector{
	Metadata: core.Metadata{ID: "ai-to-command"},
	From:     core.Endpoint{ServiceID: "analyze-response"},
	To:       core.Endpoint{ServiceID: "trigger-alert"},
	Validations: []core.ValidationRule{
		{Expression: `"content" in source.output`, Message: "AI content generation failed"},
	},
	Mappings: []core.MappingRule{
		{TargetPath: "ai_message", Expression: `source.output.content`},
	},
}

var CommandToLog = core.Connector{
	Metadata: core.Metadata{ID: "command-to-log"},
	From:     core.Endpoint{ServiceID: "trigger-alert"},
	To:       core.Endpoint{ServiceID: "log-output"},
	Mappings: []core.MappingRule{
		{TargetPath: "command_stdout", Expression: `source.output.stdout`},
		{TargetPath: "exit_code", Expression: `source.output.exit_code`},
	},
}
