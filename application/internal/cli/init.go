package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command {
	Use:   "init [Target]",
	Short: "Initialize the neuron.",
	Run: initCmdHandler,
}

func initCmdHandler(cmd *cobra.Command, args []string)  {
	var targetFile string

	if len(args) >= 1 {
		targetFile = args[0]
	}

	cwd, err := os.Getwd()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	path := filepath.Join(cwd, targetFile)

	err = createDirectory(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = createConfigFile(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func createDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("expected a directory path, but received an empty string")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, os.FileMode(0755))
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func createConfigFile(path string) error  {
	if path == "" {
		fmt.Println("expected a directory path, but received an empty string")
	}

	parent, err := filepath.Abs(path)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fullPath := filepath.Join(path, config.NeuronConfigFileName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Println("Creating config file at ", fullPath)

		file, err := os.Create(fullPath)
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}(file)

		if err != nil {
			return err
		}

		_, err = file.Write([]byte(strings.Replace(config.NeuronConfigDefaultTemplate, "$", parent, 1)))
		if err != nil {
			return err
		}
		return err
	}

	fmt.Printf("File %s already exists!\n", fullPath)
	return nil
}





func init() {
	rootCmd.AddCommand(initCmd)
}