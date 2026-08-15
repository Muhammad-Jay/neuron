package services

import (
	"development/systems/customer-system/mvp"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

var FetchAPI = mvp.NamedService(
	"fetch-api",
	"Ping Target API",
	"http",
)

func init() {
	FetchAPI.Inputs = []core.Port{
		mvp.Input(
			"target_url",
			core.ValueString,
			true,
		),
	}

	FetchAPI.ServiceConfigurations = core.ServiceConfigurations{
		"url": mvp.Template(
			"{{ input.target_url }}",
		),
		"method": "GET",
		"headers": map[string]any{
			"User-Agent": "Neuron",
		},
	}
}

var AnalyzeResponse = mvp.NamedService(
	"analyze-response",
	"AI Response Analysis",
	"ai",
)

func init() {
	AnalyzeResponse.Inputs = []core.Port{
		mvp.Input(
			"api_status",
			core.ValueNumber,
			true,
		),
		mvp.Input(
			"api_body",
			core.ValueString,
			true,
		),
	}

	AnalyzeResponse.ServiceConfigurations = core.ServiceConfigurations{
		"prompt": mvp.Template(`
The monitored API returned HTTP status {{ input.api_status }}.

Here is the response body:
{{ input.api_body }}

Provide a brief, one-sentence diagnostic summary.
`),
	}
}

var TriggerAlert = mvp.NamedService(
	"trigger-alert",
	"Run Local Alert Command",
	"command",
)

func init() {
	TriggerAlert.Inputs = []core.Port{
		mvp.Input(
			"ai_message",
			core.ValueString,
			true,
		),
	}

	TriggerAlert.ServiceConfigurations = core.ServiceConfigurations{
		"command": "echo",
		"args": []any{
			"[SYSTEM ALERT] AI Diagnostic Report:",
			mvp.Template("{{ input.ai_message }}"),
		},
	}
}

var LogOutput = mvp.NamedService(
	"log-output",
	"Log Terminal Output",
	"log",
)

func init() {
	LogOutput.Inputs = []core.Port{
		mvp.Input(
			"command_stdout",
			core.ValueString,
			true,
		),
		mvp.Input(
			"exit_code",
			core.ValueNumber,
			true,
		),
	}
}