package config

const NeuronConfigFileName = "neuron.yaml"

const NeuronConfigDefaultTemplate = `log_level: "debug"
build:
  watch: true

systems:
  metadata: $
  version: 0.0.1
`
