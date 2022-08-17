// Copyright (c) 2022 Cisco Systems, Inc. and its affiliates
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


// We call this aio because we're taking Kafka out of the middle.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"

	"github.com/cisco-open/jalapeno/topology/arangodb"
	"github.com/cisco-open/jalapeno/topology/dbclient"
	"github.com/cisco-open/jalapeno/topology/mockdb"
	"github.com/cisco-open/jalapeno/topology/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/gobmpsrv"
	"github.com/sbezverk/gobmp/pkg/pub"
//	"github.com/byzek/gobmp/pkg/gobmpsrv"
//	"github.com/byzek/gobmp/pkg/pub"
//	pubarango "github.com/byzek/gobmp/pkg/arangodb"

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
	dstPort		int
	srcPort		int
	dbSrvAddr   string
	intercept   string
	splitAF     string
	mockDB      string
	dbName      string
	dbUser      string
	dbPass		string
	perfPort    = 56768
)

func init() {
	runtime.GOMAXPROCS(1)
	flag.IntVar(&srcPort, "source-port", 5000, "port exposed to outside")
	flag.IntVar(&dstPort, "destination-port", 5050, "port openBMP is listening")
	flag.StringVar(&dbSrvAddr, "database-server", "", "{dns name}:port or X.X.X.X:port of the graph database")
	flag.StringVar(&intercept, "intercept", "false", "When intercept set \"true\", all incomming BMP messges will be copied to TCP port specified by destination-port, otherwise received BMP messages will be published to Kafka.")
	flag.StringVar(&splitAF, "split-af", "true", "When set \"true\" (default) ipv4 and ipv6 will be published in separate topics. if set \"false\" the same topic will be used for both address families.")
	flag.StringVar(&mockDB, "mock-database", "false", "when set to true, received messages are stored in the file")
	flag.StringVar(&dbName, "database-name", "", "DB name")
	flag.StringVar(&dbUser, "database-user", "", "DB User name")
	flag.StringVar(&dbPass, "database-pass", "", "DB User's password")
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
		glog.Errorf("invalid value of \"--mock-database\" parameter: %s", mockDB)
		os.Exit(1)
	}

	var notifier kafkanotifier.Event
	if !isMockDB {
		// validateDBCreds check if the user name and the password are provided either as
		// command line parameters or via files. If both are provided command line parameters
		// will be used, if neither, topology will fail.
		if err := validateDBCreds(); err != nil {
			glog.Errorf("failed to validate the database credentials with error: %+v", err)
			os.Exit(1)
		}
		glog.Infof("dbSrvAddr is %+v", dbSrvAddr)
		dbSrv, err = arangodb.NewDBSrvClient(dbSrvAddr, dbUser, dbPass, dbName, notifier)
		if err != nil {
			glog.Errorf("failed to initialize database client with error: %+v", err)
			os.Exit(1)
		}
	} else {
		dbSrv, _ = mockdb.NewDBSrvClient()
	}

	if err := dbSrv.Start(); err != nil {
		if err != nil {
			glog.Errorf("failed to connect to database with error: %+v", err)
			os.Exit(1)
		}
	}

	var publisher pub.Publisher
	publisher, err = arangodb.NewPubArango(dbSrv)
	// Initializing bmp server
	interceptFlag, err := strconv.ParseBool(intercept)
	if err != nil {
		glog.Errorf("failed to parse to bool the value of the intercept flag with error: %+v", err)
		os.Exit(1)
	}
	splitAFFlag, err := strconv.ParseBool(splitAF)
	if err != nil {
		glog.Errorf("failed to parse to bool the value of the split AF flag with error: %+v", err)
		os.Exit(1)
	}
	bmpSrv, err := gobmpsrv.NewBMPServer(srcPort, dstPort, interceptFlag, publisher, splitAFFlag)
	if err != nil {
		glog.Errorf("failed to setup new gobmp server with error: %+v", err)
		os.Exit(1)
	}
	// Starting Interceptor server
	bmpSrv.Start()


	stopCh := setupSignalHandler()
	<-stopCh

	bmpSrv.Stop()
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
