package config

import "fmt"

const NeuronConfigFileName = "neuron.yaml"

// NeuronConfigDefaultTemplate returns the starter project configuration for a
// new Neuron project named name.
//
// The template follows the canonical ProjectFile schema consumed by the
// project resolver, and exposes the sections the CLI configuration reads from
// the same file (runtime, storage, executors, inspector).
func NeuronConfigDefaultTemplate(name string) string {
	return fmt.Sprintf(`apiVersion: neuron/v1
kind: Project

metadata:
  name: %s
  version: 0.1.0
  description: A Neuron system

# Authoring language of this project. YAML is the canonical surface.
lang: yaml

# The system this project registers. Create it under systems/<name>/system.yaml.
systems:
  entry: ./systems/my-system/system.yaml

runtime:
  execution:
    mode: wait
    timeout: 30m
  workers:
    min: 1
    max: 8

storage:
  provider: local
  directory: ./.neuron/data

executors:
  registries:
    - name: github
      url: https://api.github.com
    - name: local
      url: local://

inspector:
  enabled: true
  address: 127.0.0.1:7433
`, name)
}
