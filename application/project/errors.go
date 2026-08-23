package project

import "errors"

var (
	ErrProjectNotFound    = errors.New("neuron project file not found")
	ErrInvalidProject     = errors.New("invalid neuron project")
	ErrInvalidSystem      = errors.New("invalid systems definition")
	ErrInvalidService     = errors.New("invalid service definition")
	ErrInvalidConnector   = errors.New("invalid connector definition")
	ErrCircularReference  = errors.New("circular project reference")
	ErrResolvedNotFound   = errors.New("resolved project not found")
	ErrUnsupportedKind    = errors.New("unsupported definition kind")
)