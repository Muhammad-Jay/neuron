package build

import (
	"fmt"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/application/project"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Build,
		Short: "Build the neuron systems",
		RunE:  buildCmdHandler,
	}

	cmd.Flags().BoolP("watch", "w", false, "watch for file changes")
	cmd.Flags().StringP("root", "r", "", "Set the project root")

	// Ignoring errors here is generally safe for static strings,
	// but panic on err ensures typos in flag names are caught at boot.
	if err := viper.BindPFlag("build.watch", cmd.Flags().Lookup("watch")); err != nil {
		panic(fmt.Sprintf("failed to bind watch flag: %v", err))
	}
	if err := viper.BindPFlag("build.root", cmd.Flags().Lookup("root")); err != nil {
		panic(fmt.Sprintf("failed to bind root flag: %v", err))
	}

	return cmd
}

func buildCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	verbose := viper.GetBool("verbose")
	isWatch := viper.GetBool("build.watch")
	root := viper.GetString("build.root")

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