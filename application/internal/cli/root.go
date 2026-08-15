package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "neuron",
	Short: "Neuron workflow engine CLI",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

// Execute is called by main.go
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 1. Tell Cobra to run initConfig BEFORE any command executes
	cobra.OnInitialize(initConfig)

	// 2. Define the global --config flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./neuron.yaml)")

	// 3. Define another global flag (e.g., --log-level)
	// Notice we don't bind this to a variable! We let Viper handle it.
	rootCmd.PersistentFlags().String("log-level", "info", "Set the system logging level")

	// 4. Bind the flag to Viper
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
}

// initConfig reads the config file and environment variables.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for a default "neuron.yaml" in the current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("neuron")
		viper.SetConfigType("yaml")
	}

	// Tell Viper to also look for Environment variables that start with NEURON_
	// e.g., NEURON_LOG_LEVEL=debug
	viper.SetEnvPrefix("neuron")
	viper.AutomaticEnv()

	// Attempt to read the config file (ignore error if it doesn't exist)
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}