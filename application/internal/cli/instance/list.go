package instance

import (
	"fmt"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available instances",
		RunE:  instanceListCmdHandler,
	}

	cmd.Flags().BoolP("all", "a", false, "List all instances including inactive ones")
	cmd.Flags().StringP("status", "s", "", "Filter instances by specific status (e.g., running, stopped)")

	cmd.MarkFlagsMutuallyExclusive("all", "status")

	_ = viper.BindPFlag("instance.list.all", cmd.Flags().Lookup("all"))
	_ = viper.BindPFlag("instance.list.status", cmd.Flags().Lookup("status"))

	return cmd
}

func instanceListCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	showAll := viper.GetBool("instance.list.all")
	statusFilter := viper.GetString("instance.list.status")

	c, cleanup, err := bootstrap.SetupClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

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
		render.Instances(nil)
		return nil
	}

	render.Instances(instances)
	return nil
}