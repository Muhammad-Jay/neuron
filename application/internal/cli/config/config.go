package config

const NeuronConfigFileName = "neuron.yaml"

const NeuronConfigDefaultTemplate = `log_level: "debug"
lang: yaml

systems:
  metadata: $
  version: 0.0.1
`