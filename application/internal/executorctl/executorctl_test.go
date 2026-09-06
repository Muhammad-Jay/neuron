package executorctl

import (
	"testing"

	"github.com/Muhammad-Jay/neuron/application/config"
)

func TestRequireUsesDefaultRegistriesWhenRegistryEmpty(t *testing.T) {
	catalog := &Catalog{
		cfg: config.ExecutorsConfig{
			DefaultRegistries: []string{"github", "local"},
		},
	}

	got := catalog.Require("github:read", "^1.0.0", []string{""})
	if len(got.Registries) != 2 {
		t.Fatalf("Registries = %v, want default registries", got.Registries)
	}
	if got.Registries[0] != "github" || got.Registries[1] != "local" {
		t.Errorf("Registries = %v, want [github local]", got.Registries)
	}
}

func TestRequireUsesExplicitRegistry(t *testing.T) {
	catalog := &Catalog{
		cfg: config.ExecutorsConfig{
			DefaultRegistries: []string{"github", "local"},
		},
	}

	got := catalog.Require("github:read", "^1.0.0", []string{"local"})
	if len(got.Registries) != 1 || got.Registries[0] != "local" {
		t.Errorf("Registries = %v, want [local]", got.Registries)
	}
}
