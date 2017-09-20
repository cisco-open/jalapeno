package arango

import (
	"context"
	"fmt"
	"reflect"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	log "wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
)

graphName  = "topology"

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

// Interfaces must set their own key if they want to manage their own keys
func (a *ArangoConn) Add(i DBObject) error {
	col, err := a.findCollection(i.GetType())
	if err != nil {
		return err
	}

	_, err = col.CreateDocument(context.Background(), i)
	return err
}

func (a *ArangoConn) Update(i DBObject) error {
	col, err := a.findCollection(i.GetType())
	if err != nil {
		return err
	}

	_, err = col.UpdateDocument(context.Background(), i.GetKey(), i)
	return err
}

func (a *ArangoConn) Read(i DBObject) error {
	col, err := a.findCollection(i.GetType())
	if err != nil {
		return err
	}

	_, err = col.ReadDocument(context.Background(), i.GetKey(), i)
	if err != nil {
		return err
	}
	return err
}

func (a *ArangoConn) Delete(i DBObject) error {
	col, err := a.findCollection(i.GetType())
	if err != nil {
		return err
	}

	_, err = col.RemoveDocument(context.Background(), i.GetKey())
	return err
}

func (a *ArangoConn) Query(q string, obj interface{}) ([]interface{}, error) {
	err := a.db.ValidateQuery(context.Background(), q)
	if err != nil {
		return nil, err
	}

	cursor, err := a.db.Query(context.Background(), q, nil)
	if err != nil {
		return nil, err
	}

	i := make([]interface{}, 0)

	defer cursor.Close()
	t := reflect.TypeOf(obj)
	if t == nil {
		t = reflect.TypeOf(map[string]interface{}{})
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for {
		var obj interface{}
		if t.Kind() != reflect.Struct {
			obj = reflect.New(t).Elem().Interface()
		} else {
			obj = reflect.New(t).Interface()
		}
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

func (a *ArangoConn) findCollection(n string) (driver.Collection, error) {
	if n == "" {
		return nil, fmt.Errorf("Collection name must be defined")
	}
	val, ok := a.cols[n]
	if !ok {
		return val, fmt.Errorf("Could not find collection: %q", n)
	}
	return val, nil
}
