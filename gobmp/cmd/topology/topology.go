package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"

	"github.com/golang/glog"
	"github.com/jalapeno/topology/pkg/arangodb"
	"github.com/jalapeno/topology/pkg/dbclient"
	"github.com/jalapeno/topology/pkg/kafkamessenger"
	"github.com/jalapeno/topology/pkg/kafkanotifier"
	"github.com/jalapeno/topology/pkg/messenger"
	"github.com/jalapeno/topology/pkg/mockdb"

	"net/http"
	_ "net/http/pprof"
)

const (
	// userFile defines the name of file containing base64 encoded user name
	userFile = "./credentials/.username"
	// passFile defines the name of file containing base64 encoded password
	passFile = "./credentials/.password"
	// MAXUSERNAME defines maximum length of ArangoDB user name
	MAXUSERNAME = 256
	// MAXPASS defines maximum length of ArangoDB password
	MAXPASS = 256
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
	// validateDBCreds check if the user name and the password are provided either as
	// command line parameters or via files. If both are provided command line parameters
	// will be used, if neither, topology will fail.
	if err := validateDBCreds(); err != nil {
		glog.Errorf("failed to validate the database credentials with error: %+v", err)
		os.Exit(1)
	}
	// Initializing database client
	isMockDB, err := strconv.ParseBool(mockDB)
	if err != nil {
		glog.Errorf("invalid mock-database parameter: %s", mockDB)
		os.Exit(1)
	}
	if !isMockDB {
		dbSrv, err = arangodb.NewDBSrvClient(dbSrvAddr, dbUser, dbPass, dbName, notifier)
		if err != nil {
			glog.Errorf("failed to initialize database client with error: %+v", err)
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

func validateDBCreds() error {
	// Attempting to access username and password files.
	u, err := readAndDecode(userFile, MAXUSERNAME)
	if err != nil {
		if dbUser != "" && dbPass != "" {
			return nil
		}
		return fmt.Errorf("failed to access %s with error: %+v and no username and password provided via command line arguments", userFile, err)
	}
	p, err := readAndDecode(passFile, MAXPASS)
	if err != nil {
		if dbUser != "" && dbPass != "" {
			return nil
		}
		return fmt.Errorf("failed to access %s with error: %+v and no username and password provided via command line arguments", passFile, err)
	}
	dbUser, dbPass = u, p

	return nil
}

func readAndDecode(fn string, max int) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()
	l, err := f.Stat()
	if err != nil {
		return "", err
	}
	b := make([]byte, int(l.Size()))
	n, err := io.ReadFull(f, b)
	if err != nil {
		return "", err
	}
	if n > max {
		return "", fmt.Errorf("length of data %d exceeds maximum acceptable length: %d", n, max)
	}
	b = b[:n]

	return string(b), nil
}
