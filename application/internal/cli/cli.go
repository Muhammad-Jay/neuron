package cli

import (
	"fmt"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/build"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/daemon"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/execution"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/initcmd"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/instance"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "neuron",
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
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./neuron.yaml)")
	RootCmd.PersistentFlags().String("log-level", "info", "Set the systems logging level")

	// Add global verbose and remote flags
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output (shows N.O.R.E. daemon logs)")
	RootCmd.PersistentFlags().String("remote", "", "Remote N.O.R.E. endpoint (e.g., https://api.nore.example.com)")

	// Bind flags to Viper so we can access them anywhere without passing variables around
	_ = viper.BindPFlag("log_level", RootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("remote", RootCmd.PersistentFlags().Lookup("remote"))

	RootCmd.AddCommand(
		run.New(),
		build.New(),
		instance.New(),
		execution.New(),
		daemon.New(),
		initcmd.New(),
	)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("neuron")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("neuron")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
