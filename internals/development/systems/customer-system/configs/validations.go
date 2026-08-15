package configs

import (
	"development/systems/customer-system/mvp"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type validationGroups struct {
	HTTPToAI    []core.ValidationRule
	AIToCommand []core.ValidationRule
}

var Validations = validationGroups{
	HTTPToAI: mvp.Validations.Many(
		mvp.Validation(
			"source.output.status_code != null",
			"HTTP status code is missing",
		),
	),

	AIToCommand: mvp.Validations.Many(
		mvp.Validation(
			"\"content\" in source.output",
			"AI content generation failed",
		),
	),
}