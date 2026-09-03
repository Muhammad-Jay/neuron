package loader

type Loader struct {
	Handler Contract
}

func New(loader Contract) *Loader {
	return &Loader{Handler: loader}
}
