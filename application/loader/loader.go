package loader

type Loader struct {
	loader Contract
}

func New(loader Contract) *Loader {
	return &Loader{loader: loader}
}
