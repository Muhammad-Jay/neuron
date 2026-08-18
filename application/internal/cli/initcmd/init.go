package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/config"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "init [Target]",
		Short: "Initialize a new neuron workspace",
		RunE:  initCmdHandler,
	}
}

func initCmdHandler(cmd *cobra.Command, args []string) error {
	var targetFile string
	if len(args) >= 1 {
		targetFile = args[0]
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	path := filepath.Join(cwd, targetFile)

	if err := createDirectory(path); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := createConfigFile(path); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	return nil
}

// createDirectory ensures the target path exists.
func createDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("expected a directory path, but received an empty string")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// createConfigFile generates a default configuration file in the target directory.
func createConfigFile(path string) error {
	if path == "" {
		return fmt.Errorf("expected a directory path, but received an empty string")
	}

	parent, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	fullPath := filepath.Join(path, config.NeuronConfigFileName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Printf("Creating config file at %s\n", fullPath)

		file, err := os.Create(fullPath)
		if err != nil {
			return err
		}
		defer file.Close()

		content := strings.Replace(config.NeuronConfigDefaultTemplate, "$", parent, 1)
		if _, err := file.Write([]byte(content)); err != nil {
			return err
		}
		return nil
	}

	fmt.Printf("File %s already exists!\n", fullPath)
	return nil
}
