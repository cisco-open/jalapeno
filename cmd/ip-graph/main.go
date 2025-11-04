package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"

	"github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	"github.com/cisco-open/jalapeno/ip-graph/arangodb"
	"github.com/cisco-open/jalapeno/ip-graph/kafkamessenger"
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
	msgSrvAddr string
	dbSrvAddr  string
	dbName     string
	dbUser     string
	dbPass     string
	// IGP Collections (source data)
	igpv4Graph string
	igpv6Graph string
	igpNode    string
	igpDomain  string
	// IP Graph Collections (full topology)
	ipv4Graph string
	ipv6Graph string
	// BGP Collections
	bgpNode     string
	bgpPrefixV4 string
	bgpPrefixV6 string
	// Performance settings
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

	// flag.StringVar(&msgSrvAddr, "message-server", "198.18.133.112:30092", "URL to the messages supplying server")
	// flag.StringVar(&dbSrvAddr, "database-server", "http://198.18.133.112:30852", "{dns name}:port or X.X.X.X:port of the graph database")
	// flag.StringVar(&dbName, "database-name", "jalapeno", "DB name")
	// flag.StringVar(&dbUser, "database-user", "root", "DB User name")
	// flag.StringVar(&dbPass, "database-pass", "jalapeno", "DB User's password")

	// IGP Collections (source)
	flag.StringVar(&igpv4Graph, "igpv4-graph", "igpv4_graph", "IGP IPv4 graph collection name")
	flag.StringVar(&igpv6Graph, "igpv6-graph", "igpv6_graph", "IGP IPv6 graph collection name")
	flag.StringVar(&igpNode, "igp-node", "igp_node", "IGP node collection name")
	flag.StringVar(&igpDomain, "igp-domain", "igp_domain", "IGP domain collection name")

	// IP Graph Collections (full topology)
	flag.StringVar(&ipv4Graph, "ipv4-graph", "ipv4_graph", "Full IPv4 topology graph collection name")
	flag.StringVar(&ipv6Graph, "ipv6-graph", "ipv6_graph", "Full IPv6 topology graph collection name")

	// BGP Collections
	flag.StringVar(&bgpNode, "bgp-node", "bgp_node", "BGP node collection name")
	flag.StringVar(&bgpPrefixV4, "bgp-prefix-v4", "bgp_prefix_v4", "BGP IPv4 prefix collection name")
	flag.StringVar(&bgpPrefixV6, "bgp-prefix-v6", "bgp_prefix_v6", "BGP IPv6 prefix collection name")

	// Performance settings
	flag.IntVar(&batchSize, "batch-size", 1000, "Batch size for database operations")
	flag.IntVar(&concurrentWorkers, "concurrent-workers", runtime.NumCPU()*2, "Number of concurrent workers for batch processing")
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

	// validateDBCreds check if the user name and the password are provided either as
	// command line parameters or via files. If both are provided command line parameters
	// will be used, if neither, processor will fail.
	if err := validateDBCreds(); err != nil {
		glog.Errorf("failed to validate the database credentials with error: %+v", err)
		os.Exit(1)
	}

	// Initialize event notifier for publishing IP graph events
	notifier, err := kafkanotifier.NewKafkaNotifier(msgSrvAddr)
	if err != nil {
		glog.Errorf("failed to initialize events notifier with error: %+v", err)
		os.Exit(1)
	}

	// Initialize IP graph database client
	dbSrv, err := arangodb.NewDBSrvClient(arangodb.Config{
		DatabaseServer: dbSrvAddr,
		User:           dbUser,
		Password:       dbPass,
		Database:       dbName,
		// IGP source collections
		IGPv4Graph: igpv4Graph,
		IGPv6Graph: igpv6Graph,
		IGPNode:    igpNode,
		IGPDomain:  igpDomain,
		// IP graph collections (full topology)
		IPv4Graph: ipv4Graph,
		IPv6Graph: ipv6Graph,
		// BGP collections
		BGPNode:     bgpNode,
		BGPPrefixV4: bgpPrefixV4,
		BGPPrefixV6: bgpPrefixV6,
		// Performance settings
		BatchSize:         batchSize,
		ConcurrentWorkers: concurrentWorkers,
	}, notifier)
	if err != nil {
		glog.Errorf("failed to initialize database client with error: %+v", err)
		os.Exit(1)
	}

	glog.Info("IP Graph processor starting...")
	if err := dbSrv.Start(); err != nil {
		glog.Errorf("failed to start database client with error: %+v", err)
		os.Exit(1)
	}

	// Initialize Kafka messenger for consuming BMP messages
	msgSrv, err := kafkamessenger.NewKafkaMessenger(msgSrvAddr, dbSrv)
	if err != nil {
		glog.Errorf("failed to initialize message server with error: %+v", err)
		os.Exit(1)
	}

	glog.Info("Starting Kafka messenger...")
	msgSrv.Start()

	stopCh := setupSignalHandler()
	glog.Info("IP Graph processor started successfully")
	<-stopCh

	glog.Info("Shutting down IP Graph processor...")
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
