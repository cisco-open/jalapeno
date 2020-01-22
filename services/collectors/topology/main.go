package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/cmd"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/config"
)

var jalapenoCmd = &cobra.Command{
	Use: "jalapeno",
}

func init() {
	if err := config.InitGlobalFlags(jalapenoCmd, config.InitGlobalCfg()); err != nil {
		fmt.Fprintf(os.Stderr, "--- Jalapeno encountered an Error ---\n")
		fmt.Fprintf(os.Stderr, "\t%v\n", err)
	}
}

func main() {
	jalapenoCmd.AddCommand(cmd.TopologyCmd)
	jalapenoCmd.Execute()
}
