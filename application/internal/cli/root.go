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
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute is called by main.go to start the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./neuron.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "Set the system logging level")

	// Add global verbose and remote flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output (shows N.O.R.E. daemon logs)")
	rootCmd.PersistentFlags().String("remote", "", "Remote N.O.R.E. endpoint (e.g., https://api.nore.example.com)")

	// Bind flags to Viper so we can access them anywhere without passing variables around
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("remote", rootCmd.PersistentFlags().Lookup("remote"))
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