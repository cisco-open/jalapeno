package cmd

import (
	"fmt"
	"os"
	"time"

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
	icfg, err := config.GetConfig(config.InitFrameworkCfg())
	if err != nil {
		FrameworkExit(err)
	}
	cfg := icfg.(*config.FrameworkConfig)

	jcfg, err := config.GetConfig(config.InitGlobalCfg())
	if err != nil {
		FrameworkExit(err)
	}
	gcfg := jcfg.(*config.GlobalConfig)
	log.NewLogr(gcfg.Log)

	serviceGroup := service.New()
	serviceGroup.HandleSigint(nil)
	var hndlr handler.Handler = handler.NewDefault()
	if len(cfg.Arango.URL) != 0 && len(cfg.Arango.Database) != 0 {
		arangoDB, err := arango.New(cfg.Arango)
		if err != nil && !cfg.Debug {
			FrameworkExit(err)
		} else if !cfg.Debug {
			hndlr = handler.NewArango(arangoDB)
		}
	}

	if cfg.Debug && len(cfg.Kafka.Brokers) == 0 {
		cfg.Kafka.Brokers = []string{"10.86.204.8:9092"}
	}

	consumer, err := kafka.New(cfg.Kafka, hndlr)
	if err != nil {
		FrameworkExit(err)
	}
	consumer.SetHandler(hndlr)
	serviceGroup.Add(consumer)
	serviceGroup.Start()

	if cfg.Debug { // THIS IS FOR DEV ONLY.
		time.Sleep(4 * time.Second)
		consumer.Handler.Debug()
		go func() {
			time.Sleep(100 * time.Millisecond)
			serviceGroup.Kill()
		}()
	}
	serviceGroup.Wait()
}
