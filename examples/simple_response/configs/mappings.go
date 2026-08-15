package configs

import (
	"github.com/Muhammad-Jay/neuron/application/sdk"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)
type mappingGroups struct {
	HTTPToAI     []core.MappingRule
	AIToCommand  []core.MappingRule
	CommandToLog []core.MappingRule
}

var Mappings = mappingGroups{
	HTTPToAI: sdk.Mappings.Many(
		sdk.Mapping(
			"api_status",
			sdk.Expr("source.output.status_code"),
		),
		sdk.Mapping(
			"api_body",
			sdk.Expr("source.output.body"),
		),
	),

	AIToCommand: sdk.Mappings.Many(
		sdk.Mapping(
			"ai_message",
			sdk.Expr("source.output.content"),
		),
	),

	CommandToLog: sdk.Mappings.Many(
		sdk.Mapping(
			"command_stdout",
			sdk.Expr("source.output.stdout"),
		),
		sdk.Mapping(
			"exit_code",
			sdk.Expr("source.output.exit_code"),
		),
	),
}