package build

import (
	"fmt"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/application/project"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Build,
		Short: "Build the neuron systems",
		RunE:  buildCmdHandler,
	}

	cmd.Flags().BoolP("watch", "w", false, "watch for file changes")
	cmd.Flags().StringP("root", "r", "", "Set the project root")

	return cmd
}

func buildCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	verbose, _ := cmd.Flags().GetBool("verbose")
	isWatch, _ := cmd.Flags().GetBool("watch")
	root, _ := cmd.Flags().GetString("root")

	opts := project.DefaultOptions()
	opts.Verbose = verbose
	opts.CleanArtifact = true
	opts.ProjectRoot = root

	cmd.Println("Building project...")
	if verbose {
		cmd.Printf("Project root set to: %s\n", root)
	}

	_, err := project.Resolve(ctx, opts)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if isWatch {
		cmd.Println("Watch mode is ON. Listening for file changes...")
		// TODO: Implement fsnotify watcher block here
	} else {
		cmd.Println("Build Successful")
	}

	return nil
}