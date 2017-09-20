package main

import (
	"fmt"
	"os"
	"time"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/arango"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/config"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka/handler"

	"github.com/spf13/cobra"
	"github.com/stephenrlouie/service"
)

var VoltronCmd = &cobra.Command{
	Use:   "voltron",
	Short: "Run Voltron",
	Long:  "Voltron Usage:",
	Run:   voltronRun,
}

func VoltronExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "--- Voltron encountered an Error: ---\n")
		fmt.Fprintf(os.Stderr, "\t%v\n", err)
		os.Exit(1)
	}
}

func main() {
	if err := config.InitFlags(VoltronCmd); err != nil {
		VoltronExit(err)
	}
	VoltronCmd.Execute()
}

func voltronRun(cmd *cobra.Command, args []string) {
	cfg, err := config.GetConfig()
	if err != nil {
		VoltronExit(err)
	}
	serviceGroup := service.New()
	serviceGroup.HandleSigint(nil)
	var hndlr handler.Handler = handler.NewDefaultHandler()
	if len(cfg.Arango.URL) != 0 && len(cfg.Arango.Database) != 0 {
		arangoDB, err := arango.New(cfg.Arango)
		if err != nil && !cfg.Debug {
			VoltronExit(err)
		} else if !cfg.Debug {
			hndlr = arango.NewHandler(arangoDB)
		}
	}

	if cfg.Debug && len(cfg.Kafka.Brokers) == 0 {
		cfg.Kafka.Brokers = []string{"10.86.204.8:9092"}
	}

	consumer, err := kafka.New(cfg.Kafka)
	if err != nil {
		VoltronExit(err)
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
