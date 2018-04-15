package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/services/collectors/topology/cmd"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/services/collectors/topology/config"
)

var voltrontCmd = &cobra.Command{
	Use: "voltront",
}

func init() {
	if err := config.InitGlobalFlags(voltrontCmd, config.InitGlobalCfg()); err != nil {
		fmt.Fprintf(os.Stderr, "--- Voltront encountered an Error ---\n")
		fmt.Fprintf(os.Stderr, "\t%v\n", err)
	}
}

func main() {
	voltrontCmd.AddCommand(cmd.TopologyCmd)
	voltrontCmd.Execute()
}
