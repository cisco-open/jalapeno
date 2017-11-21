package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stephenrlouie/service"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/api/v1/server"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/config"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/database"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/manager"
)

func init() {
	if err := config.InitFlags(FrameworkCmd, config.InitFrameworkCfg()); err != nil {
		FrameworkExit(err)
	}
}

var FrameworkCmd = &cobra.Command{
	Use:   "framework",
	Short: "Framework to track vServices",
	Long:  "Voltron framework to enable RvServices to know about CvServices",
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

	arangoDB, err := database.NewArango(gcfg.Arango)
	if err != nil {
		globalErr = err
		return
	}

	api, err := server.New(cfg.API, &arangoDB)
	if err != nil {
		globalErr = err
		return
	}

	mgr, err := manager.NewManager(cfg.Manager, &arangoDB)
	if err != nil {
		globalErr = err
		return
	}

	serviceGroup.Add(mgr)
	serviceGroup.Add(api)
	serviceGroup.Start()
	serviceGroup.Wait()
}
