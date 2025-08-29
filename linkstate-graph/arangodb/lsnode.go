package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/message"
)

// Find and add ls_node entries to the igp_node collection
func (a *arangoDB) processIgpNode(ctx context.Context, key string, e *message.LSNode) error {
	if e.ProtocolID == base.BGP {
		// EPE Case cannot be processed because LS Node collection does not have BGP routers
		glog.V(6).Infof("Skipping BGP protocol node: %s", e.Key)
		return nil
	}

	// Query for existing node
	query := fmt.Sprintf(`
		FOR l IN %s
		FILTER l._key == @nodeKey
		RETURN l`,
		a.lsnode.Name(),
	)

	bindVars := map[string]interface{}{
		"nodeKey": e.Key,
	}

	ncursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to query ls_node: %w", err)
	}
	defer ncursor.Close()

	var sn igpNode
	meta, err := ncursor.ReadDocument(ctx, &sn)
	if err != nil && !driver.IsNoMoreDocuments(err) {
		return fmt.Errorf("error reading ls_node document: %w", err)
	}

	// Try to create new document
	_, err = a.igpNode.CreateDocument(ctx, &sn)
	if err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create igp_node document: %w", err)
		}

		glog.V(5).Infof("Updating existing igp_node: %s with area_id %s", sn.Key, e.AreaID)

		// Document exists, update with SIDs
		if err := a.updateExistingNode(ctx, sn.Key, e); err != nil {
			return fmt.Errorf("failed to update existing node: %w", err)
		}

		return nil
	}

	// Process new document
	if err := a.processNewNode(ctx, meta.Key, e); err != nil {
		return fmt.Errorf("failed to process new node: %w", err)
	}

	return nil
}

// Helper function to update existing node
func (a *arangoDB) updateExistingNode(ctx context.Context, key string, e *message.LSNode) error {
	// Find and update prefix SID
	if err := a.findPrefixSID(ctx, key, e); err != nil {
		return fmt.Errorf("failed to find prefix SID: %w", err)
	}

	// Find and update SRv6 SID
	if err := a.findSrv6SID(ctx, key, e); err != nil {
		return fmt.Errorf("failed to find SRv6 SID: %w", err)
	}

	// Update document with latest info
	if _, err := a.igpNode.UpdateDocument(ctx, key, e); err != nil {
		return fmt.Errorf("failed to update igp_node document: %w", err)
	}

	return nil
}

// Helper function to process new node
func (a *arangoDB) processNewNode(ctx context.Context, key string, e *message.LSNode) error {
	// Process IGP node recursively if needed
	if err := a.processIgpNode(ctx, key, e); err != nil {
		glog.Errorf("Failed to process igp_node %s with error: %+v", key, err)
		return err
	}

	// Process IGP domain
	if err := a.processIgpDomain(ctx, key, e); err != nil {
		return fmt.Errorf("failed to process IGP domain: %w", err)
	}

	return nil
}

// BGP-LS generates a level-1 and a level-2 entry for level-1-2 nodes
// remove duplicate entries in the igp_node collection
func (a *arangoDB) dedupeIgpNode() error {
	ctx := context.TODO()
	dup_query := "LET duplicates = ( FOR d IN " + a.igpNode.Name() +
		" COLLECT id = d.igp_router_id, domain = d.domain_id WITH COUNT INTO count " +
		" FILTER count > 1 RETURN { id: id, domain: domain, count: count }) " +
		"FOR d IN duplicates FOR m IN igp_node " +
		"FILTER d.id == m.igp_router_id filter d.domain == m.domain_id RETURN m "
	pcursor, err := a.db.Query(ctx, dup_query, nil)
	glog.Infof("dedup query: %+v", dup_query)
	if err != nil {
		return err
	}
	defer pcursor.Close()
	for {
		var doc duplicateNode
		dupe, err := pcursor.ReadDocument(ctx, &doc)

		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return err
			}
			break
		}
		glog.Infof("Got doc with key '%s' from query\n", dupe.Key)

		if doc.ProtocolID == 1 {
			glog.Infof("remove level-1 duplicate node: %s + igp id: %s protocol id: %v +  ", doc.Key, doc.IGPRouterID, doc.ProtocolID)
			if _, err := a.igpNode.RemoveDocument(ctx, doc.Key); err != nil {
				if !driver.IsConflict(err) {
					return err
				}
			}
		}
		if doc.ProtocolID == 2 {
			update_query := "for l in " + a.igpNode.Name() + " filter l._key == " + "\"" + doc.Key + "\"" +
				" UPDATE l with { protocol: " + "\"" + "ISIS Level 1-2" + "\"" + " } in " + a.igpNode.Name() + ""
			cursor, err := a.db.Query(ctx, update_query, nil)
			glog.Infof("update query: %s ", update_query)
			if err != nil {
				return err
			}
			defer cursor.Close()
		}
	}
	return nil
}

// Nov 10 2024 - find ipv6 lsnode's ipv4 bgp router-id
func (a *arangoDB) processbgp6(ctx context.Context, key, id string, e *message.PeerStateChange) error {
	query := "for l in  " + a.igpNode.Name() +
		" filter l.router_id == " + "\"" + e.RemoteBGPID + "\""
	query += " return l"
	pcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer pcursor.Close()
	for {
		var ln igpNode
		nl, err := pcursor.ReadDocument(ctx, &ln)
		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return err
			}
			break
		}
		//glog.Infof("igp_node: %s + peer %v +  ", ln.Key, e.RemoteBGPID)

		obj := peerObject{
			BGPRouterID: e.RemoteBGPID,
		}
		glog.Infof("igp_node: %s + peer %v +  object: %+v", nl.Key, e.RemoteBGPID, obj)
		if _, err := a.igpNode.UpdateDocument(ctx, nl.Key, &obj); err != nil {
			if !driver.IsConflict(err) {
				return err
			}
		}
	}
	return nil
}

// when a new igp domain is detected, create a new entry in the igp_domain collection
func (a *arangoDB) processIgpDomain(ctx context.Context, key string, e *message.LSNode) error {
	if e.ProtocolID == base.BGP {
		// EPE Case cannot be processed because LS Node collection does not have BGP routers
		return nil
	}
	query := "for l in igp_node insert " +
		"{ _key: CONCAT_SEPARATOR(" + "\"_\", l.protocol_id, l.domain_id, l.asn), " +
		"asn: l.asn, protocol_id: l.protocol_id, domain_id: l.domain_id, protocol: l.protocol } " +
		"into igp_domain OPTIONS { ignoreErrors: true } "
	query += " return l"
	ncursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer ncursor.Close()
	var sn igpNode
	ns, err := ncursor.ReadDocument(ctx, &sn)
	if err != nil {
		if !driver.IsNoMoreDocuments(err) {
			return err
		}
	}

	if _, err := a.igpDomain.CreateDocument(ctx, &sn); err != nil {
		glog.Infof("adding igp_domain: %s with area_id %v ", sn.Key, e.ASN)
		if !driver.IsConflict(err) {
			return err
		}
		if err := a.findPrefixSID(ctx, sn.Key, e); err != nil {
			if err != nil {
				return err
			}
		}
		// The document already exists, updating it with the latest info
		if _, err := a.igpDomain.UpdateDocument(ctx, ns.Key, e); err != nil {
			return err
		}
		return nil
	}
	if err := a.processIgpDomain(ctx, ns.Key, e); err != nil {
		glog.Errorf("Failed to process igp_domain %s with error: %+v", ns.Key, err)
	}
	return nil
}

// processIgpNodeRemoval removes records from the igp_node collection which are referring to deleted LSNode
func (a *arangoDB) processIgpNodeRemoval(ctx context.Context, key string) error {
	query := "FOR d IN " + a.igpNode.Name() +
		" filter d._key == " + "\"" + key + "\""
	query += " return d"
	ncursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer ncursor.Close()

	for {
		var nm igpNode
		m, err := ncursor.ReadDocument(ctx, &nm)
		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return err
			}
			break
		}
		if _, err := a.igpNode.RemoveDocument(ctx, m.ID.Key()); err != nil {
			if !driver.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}
