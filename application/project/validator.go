package project

import (
	"fmt"
	"strings"
)

func validateProjectBasic(
	project ProjectFile,
) error {

	var errors []string

	if project.APIVersion == "" {
		errors = append(errors, "apiVersion is required")
	}

	if project.Kind != "Project" {
		errors = append(
			errors,
			fmt.Sprintf(
				"kind must be Project, got %q",
				project.Kind,
			),
		)
	}

	if strings.TrimSpace(project.Metadata.Name) == "" {
		errors = append(errors, "metadata.name is required")
	}

	if strings.TrimSpace(project.Metadata.Version) == "" {
		errors = append(errors, "metadata.version is required")
	}

	if strings.TrimSpace(project.System.Entry) == "" {
		errors = append(errors, "systems.entry is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf(
			"%w:\n- %s",
			ErrInvalidProject,
			strings.Join(errors, "\n- "),
		)
	}

	return nil
}

func validateSystemBasic(
	system SystemFile,
) error {

	var errors []string

	if system.APIVersion == "" {
		errors = append(errors, "apiVersion is required")
	}

	if system.Kind != "System" {
		errors = append(
			errors,
			fmt.Sprintf(
				"kind must be System, got %q",
				system.Kind,
			),
		)
	}

	if strings.TrimSpace(system.Metadata.Name) == "" {
		errors = append(errors, "metadata.name is required")
	}

	if strings.TrimSpace(system.Metadata.Version) == "" {
		errors = append(errors, "metadata.version is required")
	}

	if len(system.Services) == 0 {
		errors = append(
			errors,
			"systems.services must contain at least one service",
		)
	}

	if len(errors) > 0 {
		return fmt.Errorf(
			"%s",
			strings.Join(errors, "; "),
		)
	}

	return nil
}

func validateServiceBasic(
	service ServiceFile,
) error {

	var errors []string

	if service.APIVersion == "" {
		errors = append(errors, "apiVersion is required")
	}

	if service.Kind != "Service" {
		errors = append(
			errors,
			fmt.Sprintf(
				"kind must be Service, got %q",
				service.Kind,
			),
		)
	}

	if strings.TrimSpace(service.Metadata.Name) == "" {
		errors = append(errors, "metadata.name is required")
	}

	if strings.TrimSpace(service.Metadata.Version) == "" {
		errors = append(errors, "metadata.version is required")
	}

	if strings.TrimSpace(service.Spec.Executor.Type) == "" {
		errors = append(
			errors,
			"spec.executor.type is required",
		)
	}

	if len(errors) > 0 {
		return fmt.Errorf(
			"%s",
			strings.Join(errors, "; "),
		)
	}

	return nil
}