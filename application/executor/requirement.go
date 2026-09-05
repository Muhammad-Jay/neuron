package executor

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Requirement is what a System asks for: a logical executor type, an optional
// version constraint, and the registries allowed to satisfy it.
//
//	type:    "github:read"
//	version: "^1.0.0"
//	registries: ["github"]
//
// A Requirement is a request, not an installation. Resolution turns it into an
// immutable Package, and installation turns that Package into an Installed
// executor. N.O.R.E. never stores Requirements as resolved executors.
type Requirement struct {
	Type       string   `json:"type"`
	Version    string   `json:"version,omitempty"`
	Registries []string `json:"registries,omitempty"`
}

// Validate verifies a Requirement is addressable.
func (r Requirement) Validate() error {
	if strings.TrimSpace(r.Type) == "" {
		return fmt.Errorf("executor requirement has no type")
	}
	if _, err := ParseType(r.Type); err != nil {
		return err
	}
	return nil
}

// NameSplit holds the structural components of a logical executor name.
//
// A logical name is a ':' separated path whose first segment is the registry
// owner and whose remaining segments are the functional path:
//
//	"github:read"                 -> Owner: "github", PathSegments: ["read"]
//	"Muhammad-Jay:github:read"    -> Owner: "Muhammad-Jay", PathSegments: ["github", "read"]
//	"hashicorp:vault:auth"        -> Owner: "hashicorp", PathSegments: ["vault", "auth"]
type NameSplit struct {
	Owner        string   // e.g. "Muhammad-Jay"
	PathSegments []string // e.g. ["github", "read"]
}

// ParseType decomposes a logical executor name such as "Muhammad-Jay:github:read"
// or "github:read". The first segment is always the owner; at least one
// functional segment must follow it.
func ParseType(executorType string) (*NameSplit, error) {
	parts := strings.Split(executorType, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid executor type format %q; must contain at least one ':' separator", executorType)
	}

	owner := strings.TrimSpace(parts[0])
	if owner == "" {
		return nil, fmt.Errorf("invalid executor type format %q; owner segment is empty", executorType)
	}

	rest := make([]string, 0, len(parts)-1)
	for _, seg := range parts[1:] {
		if seg == "" {
			return nil, fmt.Errorf("invalid executor type format %q; empty path segment", executorType)
		}
		rest = append(rest, seg)
	}

	return &NameSplit{
		Owner:        owner,
		PathSegments: rest,
	}, nil
}

// ToGitHubRepo builds the "owner/repo" string GitHub's API expects. Trailing
// path segments are hyphen-joined because GitHub has no nested repositories.
//
//	"github:read"              -> "github/read"
//	"Muhammad-Jay:github:read" -> "Muhammad-Jay/github-read"
func (n *NameSplit) ToGitHubRepo() string {
	repoName := strings.Join(n.PathSegments, "-")
	return fmt.Sprintf("%s/%s", n.Owner, repoName)
}

// ToLocalPath builds the multi-level filesystem path segments used under the
// executor store. The ':' delimiters become directory separators, which is
// both cross-platform safe and keeps related executors grouped by owner.
func (n *NameSplit) ToLocalPath() string {
	return joinSegments(append([]string{n.Owner}, n.PathSegments...))
}

// TypePath converts a logical name directly to its store-relative path,
// e.g. "Muhammad-Jay:github:read" -> "Muhammad-Jay/github/read".
func TypePath(executorType string) (string, error) {
	split, err := ParseType(executorType)
	if err != nil {
		return "", err
	}
	return split.ToLocalPath(), nil
}

// NormalizeType rounds a logical name to its canonical ':' form. This lets
// callers compare requirements across the repository naming and storage path
// variations.
func NormalizeType(executorType string) (string, error) {
	split, err := ParseType(executorType)
	if err != nil {
		return "", err
	}
	return split.String(), nil
}

// String renders the canonical logical name.
func (n *NameSplit) String() string {
	return strings.Join(append([]string{n.Owner}, n.PathSegments...), ":")
}

func joinSegments(segments []string) string {
	return filepath.FromSlash(strings.Join(segments, "/"))
}
