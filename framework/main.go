package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/cmd"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/config"
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
	voltronCmd.AddCommand(cmd.FrameworkCmd)
	voltronCmd.AddCommand(cmd.TopologyCmd)
	voltronCmd.Execute()
}
