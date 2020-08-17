package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/golang/glog"
	"github.com/sbezverk/jalapeno-gateway/pkg/types"
)

var (
	srvAddr string
	dbname  string
	dbuser  string
	dbpass  string
	prefix  string
	rt      string
)

func init() {
	//flag.StringVar(&srvAddr, "server", "http://localhost:8529", "{dns name}:port or X.X.X.X:port of the graph database")
	flag.StringVar(&srvAddr, "server", "http://arangodb.jalapeno:8529", "{dns name}:port or X.X.X.X:port of the graph database")
	flag.StringVar(&dbname, "db-name", "jalapeno", "DB name")
	flag.StringVar(&dbuser, "db-user", "root", "DB User name")
	flag.StringVar(&dbpass, "db-pass", "jalapeno", "DB User's password")
	flag.StringVar(&prefix, "prefix-collection", "L3VPNPrefix", "Collection with L3VPN Prefixes inofmration")
	flag.StringVar(&rt, "rt-collection", "L3VPN_RT", "Resulting collection RT and links to corresponding prefixes")
}

// RTRecord defines route target record
type RTRecord struct {
	ID       string            `json:"_id,omitempty"`
	Key      string            `json:"_key,omitempty"`
	RT       string            `json:"RT,omitempty"`
	Prefixes map[string]string `json:"Prefixes,omitempty"`
}

func main() {
	flag.Parse()
	_ = flag.Set("logtostderr", "true")

	a, err := newArangodb(srvAddr, dbuser, dbpass, dbname, prefix, rt)
	if err != nil {
		glog.Errorf("failed to instantiate new Arangodb connection with error: %+v", err)
		os.Exit(1)
	}
	if err := updateRT(a); err != nil {
		glog.Errorf("failed to update RT collection %s with error: %+v", rt, err)
		os.Exit(1)
	}
	glog.Infof("Updating RT collection %s succeeded", rt)

	os.Exit(0)
}

var (
	arangoDBConnectTimeout = time.Duration(time.Second * 10)
)

type arangodb struct {
	user   string
	pass   string
	dbName string
	conn   driver.Connection
	client driver.Client
	db     driver.Database
	prefix driver.Collection
	rt     driver.Collection
}

func updateRT(a *arangodb) error {
	c, err := a.prefix.Count(context.TODO())
	if err != nil {
		return err
	}
	glog.Infof("Prefix collection %s contains %d record", a.prefix.Name(), c)
	ctx := context.Background()
	query := fmt.Sprintf("FOR d IN %s RETURN d", a.prefix.Name())
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var doc types.SRv6L3Record
		meta, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		// TODO for efficiency add multiple worker model
		if err := processPrefixRT(ctx, a, a.prefix.Name(), meta.Key, doc.ExtComm); err != nil {
			return err
		}

	}

	return nil
}

func processPrefixRT(ctx context.Context, a *arangodb, prefix string, key string, extComm []string) error {
	for _, ext := range extComm {
		if !strings.HasPrefix(ext, "rt=") {
			continue
		}
		rt := strings.TrimPrefix(ext, "rt=")
		// glog.Infof("for prefix key: %s found route target: %s", key, rt)
		found, err := a.rt.DocumentExists(ctx, rt)
		if err != nil {
			return err
		}
		if found {
			glog.Infof("route target: %s exists in rt collection %s", rt, a.rt.Name())
			rtr := &RTRecord{}
			_, err := a.rt.ReadDocument(ctx, rt, rtr)
			if err != nil {
				glog.Errorf("read doc error: %+v", err)
			}
			if _, ok := rtr.Prefixes[prefix+"/"+key]; ok {
				continue
			}
			rtr.Prefixes[prefix+"/"+key] = key
			b, err := json.Marshal(rtr)
			if err != nil {
				glog.Errorf("marshal error: %+v", err)
				return err
			}
			glog.Infof("resulting RT record: %s", string(b))
			if _, err := a.rt.UpdateDocument(ctx, rt, rtr); err != nil {
				glog.Errorf("update doc error: %+v", err)
				return err
			}
		} else {
			glog.Infof("route target: %s does not exist in rt collection %s", rt, a.rt.Name())
			rtr := &RTRecord{
				ID:  a.rt.Name() + "/" + rt,
				Key: rt,
				RT:  rt,
				Prefixes: map[string]string{
					prefix + "/" + key: key,
				},
			}
			if _, err := a.rt.CreateDocument(ctx, rtr); err != nil {
				glog.Errorf("create doc error: %+v", err)
				return err
			}
		}
	}

	return nil
}

func newArangodb(addr, user, pass, dbName, prefix, rt string) (*arangodb, error) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{addr},
	})
	if err != nil {
		return nil, err
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(user, pass),
	})
	if err != nil {
		return nil, err
	}
	a := &arangodb{}
	a.conn = conn
	a.client = c
	ctx, cancel := context.WithTimeout(context.TODO(), arangoDBConnectTimeout)
	defer cancel()
	db, err := c.Database(ctx, dbName)
	if err != nil {
		return nil, err
	}
	a.db = db

	// Prefix collection is mandatory, if it does not exist, there is nothing to do.
	if found, err := a.db.CollectionExists(context.TODO(), prefix); err != nil || !found {
		return nil, err
	}
	a.prefix, err = a.db.Collection(context.TODO(), prefix)
	if err != nil {
		return nil, err
	}
	// rt collection might or might not exist, if it does not exist, need to create it
	retries := 0
	for {
		found, err := a.db.CollectionExists(context.TODO(), rt)
		if err != nil {
			return nil, err
		}
		if !found {
			// Required collection does not exists, need to create it and wait until it becomes available
			if a.rt, err = a.db.CreateCollection(context.TODO(), rt, &driver.CreateCollectionOptions{}); err == nil {
				break
			}
			time.Sleep(time.Second * 1)
			if retries+1 > 5 {
				return nil, fmt.Errorf("Unable to create collection %s with error: %+v", rt, err)
			}
			retries++
			continue
		}
		a.rt, err = a.db.Collection(context.TODO(), rt)
		if err == nil {
			break
		}
	}
	a.user = user
	a.pass = pass
	a.dbName = dbName

	return a, nil
}
