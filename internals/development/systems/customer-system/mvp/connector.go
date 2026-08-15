package mvp

import (
	"fmt"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

// Connector is the developer-facing representation of a types Connector.
//
// It deliberately exposes only the operations that make sense while
// constructing a transition:
//
//   - Metadata
//   - AddMapping / AddMappings
//   - AddValidation / AddValidations
//
// Execution, compilation and validation semantics remain in N.O.R.E.
type Connector struct {
	connector core.Connector
}

func NewConnector(
	sourceID core.ID,
	targetID core.ID,
) *Connector {
	return &Connector{
		connector: core.Connector{
			Metadata: core.Metadata{
				ID: core.NewID("connector_"),
			},
			From: core.Endpoint{
				ServiceID: sourceID,
			},
			To: core.Endpoint{
				ServiceID: targetID,
			},
			Mappings:    make([]core.MappingRule, 0),
			Validations: make([]core.ValidationRule, 0),
		},
	}
}

// Metadata configures connector metadata.
//
// Example:
//
//	sys.Connector(a, b).
//		Metadata("customer-to-email", "Customer to Email")
func (c *Connector) Metadata(
	id, name string,
) *Connector {
	c.connector.Metadata.ID = core.ID(id)
	c.connector.Metadata.Name = name

	return c
}

func (c *Connector) Description(
	description string,
) *Connector {
	c.connector.Metadata.Description = description

	return c
}

func (c *Connector) Version(
	version string,
) *Connector {
	c.connector.Metadata.Version = version

	return c
}

// AddMapping adds one mapping.
//
// Example:
//
//	connector.AddMapping(
//		Mapping("customer.name", Expr("input.name")),
//	)
func (c *Connector) AddMapping(
	mapping core.MappingRule,
) *Connector {
	c.connector.Mappings =
		append(
			c.connector.Mappings,
			mapping,
		)

	return c
}

// AddMappings adds multiple mappings.
//
// This is the preferred API for configuration packages.
func (c *Connector) AddMappings(
	mappings ...core.MappingRule,
) *Connector {
	c.connector.Mappings =
		append(
			c.connector.Mappings,
			mappings...,
		)

	return c
}

// AddValidation adds one validation rule.
func (c *Connector) AddValidation(
	validation core.ValidationRule,
) *Connector {
	c.connector.Validations =
		append(
			c.connector.Validations,
			validation,
		)

	return c
}

// AddValidations adds multiple validation rules.
func (c *Connector) AddValidations(
	validations ...core.ValidationRule,
) *Connector {
	c.connector.Validations =
		append(
			c.connector.Validations,
			validations...,
		)

	return c
}

// Mapping creates a MappingRule.
//
// The API intentionally uses Target + Expression rather than exposing
// the types struct directly in normal developer code.
func Mapping(
	target string,
	expression string,
) core.MappingRule {
	if target == "" {
		panic("mvp: mapping target cannot be empty")
	}

	if expression == "" {
		panic("mvp: mapping expression cannot be empty")
	}

	return core.MappingRule{
		TargetPath: target,
		Expression: expression,
	}
}

// Validation creates a ValidationRule.
func Validation(
	expression string,
	message string,
) core.ValidationRule {
	if expression == "" {
		panic("mvp: validation expression cannot be empty")
	}

	return core.ValidationRule{
		Expression: expression,
		Message:    message,
	}
}

// Build exposes the underlying types Connector to package-level
// configuration and import/export tooling.
func (c *Connector) Build() core.Connector {
	if c == nil {
		panic("mvp: nil connector")
	}

	return c.connector
}

// Core is an alias for Build for callers that prefer explicit terminology.
func (c *Connector) Core() core.Connector {
	return c.Build()
}

// Source and Target are intentionally read-only developer helpers.
func (c *Connector) Source() core.ID {
	return c.connector.From.ServiceID
}

func (c *Connector) Target() core.ID {
	return c.connector.To.ServiceID
}

// Must ensures a connector is structurally complete before it is exported.
func (c *Connector) Must() *Connector {
	if c == nil {
		panic("mvp: connector is nil")
	}

	if c.connector.From.ServiceID == "" {
		panic("mvp: connector source service is required")
	}

	if c.connector.To.ServiceID == "" {
		panic("mvp: connector target service is required")
	}

	return c
}

func (c *Connector) String() string {
	if c == nil {
		return "<nil>"
	}

	return fmt.Sprintf(
		"%s -> %s",
		c.connector.From.ServiceID,
		c.connector.To.ServiceID,
	)
}