// Package language enumerates the authoring languages a Neuron project can be
// written in. A language identifies how a project's source is turned into the
// canonical .neuron/manifest.json. It is intentionally distinct from both the
// executor protocol and the manifest schema: adding a new authoring language
// must not require runtime changes.
package language

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Language is one canonical project authoring language.
type Language string

const (
	// YAML is the canonical identifier for the YAML project frontend. The
	// short alias "yml" normalizes to it.
	YAML Language = "yaml"

	// TypeScript is the canonical identifier for the TypeScript SDK project
	// frontend. The short alias "ts" normalizes to it.
	TypeScript Language = "typescript"
)

// sentinel errors for language resolution.
var (
	// ErrLanguageRequired reports that no language could be determined from
	// either the command line or the project configuration.
	ErrLanguageRequired = fmt.Errorf("project language is required: pass --lang (yaml, yml, typescript, ts) or set `lang` in the project configuration")

	// ErrLanguageUnknown reports an unrecognized language token.
	ErrLanguageUnknown = fmt.Errorf("unknown project language")
)

// aliases maps every accepted authoring token to its canonical Language.
var aliases = map[string]Language{
	"yaml":       YAML,
	"yml":        YAML,
	"typescript": TypeScript,
	"ts":         TypeScript,
}

// String returns the canonical name.
func (l Language) String() string {
	return string(l)
}

// Is reports whether l equals the canonical form.
func (l Language) Is(candidate Language) bool {
	return l == candidate
}

// Normalize canonicalizes an author-provided language token. It accepts
// "yaml", "yml", "typescript", and "ts" (case-insensitive).
func Normalize(s string) (Language, error) {
	canonical, ok := aliases[strings.ToLower(strings.TrimSpace(s))]
	if !ok {
		return "", fmt.Errorf("%w: %q (supported: yaml, yml, typescript, ts)", ErrLanguageUnknown, s)
	}
	return canonical, nil
}

// Resolve determines the effective project language. The CLI flag wins;
// otherwise the project configuration value is used, then a heuristic project
// sniff: TypeScript markers (neuron.config.ts, index.ts) select the SDK, and a
// YAML project config file selects YAML. Neither present is an error.
func Resolve(flagValue, configValue, projectDir string) (Language, error) {
	if strings.TrimSpace(flagValue) != "" {
		return Normalize(flagValue)
	}
	if strings.TrimSpace(configValue) != "" {
		return Normalize(configValue)
	}
	if detected, ok := detectFromProjectDir(projectDir); ok {
		return detected, nil
	}
	return "", ErrLanguageRequired
}

// detectFromProjectDir infers the authoring language from the project root.
//
// A TypeScript project is recognized by the neuron-sdk config file
// (neuron.config.ts/js/mjs) or a conventional index.ts entry point and wins
// when present. Otherwise a YAML project config file (neuron.yaml and its
// aliases) selects YAML, so plain YAML projects work without --lang or a lang
// declaration.
func detectFromProjectDir(projectDir string) (Language, bool) {
	if projectDir == "" {
		return "", false
	}

	tsMarkers := []string{
		"neuron.config.ts",
		"neuron.config.js",
		"neuron.config.mjs",
		"index.ts",
	}

	for _, marker := range tsMarkers {
		if info, err := os.Stat(filepath.Join(projectDir, marker)); err == nil && !info.IsDir() {
			return TypeScript, true
		}
	}

	yamlMarkers := []string{
		"neuron.yaml",
		"neuron.yml",
		"neuron.config.yaml",
		"neuron.config.yml",
		"neuron.config.json",
	}

	for _, marker := range yamlMarkers {
		if info, err := os.Stat(filepath.Join(projectDir, marker)); err == nil && !info.IsDir() {
			return YAML, true
		}
	}

	return "", false
}
