package customersystem

import (
	"development/systems/customer-system/configs"
	"development/systems/customer-system/mvp"
	"development/systems/customer-system/services"
)

var System = func() *mvp.System {
	sys := mvp.New(
		"Customer Platform",
		"1.0.0",
	)

	sys.Services(
		services.FetchAPI,
		services.AnalyzeResponse,
		services.TriggerAlert,
		services.LogOutput,
	)

	httpToAI := sys.
		Connector(
			services.FetchAPI,
			services.AnalyzeResponse,
		).
		Metadata(
			"http-to-ai",
			"HTTP Response → AI Analysis",
		)

	httpToAI.AddMappings(
		configs.Mappings.HTTPToAI...,
	)

	httpToAI.AddValidations(
		configs.Validations.HTTPToAI...,
	)

	aiToCommand := sys.
		Connector(
			services.AnalyzeResponse,
			services.TriggerAlert,
		).
		Metadata(
			"ai-to-command",
			"AI Analysis → Alert Command",
		)

	aiToCommand.AddMappings(
		configs.Mappings.AIToCommand...,
	)

	aiToCommand.AddValidations(
		configs.Validations.AIToCommand...,
	)

	sys.Connector(
		services.TriggerAlert,
		services.LogOutput,
	).
		Metadata(
			"command-to-log",
			"Alert Command → Log",
		).
		AddMappings(
			configs.Mappings.CommandToLog...,
		)

	return sys
}()