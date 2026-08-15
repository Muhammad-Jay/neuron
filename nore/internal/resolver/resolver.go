package resolver

import "context"

// Environment is available to Connector mapping and validation expressions.
type Environment struct {
	Source    map[string]any
	Execution map[string]any
}

// ServiceEnvironment is available to Service configuration templates.
type ServiceEnvironment struct {
	Input     map[string]any
	Execution map[string]any
	Service   map[string]any
}

// Program is a compiled Connector mapping/validation expression.
type Program interface {
	Evaluate(ctx context.Context, environment Environment) (any, error)
	Expression() string
}

// ConfigurationProgram is a recursively compiled Service configuration map.
type ConfigurationProgram interface {
	Resolve(ctx context.Context, environment ServiceEnvironment) (map[string]any, error)
}

// Compiler compiles both Connector expressions and Service configuration templates.
type Compiler interface {
	CompileTransitionExpression(expression string) (Program, error)
	CompileServiceConfigurations(config map[string]any) (ConfigurationProgram, error)
}
