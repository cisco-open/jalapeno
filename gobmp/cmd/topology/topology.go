package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"

	"github.com/golang/glog"
	"github.com/jalapeno-sdn/topology/pkg/arangodb"
	"github.com/jalapeno-sdn/topology/pkg/dbclient"
	"github.com/jalapeno-sdn/topology/pkg/kafkamessenger"
	"github.com/jalapeno-sdn/topology/pkg/kafkanotifier"
	"github.com/jalapeno-sdn/topology/pkg/messenger"
	"github.com/jalapeno-sdn/topology/pkg/mockdb"

	"net/http"
	_ "net/http/pprof"
)

var (
	msgSrvAddr  string
	dbSrvAddr   string
	mockDB      string
	mockMsg     string
	dbName      string
	dbUser      string
	dbPass      string
	notifyEvent string
	perfPort    = 56768
)

func init() {
	runtime.GOMAXPROCS(1)
	flag.StringVar(&msgSrvAddr, "message-server", "", "URL to the messages supplying server")
	flag.StringVar(&dbSrvAddr, "database-server", "", "{dns name}:port or X.X.X.X:port of the graph database")
	flag.StringVar(&mockDB, "mock-database", "false", "when set to true, received messages are stored in the file")
	flag.StringVar(&mockMsg, "mock-messenger", "false", "when set to true, message server is disabled.")
	flag.StringVar(&dbName, "database-name", "", "DB name")
	flag.StringVar(&dbUser, "database-user", "", "DB User name")
	flag.StringVar(&dbPass, "database-pass", "", "DB User's password")
	flag.StringVar(&notifyEvent, "notify-event", "false", "when true, a completion message is sent to kafka, indicating and end of processing of the topic's message")
}

var (
	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Interrupt}
)

func setupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func main() {
	flag.Parse()
	_ = flag.Set("logtostderr", "true")

	// Starting performance collecting http server
	go func() {
		glog.Infof("Starting performance debugging server on %d", perfPort)
		glog.Info(http.ListenAndServe(fmt.Sprintf(":%d", perfPort), nil))
	}()
	var err error
	isNotify, err := strconv.ParseBool(notifyEvent)
	if err != nil {
		glog.Errorf("invalid mock-database parameter: %s", mockDB)
		os.Exit(1)
	}
	var notifier kafkanotifier.Event
	if isNotify {
		notifier, err = kafkanotifier.NewKafkaNotifier(msgSrvAddr)
		if err != nil {
			glog.Errorf("failed to initialize events notifier with error: %+v", err)
			os.Exit(1)
		}
	}
	var dbSrv dbclient.Srv
	// Initializing databse client
	isMockDB, err := strconv.ParseBool(mockDB)
	if err != nil {
		glog.Errorf("invalid mock-database parameter: %s", mockDB)
		os.Exit(1)
	}
	if !isMockDB {
		dbSrv, err = arangodb.NewDBSrvClient(dbSrvAddr, dbUser, dbPass, dbName, notifier)
		if err != nil {
			glog.Errorf("failed to initialize databse client with error: %+v", err)
			os.Exit(1)
		}
	} else {
		dbSrv, _ = mockdb.NewDBSrvClient("")
	}

	if err := dbSrv.Start(); err != nil {
		if err != nil {
			glog.Errorf("failed to connect to database with error: %+v", err)
			os.Exit(1)
		}
	}

	// Initializing messenger process
	isMockMsg, err := strconv.ParseBool(mockMsg)
	if err != nil {
		glog.Errorf("invalid mock-messenger parameter: %s", mockMsg)
		os.Exit(1)
	}
	var msgSrv messenger.Srv
	if !isMockMsg {
		msgSrv, err = kafkamessenger.NewKafkaMessenger(msgSrvAddr, dbSrv.GetInterface())
		if err != nil {
			glog.Errorf("failed to initialize message server with error: %+v", err)
			os.Exit(1)
		}
	} else {
		// msgSrv, _ = mockmessenger.NewMockMessenger(dbSrv.GetInterface())
	}

	msgSrv.Start()

	stopCh := setupSignalHandler()
	<-stopCh

	msgSrv.Stop()
	dbSrv.Stop()

	os.Exit(0)
}
