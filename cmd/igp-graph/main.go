// Copyright (c) 2024 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// The contents of this file are licensed under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with the
// License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"

	"github.com/cisco-open/jalapeno/igp-graph/arangodb"
	"github.com/cisco-open/jalapeno/igp-graph/kafkamessenger"
	"github.com/cisco-open/jalapeno/topology/kafkanotifier"
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
	lsprefix   string
	lslink     string
	lssrv6sid  string
	lsnode     string
	lsnodeExt  string
	igpDomain  string
	lsv4Graph  string
	lsv6Graph  string
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
	flag.StringVar(&lsnodeExt, "ls_node_extended", "ls_node_extended", "ls_node_extended Collection name, default \"ls_node_extended\"")
	flag.StringVar(&igpDomain, "igp_domain", "igp_domain", "igp_domain Collection name, default \"igp_domain\"")
	flag.StringVar(&lsv4Graph, "lsv4_graph", "lsv4_graph", "lsv4_graph Collection name, default \"lsv4_graph\"")
	flag.StringVar(&lsv6Graph, "lsv6_graph", "lsv6_graph", "lsv6_graph Collection name, default \"lsv6_graph\"")
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

	// initialize kafkanotifier to write back processed events into ls_node_edge_events topic
	notifier, err := kafkanotifier.NewKafkaNotifier(msgSrvAddr)
	if err != nil {
		glog.Errorf("failed to initialize events notifier with error: %+v", err)
		os.Exit(1)
	}

	dbSrv, err := arangodb.NewDBSrvClient(dbSrvAddr, dbUser, dbPass, dbName, lsprefix, lslink, lssrv6sid, lsnode,
		lsnodeExt, igpDomain, lsv4Graph, lsv6Graph, notifier)
	if err != nil {
		glog.Errorf("failed to initialize databse client with error: %+v", err)
		os.Exit(1)
	}

	if err := dbSrv.Start(); err != nil {
		if err != nil {
			glog.Errorf("failed to connect to database with error: %+v", err)
			os.Exit(1)
		}
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
