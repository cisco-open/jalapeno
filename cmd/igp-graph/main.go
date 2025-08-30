package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"

	"github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	"github.com/cisco-open/jalapeno/igp-graph/arangodb"
	"github.com/cisco-open/jalapeno/igp-graph/kafkamessenger"
	"github.com/golang/glog"

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
	msgSrvAddr        string
	dbSrvAddr         string
	dbName            string
	dbUser            string
	dbPass            string
	lsprefix          string
	lslink            string
	lssrv6sid         string
	lsnode            string
	igpDomain         string
	igpNode           string
	igpv4Graph        string
	igpv6Graph        string
	lsNodeEdge        string
	batchSize         int
	concurrentWorkers int
)

func init() {
	runtime.GOMAXPROCS(1)
	flag.StringVar(&msgSrvAddr, "message-server", "", "URL to the messages supplying server")
	flag.StringVar(&dbSrvAddr, "database-server", "", "{dns name}:port or X.X.X.X:port of the graph database")
	flag.StringVar(&dbName, "database-name", "", "DB name")
	flag.StringVar(&dbUser, "database-user", "", "DB User name")
	flag.StringVar(&dbPass, "database-pass", "", "DB User's password")

	flag.StringVar(&lsprefix, "ls_prefix", "ls_prefix", "ls_prefix Collection name, default: \"ls_prefix\"")
	flag.StringVar(&lslink, "ls_link", "ls_link", "ls_link Collection name, default \"ls_link\"")
	flag.StringVar(&lssrv6sid, "ls_srv6_sid", "ls_srv6_sid", "ls_srv6_sid Collection name, default: \"ls_srv6_sid\"")
	flag.StringVar(&lsnode, "ls_node", "ls_node", "ls_node Collection name, default \"ls_node\"")
	flag.StringVar(&igpDomain, "igp_domain", "igp_domain", "igp_domain Collection name, default \"igp_domain\"")
	flag.StringVar(&igpNode, "igp_node", "igp_node", "igp_node Collection name, default \"igp_node\"")
	flag.StringVar(&igpv4Graph, "igpv4_graph", "igpv4_graph", "igpv4_graph Collection name, default \"igpv4_graph\"")
	flag.StringVar(&igpv6Graph, "igpv6_graph", "igpv6_graph", "igpv6_graph Collection name, default \"igpv6_graph\"")
	flag.StringVar(&lsNodeEdge, "ls_node_edge", "ls_node_edge", "ls_node_edge Collection name, default \"ls_node_edge\"")

	// Performance tuning flags
	flag.IntVar(&batchSize, "batch_size", 1000, "Batch size for bulk operations, default: 1000")
	flag.IntVar(&concurrentWorkers, "concurrent_workers", 0, "Number of concurrent workers, default: 2x CPU cores")
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

	// Set default concurrent workers if not specified
	if concurrentWorkers == 0 {
		concurrentWorkers = runtime.NumCPU() * 2
	}

	glog.Infof("IGP Graph processor starting with batch_size=%d, workers=%d", batchSize, concurrentWorkers)

	// validateDBCreds check if the user name and the password are provided either as
	// command line parameters or via files. If both are provided command line parameters
	// will be used, if neither, processor will fail.
	if err := validateDBCreds(); err != nil {
		glog.Errorf("failed to validate the database credentials with error: %+v", err)
		os.Exit(1)
	}

	// initialize kafkanotifier to write back processed events into igp graph topics
	notifier, err := kafkanotifier.NewKafkaNotifier(msgSrvAddr)
	if err != nil {
		glog.Errorf("failed to initialize events notifier with error: %+v", err)
		os.Exit(1)
	}

	// Initialize the unified IGP graph database client
	dbSrv, err := arangodb.NewDBSrvClient(arangodb.Config{
		URL:               dbSrvAddr,
		User:              dbUser,
		Password:          dbPass,
		Database:          dbName,
		LSPrefix:          lsprefix,
		LSLink:            lslink,
		LSSRv6SID:         lssrv6sid,
		LSNode:            lsnode,
		IGPDomain:         igpDomain,
		IGPNode:           igpNode,
		IGPv4Graph:        igpv4Graph,
		IGPv6Graph:        igpv6Graph,
		LSNodeEdge:        lsNodeEdge,
		BatchSize:         batchSize,
		ConcurrentWorkers: concurrentWorkers,
		Notifier:          notifier,
	})
	if err != nil {
		glog.Errorf("failed to initialize database client with error: %+v", err)
		os.Exit(1)
	}

	if err := dbSrv.Start(); err != nil {
		glog.Errorf("failed to start database client with error: %+v", err)
		os.Exit(1)
	}

	// Initializing messenger process
	msgSrv, err := kafkamessenger.NewKafkaMessenger(msgSrvAddr, dbSrv.GetInterface())
	if err != nil {
		glog.Errorf("failed to initialize message server with error: %+v", err)
		os.Exit(1)
	}

	msgSrv.Start()

	stopCh := setupSignalHandler()
	<-stopCh

	glog.Info("Shutting down IGP Graph processor...")
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
