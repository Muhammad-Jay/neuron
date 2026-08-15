package configs

import (
	"development/systems/customer-system/mvp"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type mappingGroups struct {
	HTTPToAI     []core.MappingRule
	AIToCommand  []core.MappingRule
	CommandToLog []core.MappingRule
}

var Mappings = mappingGroups{
	HTTPToAI: mvp.Mappings.Many(
		mvp.Mapping(
			"api_status",
			mvp.Expr("source.output.status_code"),
		),
		mvp.Mapping(
			"api_body",
			mvp.Expr("source.output.body"),
		),
	),

	AIToCommand: mvp.Mappings.Many(
		mvp.Mapping(
			"ai_message",
			mvp.Expr("source.output.content"),
		),
	),

	CommandToLog: mvp.Mappings.Many(
		mvp.Mapping(
			"command_stdout",
			mvp.Expr("source.output.stdout"),
		),
		mvp.Mapping(
			"exit_code",
			mvp.Expr("source.output.exit_code"),
		),
	),
}