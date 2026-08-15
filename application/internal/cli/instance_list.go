package cli

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/Muhammad-Jay/neuron/shared/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available instances",
	RunE:  instanceListCmdHandler,
}

func instanceListCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	showAll := viper.GetBool("instance.list.all")
	statusFilter := viper.GetString("instance.list.status")

	// 1. Setup client and ensure Daemon is running
	c, cleanup, err := setupNoreClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	// 2. Build query parameters
	var query string
	if showAll {
		query = "?all=true"
	} else if statusFilter != "" {
		query = fmt.Sprintf("?status=%s", statusFilter)
	} else {
		query = "?status=running"
	}

	// 3. Fetch instances
	instances, err := c.ListInstances(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	// 4. Render results
	if len(instances) == 0 {
		printInstances([]protocol.InstanceResponse{})
		return nil
	}

	printInstances(instances)
	return nil
}

// printInstances renders a slice of instances as a formatted table to stdout.
func printInstances(instances []protocol.InstanceResponse) {
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
	for _, inst := range instances {
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

	if err := utils.RenderTable(os.Stdout, columns, rows, utils.DefaultTableOptions()); err != nil {
		fmt.Printf("failed to render table: %v\n", err)
	}
}

func init() {
	instanceCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("all", "a", false, "List all instances including inactive ones")
	listCmd.Flags().StringP("status", "s", "", "Filter instances by specific status (e.g., running, stopped)")

	listCmd.MarkFlagsMutuallyExclusive("all", "status")

	_ = viper.BindPFlag("instance.list.all", listCmd.Flags().Lookup("all"))
	_ = viper.BindPFlag("instance.list.status", listCmd.Flags().Lookup("status"))
}