package loader

type Contract interface {
	Build()  error
}
