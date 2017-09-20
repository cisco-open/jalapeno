package arango

import (
	"context"
	"fmt"
	"reflect"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	log "wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
)

type ArangoConfig struct {
	URL      string `desc:"Arangodb server URL"`
	User     string `desc:"Arangodb server username"`
	Password string `desc:"Arangodb server user password"`
	Database string `desc:"Arangodb database name"`
}

func NewConfig() ArangoConfig {
	return ArangoConfig{}
}

type ArangoConn struct {
	db   driver.Database
	g    driver.Graph
	cols map[string]driver.Collection
}

func New(cfg ArangoConfig) (ArangoConn, error) {
	// Connect to DB

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{cfg.URL},
	})
	if err != nil {
		log.Errorf("Failed to create HTTP connection: %v", err)
		return ArangoConn{}, err
	}

	// Authenticate with DB
	conn, err = conn.SetAuthentication(driver.BasicAuthentication(cfg.User, cfg.Password))
	if err != nil {
		log.Errorf("Failed to authenticate with arango: %v", err)
		return ArangoConn{}, err
	}

	c, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
		return ArangoConn{}, err
	}

	db, err := ensureDatabase(c, cfg)
	if err != nil {
		log.WithError(err).Errorf("Failed to create DB")
		return ArangoConn{}, err
	}

	g, err := ensureGraph(db, graphName)
	if err != nil {
		log.WithError(err).Errorf("Failed to create Graph")
		return ArangoConn{}, err
	}

	// Create / Connect  collections
	cols := make(map[string]driver.Collection)
	cols[prefixName], err = ensureVertexCollection(g, prefixName)
	if err != nil {
		log.WithError(err).Errorf("Failed to connect to collection %q", prefixName)
	}

	cols[routerName], err = ensureVertexCollection(g, routerName)
	if err != nil {
		log.WithError(err).Errorf("Failed to connect to collection %q", routerName)
	}

	cols[asName], err = ensureEdgeCollection(g, asName, []string{prefixName}, []string{prefixName})
	if err != nil {
		log.WithError(err).Errorf("Failed to connect to collection %q", asName)
	}

	cols[linkName], err = ensureEdgeCollection(g, linkName, []string{routerName}, []string{routerName})
	if err != nil {
		log.WithError(err).Errorf("Failed to connect to collection %q", linkName)
	}

	return ArangoConn{db: db, g: g, cols: cols}, nil
}

func ensureDatabase(c driver.Client, cfg ArangoConfig) (driver.Database, error) {
	var db driver.Database

	exists, err := c.DatabaseExists(context.Background(), cfg.Database)
	if err != nil {
		return db, err
	}

	if !exists {
		// Create database
		db, err = c.CreateDatabase(context.Background(), cfg.Database, nil)
		if err != nil {
			return db, err
		}
	} else {
		db, err = c.Database(context.Background(), cfg.Database)
		if err != nil {
			return db, err
		}
	}
	return db, nil
}

func ensureGraph(db driver.Database, name string) (driver.Graph, error) {
	var g driver.Graph
	exists, err := db.GraphExists(context.Background(), name)
	if err != nil {
		return g, err
	}

	if !exists {
		// Create database
		g, err = db.CreateGraph(context.Background(), name, nil)
		if err != nil {
			return g, err
		}
	} else {
		g, err = db.Graph(context.Background(), name)
		if err != nil {
			return g, err
		}
	}
	return g, nil
}

func ensureVertexCollection(g driver.Graph, name string) (driver.Collection, error) {
	var col driver.Collection
	exists, err := g.VertexCollectionExists(context.Background(), name)
	if err != nil {
		return col, err
	}

	if !exists {
		col, err = g.CreateVertexCollection(context.Background(), name)
		if err != nil {
			return col, err
		}
	} else {
		col, err = g.VertexCollection(context.Background(), name)
		if err != nil {
			return col, err
		}
	}
	return col, nil
}

func ensureEdgeCollection(g driver.Graph, name string, from []string, to []string) (driver.Collection, error) {
	var col driver.Collection
	exists, err := g.EdgeCollectionExists(context.Background(), name)
	if err != nil {
		return col, err
	}

	if !exists {
		col, err = g.CreateEdgeCollection(context.Background(), name, driver.VertexConstraints{From: from, To: to})
		if err != nil {
			return col, err
		}
	} else {
		// ignoring vertex constraints for now
		col, _, err = g.EdgeCollection(context.Background(), name)
		if err != nil {
			return col, err
		}
	}
	return col, nil
}

func (a *ArangoConn) Add(i interface{}) error {
	col, err := a.findCollection(i)
	if err != nil {
		return err
	}

	_, err = col.CreateDocument(context.Background(), i)
	return err
}

func (a *ArangoConn) Update(key string, i interface{}) error {
	col, err := a.findCollection(i)
	if err != nil {
		return err
	}

	_, err = col.UpdateDocument(context.Background(), key, i)
	return err
}

func (a *ArangoConn) Read(key string, i interface{}) (interface{}, error) {
	col, err := a.findCollection(i)
	if err != nil {
		return nil, err
	}
	obj, err := a.findObj(i)
	if err != nil {
		return nil, err
	}
	_, err = col.ReadDocument(context.Background(), key, &obj)
	return obj, err
}

func (a *ArangoConn) Delete(key string, i interface{}) error {
	col, err := a.findCollection(i)
	if err != nil {
		return err
	}

	_, err = col.RemoveDocument(context.Background(), key)
	return err
}

func (a *ArangoConn) Query(q string) ([]interface{}, error) {
	err := a.db.ValidateQuery(context.Background(), q)
	if err != nil {
		return nil, err
	}

	cursor, err := a.db.Query(context.Background(), q, nil)
	if err != nil {
		return nil, err
	}

	i := make([]interface{}, 0)
	obj, err := a.findObj(i)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()
	for {
		_, err := cursor.ReadDocument(context.Background(), &obj)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			// handle other errors
			return nil, err
		}
		i = append(i, obj)
	}
	return i, nil
}

func (a *ArangoConn) findCollection(i interface{}) (driver.Collection, error) {
	n := ""
	switch r := reflect.TypeOf(i).Name(); r {
	case prefixName:
		n = prefixName
	case routerName:
		n = routerName
	case asName:
		n = asName
	case linkName:
		n = linkName
	}

	val, ok := a.cols[n]
	if !ok {
		return val, fmt.Errorf("Could not find collection: %q", n)
	}
	return val, nil
}

func (a *ArangoConn) findObj(i interface{}) (interface{}, error) {
	var n interface{}
	switch r := reflect.TypeOf(i).Name(); r {
	case prefixName:
		n = Prefix{}
	case routerName:
		n = Router{}
	case asName:
		n = ASEdge{}
	case linkName:
		n = LinkEdge{}
	}

	return n, nil
}
