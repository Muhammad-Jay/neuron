package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Muhammad-Jay/neuron/application/config"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/build"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/daemon"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/execution"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/initcmd"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/instance"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/register"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/run"
	"github.com/spf13/cobra"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   command.Neuron,
	Short: "Neuron workflow engine CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute is called by main.go to start the CLI.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "project config file (defaults to ./neuron.yaml)")
	RootCmd.PersistentFlags().String("log-level", "info", "Set the systems logging level")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output (shows N.O.R.E. daemon logs)")
	RootCmd.PersistentFlags().String("remote", "", "Remote N.O.R.E. endpoint (e.g., https://api.nore.example.com)")
	RootCmd.PersistentFlags().String("nore-path", "", "Path to the nore daemon binary")

	// Load the effective configuration once flags are parsed and inject it on
	// the command context so every subcommand can read it without touching
	// Viper itself.
	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		cmd.SetContext(config.NewContext(cmd.Context(), cfg))
		return nil
	}

	RootCmd.AddCommand(
		run.New(),
		build.New(),
		instance.New(),
		execution.New(),
		daemon.New(),
		initcmd.New(),
		register.New(),
	)
}

// loadConfig assembles the effective configuration from the resolved project
// directory and the parsed command-line overrides.
func loadConfig(cmd *cobra.Command) (config.Config, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return config.Config{}, fmt.Errorf("get current directory: %w", err)
	}

	cli := map[string]any{}

	if remote, _ := cmd.Flags().GetString("remote"); remote != "" {
		cli["daemon.endpoint"] = remote
	}
	if path, _ := cmd.Flags().GetString("nore-path"); path != "" {
		cli["daemon.norePath"] = path
	}

	opts := config.Options{
		ProjectDir:  projectDir,
		Environment: true,
		CLI:         cli,
	}

	if cfgFile != "" {
		opts.ProjectPath = cfgFile
		opts.ProjectDir = filepath.Dir(cfgFile)
	}

	return config.Load(opts)
}
