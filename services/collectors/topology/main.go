package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/cmd"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/config"
)

var voltronCmd = &cobra.Command{
	Use: "voltron",
}

func init() {
	if err := config.InitGlobalFlags(voltronCmd, config.InitGlobalCfg()); err != nil {
		fmt.Fprintf(os.Stderr, "--- Voltron encountered an Error ---\n")
		fmt.Fprintf(os.Stderr, "\t%v\n", err)
	}
}

func main() {
	voltronCmd.AddCommand(cmd.TopologyCmd)
	voltronCmd.Execute()
}
