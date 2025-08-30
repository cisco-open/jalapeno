package arangodb

import (
	"context"
	"crypto/tls"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/golang/glog"
)

// ArangoConn provides ArangoDB connection management
type ArangoConn struct {
	client driver.Client
	db     driver.Database
}

// ArangoConfig holds ArangoDB connection configuration
type ArangoConfig struct {
	URL      string
	User     string
	Password string
	Database string
}

// NewArango creates a new ArangoDB connection
func NewArango(config ArangoConfig) (*ArangoConn, error) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{config.URL},
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP connection: %w", err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(config.User, config.Password),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ArangoDB client: %w", err)
	}

	// Test connection
	ctx := context.Background()
	if _, err := client.Version(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to ArangoDB: %w", err)
	}

	// Get or create database
	db, err := client.Database(ctx, config.Database)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			// Database doesn't exist, create it
			db, err = client.CreateDatabase(ctx, config.Database, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create database %s: %w", config.Database, err)
			}
			glog.Infof("Created database: %s", config.Database)
		} else {
			return nil, fmt.Errorf("failed to access database %s: %w", config.Database, err)
		}
	}

	glog.Infof("Connected to ArangoDB: %s, database: %s", config.URL, config.Database)

	return &ArangoConn{
		client: client,
		db:     db,
	}, nil
}

// Client returns the ArangoDB client
func (ac *ArangoConn) Client() driver.Client {
	return ac.client
}

// Database returns the ArangoDB database
func (ac *ArangoConn) Database() driver.Database {
	return ac.db
}

// Close closes the ArangoDB connection
func (ac *ArangoConn) Close() error {
	// ArangoDB Go driver doesn't require explicit connection closing
	glog.Info("ArangoDB connection closed")
	return nil
}
