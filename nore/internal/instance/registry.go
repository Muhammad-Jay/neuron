package instance

// Registry is deliberately implemented by Manager for the MVP.
//
// Manager is the in-process runtime registry. It is not a database registry.
// If N.O.R.E. later becomes multi-process/multi-node, this package can add a
// persistent registry without changing the client protocol.
type Registry interface {
	Get(Key) (*Instance, bool)
	GetByID(string) (*Instance, bool)
	List() []*Instance
}
