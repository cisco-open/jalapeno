package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// DedupeNode represents a duplicate node for deduplication processing
type DedupeNode struct {
	Key         string `json:"_key,omitempty"`
	DomainID    int64  `json:"domain_id"`
	IGPRouterID string `json:"igp_router_id,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	ProtocolID  int    `json:"protocol_id,omitempty"`
	Name        string `json:"name,omitempty"`
}

// dedupeIGPNode removes duplicate entries in the igp_node collection
// BGP-LS generates a level-1 and a level-2 entry for level-1-2 nodes
// This function removes duplicate entries following the original logic:
// - Remove Level-1 duplicates (protocol_id = 1)
// - Mark Level-2 nodes as "ISIS Level 1-2" when duplicates exist
func (a *arangoDB) dedupeIGPNode() error {
	ctx := context.TODO()

	// Query to find duplicate nodes (same igp_router_id and domain_id)
	dupQuery := fmt.Sprintf(`
		LET duplicates = (
			FOR d IN %s
			COLLECT id = d.igp_router_id, domain = d.domain_id WITH COUNT INTO count
			FILTER count > 1
			RETURN { id: id, domain: domain, count: count }
		)
		FOR d IN duplicates
		FOR m IN %s
		FILTER d.id == m.igp_router_id
		FILTER d.domain == m.domain_id
		RETURN m`, a.config.IGPNode, a.config.IGPNode)

	glog.V(6).Infof("Deduplication query: %s", dupQuery)

	cursor, err := a.db.Query(ctx, dupQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to execute deduplication query: %w", err)
	}
	defer cursor.Close()

	for {
		var doc DedupeNode
		meta, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return fmt.Errorf("error reading duplicate node document: %w", err)
		}

		glog.V(6).Infof("Processing duplicate node with key '%s', IGP ID: %s, Protocol ID: %d",
			meta.Key, doc.IGPRouterID, doc.ProtocolID)

		// Remove Level-1 duplicates (protocol_id = 1)
		if doc.ProtocolID == 1 {
			glog.Infof("Removing Level-1 duplicate node: %s (IGP ID: %s, Protocol ID: %d)",
				meta.Key, doc.IGPRouterID, doc.ProtocolID)

			if _, err := a.igpNode.RemoveDocument(ctx, meta.Key); err != nil {
				if !driver.IsNotFound(err) {
					return fmt.Errorf("failed to remove duplicate Level-1 node %s: %w", meta.Key, err)
				}
			}
		}

		// Update Level-2 nodes to indicate they are Level 1-2 routers
		if doc.ProtocolID == 2 {
			updateQuery := fmt.Sprintf(`
				FOR l IN %s
				FILTER l._key == @key
				UPDATE l WITH { protocol: "ISIS Level 1-2" } IN %s`,
				a.config.IGPNode, a.config.IGPNode)

			bindVars := map[string]interface{}{
				"key": meta.Key,
			}

			glog.V(6).Infof("Updating Level-2 node to Level 1-2: %s", meta.Key)

			updateCursor, err := a.db.Query(ctx, updateQuery, bindVars)
			if err != nil {
				return fmt.Errorf("failed to execute update query for node %s: %w", meta.Key, err)
			}
			updateCursor.Close()
		}
	}

	return nil
}

// runDeduplication runs the deduplication process and logs results
func (a *arangoDB) runDeduplication() error {
	glog.Info("Starting IGP node deduplication process...")

	if err := a.dedupeIGPNode(); err != nil {
		return fmt.Errorf("deduplication failed: %w", err)
	}

	glog.Info("IGP node deduplication completed successfully")
	return nil
}
