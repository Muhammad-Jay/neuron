package core

// MappingRule maps a value from the source Service/execution context into
// a target Service input path.
type MappingRule struct {
	TargetPath string
	Expression string
}

// ValidationRule is an optional transition assertion evaluated against
// the source Service output and execution context.
// The expression must return bool.
type ValidationRule struct {
	Expression string
	Message    string
}

// Connector has three responsibilities only:
//  1. establish execution order;
//  2. optionally map source output into target input;
//  3. optionally assert that the transition is allowed.
type Connector struct {
	Metadata Metadata

	From Endpoint
	To   Endpoint

	Mappings    []MappingRule
	Validations []ValidationRule
}