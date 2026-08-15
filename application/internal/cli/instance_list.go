package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/Muhammad-Jay/neuron/shared/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the "instance list" command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available instances",
	RunE:  instanceListCmdHandler,
}

func instanceListCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	showAll := viper.GetBool("instance.list.all")
	statusFilter := viper.GetString("instance.list.status")

	c := client.NewDefaultLocal()
	defer c.Close()

	var query string
	if showAll {
		query = "?all=true"
	} else if statusFilter != "" {
		query = fmt.Sprintf("?status=%s", statusFilter)
	} else {
		query = "?status=running"
	}

	instances, err := c.ListInstances(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		printInstances(&[]protocol.InstanceResponse{})
		return nil
	}

	printInstances(&instances)

	return nil
}

func printInstances(instances *[]protocol.InstanceResponse) {

	columns := []utils.Column{
		{Title: "ID"},
		{Title: "System ID"},
		{Title: "Blueprint Name"},
		{Title: "Status"},
		{Title: "Version"},
		{Title: "Hash"},
		{Title: "Env"},
	}

	var rows [][]string
	for _, inst := range *instances {
		rows = append(rows, []string{
			inst.ID,
			inst.SystemID,
			inst.BlueprintMetadata.Name,
			inst.Status,
			inst.Version,
			inst.Hash,
			inst.Env,
		})
	}

	err := utils.RenderTable(os.Stdout, columns, rows, utils.DefaultTableOptions())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func init() {
	instanceCmd.AddCommand(listCmd)

	// Define the flags
	listCmd.Flags().BoolP("all", "a", false, "List all instances including inactive ones")
	listCmd.Flags().StringP("status", "s", "", "Filter instances by specific status (e.g., running, stopped)")

	// Tell Cobra these flags cannot be used together.
	// If the user runs `neuron instance list -a -s running`, Cobra will automatically throw a helpful error!
	listCmd.MarkFlagsMutuallyExclusive("all", "status")

	// Bind to Viper
	viper.BindPFlag("instance.list.all", listCmd.Flags().Lookup("all"))
	viper.BindPFlag("instance.list.status", listCmd.Flags().Lookup("status"))
}