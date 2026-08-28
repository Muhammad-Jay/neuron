package register

import (
	"context"

	"github.com/Muhammad-Jay/neuron/application/parser"
	"github.com/Muhammad-Jay/neuron/application/project"
)

func resolveAndParse(ctx context.Context) (parser.Parser, error) {
	opts := project.DefaultOptions()

	results, err := project.Resolve(ctx, opts)
	if err != nil {
		return parser.Parser{}, err
	}

	p := parser.NewParser(results.Project)

	return p, nil
}
