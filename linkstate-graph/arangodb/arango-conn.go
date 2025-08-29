// Borrowed from https://github.com/cisco-ie/jalapeno

package arangodb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/golang/glog"
)

var (
	ErrEmptyConfig = errors.New("ArangoDB Config has an empty field")
	ErrUpSafe      = errors.New("Failed to UpdateSafe. Requires *DBObjects")
	ErrNilObject   = errors.New("Failed to operate on NIL object")
	ErrNotFound    = errors.New("Document not found")
)

type ArangoConfig struct {
	URL      string `desc:"Arangodb server URL (http://127.0.0.1:8529)"`
	User     string `desc:"Arangodb server username"`
	Password string `desc:"Arangodb server user password"`
	Database string `desc:"Arangodb database name"`
}

func NewConfig() ArangoConfig {
	return ArangoConfig{}
}

type ArangoConn struct {
	db driver.Database
}

var (
	ErrCollectionNotFound = fmt.Errorf("Could not find collection")
)

func NewArango(cfg ArangoConfig) (*ArangoConn, error) {
	// Connect to DB
	if cfg.URL == "" || cfg.User == "" || cfg.Password == "" || cfg.Database == "" {
		return nil, ErrEmptyConfig
	}
	if !strings.Contains(cfg.URL, "http") {
		cfg.URL = "http://" + cfg.URL
	}
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{cfg.URL},
	})
	if err != nil {
		glog.Errorf("Failed to create HTTP connection: %v", err)
		return nil, err
	}

	// Authenticate with DB
	conn, err = conn.SetAuthentication(driver.BasicAuthentication(cfg.User, cfg.Password))
	if err != nil {
		glog.Errorf("Failed to authenticate with arango: %v", err)
		return nil, err
	}

	c, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		glog.Errorf("Failed to create client: %v", err)
		return nil, err
	}

	// If Jalapeno databse does not exist, the topology is not running, goig into a crash loop
	db, err := c.Database(context.Background(), cfg.Database)
	if err != nil {
		return nil, err
	}

	return &ArangoConn{db: db}, nil
}
