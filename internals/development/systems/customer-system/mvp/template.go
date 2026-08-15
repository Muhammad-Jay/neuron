package mvp

// Template marks a string as a Neuron runtime template.
//
// The MVP implementation deliberately does not interpret the template.
//
// N.O.R.E. will receive the string and resolve {{ ... }} expressions
// during compilation/runtime according to the final resolver architecture.
//
// Example:
//
//	Prompt: Template(`
//		Hello {{ input.name }}.
//
//		{{ input.vip
//			? "Thank you for being a VIP customer."
//			: "Welcome to our platform."
//		}}
//	`)
func Template(value string) string {
	return value
}

// Expr is an explicit expression helper for places where a raw CEL
// expression is expected.
//
// Example:
//
//	mvp.Mapping(
//		"customer.name",
//		mvp.Expr("source.output.name"),
//	)
func Expr(expression string) string {
	return expression
}