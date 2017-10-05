package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stephenrlouie/service"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/arango"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/config"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/handler"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
)

func init() {
	if err := config.InitFlags(FrameworkCmd, config.InitFrameworkCfg()); err != nil {
		FrameworkExit(err)
	}
}

var FrameworkCmd = &cobra.Command{
	Use:   "framework",
	Short: "Run Voltron Framework",
	Long:  "Voltron Usage: TBD",
	Run:   frameworkRun,
}

func FrameworkExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "--- Framework encountered an Error: ---\n")
		fmt.Fprintf(os.Stderr, "\t%v\n", err)
		os.Exit(1)
	}
}

func frameworkRun(cmd *cobra.Command, args []string) {
	var globalErr error
	defer func() {
		FrameworkExit(globalErr)
	}()

	icfg, err := config.GetConfig(config.InitFrameworkCfg())
	if err != nil {
		globalErr = err
		return
	}
	cfg := icfg.(*config.FrameworkConfig)

	jcfg, err := config.GetConfig(config.InitGlobalCfg())
	if err != nil {
		globalErr = err
		return
	}
	gcfg := jcfg.(*config.GlobalConfig)
	log.NewLogr(gcfg.Log)

	serviceGroup := service.New()
	serviceGroup.HandleSigint(nil)
	var hndlr handler.Handler = handler.NewDefault()
	arangoDB, err := arango.New(cfg.Arango)
	if err != nil {
		globalErr = err
		return
	}
	hndlr = handler.NewArango(arangoDB)

	consumer, err := kafka.New(cfg.Kafka, hndlr)
	if err != nil {
		globalErr = err
		return
	}

	consumer.SetHandler(hndlr)
	serviceGroup.Add(consumer)
	serviceGroup.Start()
	serviceGroup.Wait()
}
