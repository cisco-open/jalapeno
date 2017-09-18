package main

import (
	"time"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka/handler"

	"github.com/stephenrlouie/service"
)

func main() {
	serviceGroup := service.New()
	serviceGroup.HandleSigint(nil)
	brokers := []string{"10.86.204.8:9092"}
	consumer := kafka.New(brokers, "testconsgrp5")
	//consumer.SetHandler(arango.NewArangoHandler())
	serviceGroup.Add(consumer)

	serviceGroup.Start()
	//serviceGroup.Wait()
	time.Sleep(4 * time.Second)
	h := consumer.Handler.(*handler.DefaultHandler)
	h.TestPrint()
	h.TestPrint2()

}
