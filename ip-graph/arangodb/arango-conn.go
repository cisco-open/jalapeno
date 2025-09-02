package arangodb

import (
	"context"
	"crypto/tls"
	"fmt"

	driver "github.com/arangodb/go-driver"
	arangohttp "github.com/arangodb/go-driver/http"
	"github.com/golang/glog"
)

// ArangoConfig holds ArangoDB connection configuration
type ArangoConfig struct {
	URL      string
	User     string
	Password string
	Database string
}

// ArangoConn represents an ArangoDB connection
type ArangoConn struct {
	client driver.Client
	db     driver.Database
}

// NewArango creates a new ArangoDB connection
func NewArango(config ArangoConfig) (*ArangoConn, error) {
	ctx := context.TODO()

	// Create HTTP connection with TLS support
	conn, err := arangohttp.NewConnection(arangohttp.ConnectionConfig{
		Endpoints: []string{config.URL},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP connection: %w", err)
	}

	// Create client with authentication
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(config.User, config.Password),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ArangoDB client: %w", err)
	}

	// Connect to database
	db, err := client.Database(ctx, config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", config.Database, err)
	}

	glog.Infof("Connected to ArangoDB: %s, database: %s", config.URL, config.Database)

	return &ArangoConn{
		client: client,
		db:     db,
	}, nil
}

// Database returns the database instance
func (a *ArangoConn) Database() driver.Database {
	return a.db
}

// Client returns the client instance
func (a *ArangoConn) Client() driver.Client {
	return a.client
}

// Query executes an AQL query
func (a *ArangoConn) Query(ctx context.Context, query string, bindVars map[string]interface{}) (driver.Cursor, error) {
	return a.db.Query(ctx, query, bindVars)
}

// Collection returns a collection instance
func (a *ArangoConn) Collection(ctx context.Context, name string) (driver.Collection, error) {
	return a.db.Collection(ctx, name)
}

// CreateCollection creates a new collection if it doesn't exist
func (a *ArangoConn) CreateCollection(ctx context.Context, name string, options *driver.CreateCollectionOptions) (driver.Collection, error) {
	exists, err := a.db.CollectionExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if collection %s exists: %w", name, err)
	}

	if exists {
		return a.db.Collection(ctx, name)
	}

	if options == nil {
		options = &driver.CreateCollectionOptions{}
	}

	collection, err := a.db.CreateCollection(ctx, name, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection %s: %w", name, err)
	}

	glog.V(6).Infof("Created collection: %s", name)
	return collection, nil
}

// CreateGraph creates a new graph if it doesn't exist
func (a *ArangoConn) CreateGraph(ctx context.Context, name string, options driver.CreateGraphOptions) (driver.Graph, error) {
	exists, err := a.db.GraphExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if graph %s exists: %w", name, err)
	}

	if exists {
		return a.db.Graph(ctx, name)
	}

	graph, err := a.db.CreateGraphV2(ctx, name, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph %s: %w", name, err)
	}

	glog.V(6).Infof("Created graph: %s", name)
	return graph, nil
}

// EnsureCollection ensures a collection exists with the specified options
func (a *ArangoConn) EnsureCollection(ctx context.Context, name string, isEdge bool) (driver.Collection, error) {
	options := &driver.CreateCollectionOptions{
		Type: driver.CollectionTypeDocument,
	}

	if isEdge {
		options.Type = driver.CollectionTypeEdge
	}

	return a.CreateCollection(ctx, name, options)
}
