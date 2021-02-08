package arangodb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	notifier "github.com/jalapeno/topology/pkg/kafkanotifier"
	"github.com/sbezverk/gobmp/pkg/message"
)

func (a *arangoDB) lsLinkHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.edge.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.edge.Name(), c)
	}
	glog.V(5).Infof("Processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)
	var o message.LSLink
	_, err := a.edge.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a LSLink removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}
		return a.processEdgeRemoval(ctx, obj.Key)
	}
	switch obj.Action {
	case "add":
		fallthrough
	case "update":
		if err := a.processEdge(ctx, obj.Key, &o); err != nil {
			return fmt.Errorf("failed to process action %s for edge %s with error: %+v", obj.Action, obj.Key, err)
		}
	}

	return nil
}

func (a *arangoDB) lsNodeHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.edge.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.edge.Name(), c)
	}
	glog.V(5).Infof("Processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)
	var o message.LSNode
	_, err := a.vertex.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a LSNode removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}
		return a.processVertexRemoval(ctx, obj.Key)
	}
	switch obj.Action {
	case "add":
		fallthrough
	case "update":
		if err := a.processVertex(ctx, obj.Key, &o); err != nil {
			return fmt.Errorf("failed to process action %s for vertex %s with error: %+v", obj.Action, obj.Key, err)
		}
	}

	return nil
}

type lsNodeEdgeObject struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
}

func (a *arangoDB) processEdge(ctx context.Context, key string, e *message.LSLink) error {
	//routerID := e.IGPRouterID
	//if e.ProtocolID == 7 {
	//	routerID = e.BGPRouterID
	//}
	query := "FOR d IN " + a.vertex.Name() +
		" filter d.igp_router_id == " + "\"" + e.IGPRouterID + "\"" +
		" filter d.domain_id == " + strconv.Itoa(int(e.DomainID)) +
		" filter d.protocol_id == " + strconv.Itoa(int(e.ProtocolID))
	query += " return d"
	glog.Infof("1st vertex: %+v, protocol: %+v, linkIP: %+v", e.IGPRouterID, e.Protocol, e.LocalLinkIP)
	lcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer lcursor.Close()
	var ln message.LSNode
	lm, err := lcursor.ReadDocument(ctx, &ln)
	if err != nil {
		if !driver.IsNoMoreDocuments(err) {
			return err
		}
	}

	//remoteRouterID := e.RemoteIGPRouterID
	//if e.ProtocolID == 7 {
	//	remoteRouterID = e.BGPRemoteRouterID
	//}
	query = "FOR d IN " + a.vertex.Name() +
		" filter d.igp_router_id == " + "\"" + e.RemoteIGPRouterID + "\"" +
		" filter d.domain_id == " + strconv.Itoa(int(e.DomainID)) +
		" filter d.protocol_id == " + strconv.Itoa(int(e.ProtocolID))
	query += " return d"
	glog.Infof("2nd vertex: %+v, protocol: %+v, linkIP: %+v", e.RemoteIGPRouterID, e.Protocol, e.LocalLinkIP)
	rcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer rcursor.Close()
	var rn message.LSNode
	rm, err := rcursor.ReadDocument(ctx, &rn)
	if err != nil {
		if !driver.IsNoMoreDocuments(err) {
			return err
		}
	}

	glog.V(6).Infof("Local node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
		ln.ProtocolID, ln.DomainID, ln.IGPRouterID)
	glog.V(6).Infof("Remote node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
		rn.ProtocolID, rn.DomainID, rn.IGPRouterID)

	ne := lsNodeEdgeObject{
		Key:  key,
		From: lm.ID.String(),
		To:   rm.ID.String(),
	}

	if _, err := a.graph.CreateDocument(ctx, &ne); err != nil {
		if !driver.IsConflict(err) {
			return err
		}
		// The document already exists, updating it with the latest info
		if _, err := a.graph.UpdateDocument(ctx, e.Key, &ne); err != nil {
			return err
		}
	}

	return nil
}

func (a *arangoDB) processVertex(ctx context.Context, key string, e *message.LSNode) error {
	// Check if there is an edge with matching to LSNode's e.IGPRouterID, e.AreaID, e.DomainID and e.ProtocolID
	query := "FOR d IN " + a.vertex.Name() +
		" filter d.igp_router_id == " + "\"" + e.IGPRouterID + "\"" +
		" filter d.area_id == " + e.AreaID +
		" filter d.domain_id == " + strconv.Itoa(int(e.DomainID)) +
		" filter d.protocol_id == " + strconv.Itoa(int(e.ProtocolID))
	query += " return d"
	lcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer lcursor.Close()
	var ln message.LSLink
	lm, err := lcursor.ReadDocument(ctx, &ln)
	if err != nil {
		if !driver.IsNoMoreDocuments(err) {
			return err
		}
	}
	// Check if there is a second link LS Link with with matching to LSNode's e.IGPRouterID, e.AreaID, e.DomainID and e.ProtocolID
	query = "FOR d IN " + a.vertex.Name() +
		" filter d.remote_igp_router_id == " + "\"" + e.IGPRouterID + "\"" +
		" filter d.area_id == " + e.AreaID +
		" filter d.domain_id == " + strconv.Itoa(int(e.DomainID)) +
		" filter d.protocol_id == " + strconv.Itoa(int(e.ProtocolID))
	query += " return d"
	rcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer rcursor.Close()
	var rn message.LSNode
	rm, err := rcursor.ReadDocument(ctx, &rn)
	if err != nil {
		if !driver.IsNoMoreDocuments(err) {
			return err
		}
	}

	glog.V(6).Infof("Local link: %s", lm.ID.Key())
	glog.V(6).Infof("Remote link: %s", rm.ID.Key())

	ne := lsNodeEdgeObject{
		Key:  key,
		From: lm.ID.String(),
		To:   rm.ID.String(),
	}

	if _, err := a.graph.CreateDocument(ctx, &ne); err != nil {
		if !driver.IsConflict(err) {
			return err
		}
		// The document already exists, updating it with the latest info
		if _, err := a.graph.UpdateDocument(ctx, e.Key, &ne); err != nil {
			return err
		}
	}

	return nil
}

// processVertexRemoval removes a record from Node's graph collection
// since the key matches in both collections (Nodes and Nodes' Graph) deleting the record directly.
func (a *arangoDB) processVertexRemoval(ctx context.Context, key string) error {
	if _, err := a.graph.RemoveDocument(ctx, key); err != nil {
		if !driver.IsNotFound(err) {
			return err
		}
		glog.Warningf("Document %s/%s was requested to be delete but it does not exist", a.graph.Name(), key)
		return nil
	}

	return nil
}

// processEdgeRemoval removes all documents where removed Edge (LS Link) is referenced in "_to" or "_from"
func (a *arangoDB) processEdgeRemoval(ctx context.Context, key string) error {
	query := "FOR d IN" + a.graph.Name() +
		" filter d._to == " + "\"" + key + "\"" + " OR" + " d._from == " + "\"" + key + "\"" +
		" return d"
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p lsNodeEdgeObject
		meta, err := cursor.ReadDocument(ctx, &p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if _, err := a.graph.RemoveDocument(ctx, meta.ID.Key()); err != nil {
			if !driver.IsNotFound(err) {
				return err
			}
			glog.Warningf("Document %s/%s was requested to be delete but it does not exist", a.graph.Name(), key)
			return nil
		}
	}

	return nil
}
