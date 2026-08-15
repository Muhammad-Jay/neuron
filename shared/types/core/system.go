package core

type System struct {
	Metadata      Metadata
	Specification SystemSpec
}

type SystemSpec struct {
	Services   []Service
	Triggers   []Trigger
	Connectors []Connector
}