package render

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/Muhammad-Jay/neuron/shared/utils"
)

// Instances renders a slice of instances as a formatted table to stdout.
func Instances(instances []protocol.InstanceResponse) {
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