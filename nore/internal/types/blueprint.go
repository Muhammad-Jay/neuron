package types

import (
	"github.com/Muhammad-Jay/neuron/nore/internal/resolver"
	shared "github.com/Muhammad-Jay/neuron/shared/types/core"
)

// ExecutionBlueprint is the compiled, reusable representation of a System.
// It is an in-memory runtime object and should not be serialized directly.
type ExecutionBlueprint struct {
	Metadata shared.Metadata
	Nodes    map[shared.ID]ExecutionNode

	EntryServiceIDs []shared.ID
}

type ExecutionNode struct {
	Service shared.Service

	// Configurations is compiled once during System compilation and resolved
	// for each Service execution.
	Configurations resolver.ConfigurationProgram

	Next []ExecutionTransition
}

type CompiledMapping struct {
	TargetPath string
	Expression string
	Program    resolver.Program
}

type CompiledValidation struct {
	Expression string
	Message    string
	Program    resolver.Program
}

type ExecutionTransition struct {
	ConnectorID     shared.ID
	TargetServiceID shared.ID

	Mappings    []CompiledMapping
	Validations []CompiledValidation
}
