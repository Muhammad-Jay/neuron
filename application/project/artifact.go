package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	neuronDirectory   = ".neuron"
	resolvedDirectory = "resolved"

	resolvedProjectFile = "project.json"
	resolvedManifestFile = "manifest.json"
)

type ArtifactManifest struct {
	FormatVersion string    `json:"formatVersion"`
	ResolvedAt    time.Time `json:"resolvedAt"`
	Project       string    `json:"project"`
	System        string    `json:"systems"`
}

// ArtifactPath returns the generated project artifact directory.
func ArtifactPath(projectRoot string) string {
	return filepath.Join(
		projectRoot,
		neuronDirectory,
		resolvedDirectory,
	)
}

// SaveResolved writes a resolved project to .neuron/resolved.
//
// The write is performed through a temporary directory and then
// renamed into place so consumers do not observe a half-written
// artifact.
func SaveResolved(
	projectRoot string,
	project *ResolvedProject,
) error {

	if project == nil {
		return fmt.Errorf("resolved project is nil")
	}

	dir := ArtifactPath(projectRoot)

	parent := filepath.Dir(dir)

	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf(
			"create .neuron directory: %w",
			err,
		)
	}

	tempDir, err := os.MkdirTemp(
		parent,
		"resolved-*",
	)
	if err != nil {
		return fmt.Errorf(
			"create temporary resolved directory: %w",
			err,
		)
	}

	defer os.RemoveAll(tempDir)

	projectData, err := json.MarshalIndent(
		project,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf(
			"encode resolved project: %w",
			err,
		)
	}

	manifest := ArtifactManifest{
		FormatVersion: project.FormatVersion,
		ResolvedAt:    project.ResolvedAt,
		Project:       project.Project.Metadata.Name,
		System:        project.System.Definition.Metadata.Name,
	}

	manifestData, err := json.MarshalIndent(
		manifest,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf(
			"encode artifact manifest: %w",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			tempDir,
			resolvedProjectFile,
		),
		projectData,
		0644,
	); err != nil {
		return fmt.Errorf(
			"write resolved project: %w",
			err,
		)
	}

	if err := os.WriteFile(
		filepath.Join(
			tempDir,
			resolvedManifestFile,
		),
		manifestData,
		0644,
	); err != nil {
		return fmt.Errorf(
			"write resolved manifest: %w",
			err,
		)
	}

	// MVP behavior requested: replace the previous artifact.
	_ = os.RemoveAll(dir)

	if err := os.Rename(tempDir, dir); err != nil {
		return fmt.Errorf(
			"activate resolved project: %w",
			err,
		)
	}

	return nil
}

// LoadResolved loads the previously resolved project.
//
// This does NOT inspect or resolve source YAML files.
func LoadResolved(
	projectRoot string,
) (*ResolvedProject, error) {

	path := filepath.Join(
		ArtifactPath(projectRoot),
		resolvedProjectFile,
	)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrResolvedNotFound
		}

		return nil, fmt.Errorf(
			"read resolved project: %w",
			err,
		)
	}

	var project ResolvedProject

	if err := json.Unmarshal(
		data,
		&project,
	); err != nil {
		return nil, fmt.Errorf(
			"decode resolved project: %w",
			err,
		)
	}

	return &project, nil
}