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

// dedupeIGPPrefix removes duplicate prefix entries in the ls_prefix collection
// ISIS generates both Level-1 and Level-2 entries when prefixes are re-advertised
// This function keeps only the original prefix (Level 1, r_flag: false) and removes
// re-advertised duplicates (Level 2, r_flag: true) to ensure prefixes are attached
// to their true originating routers
func (a *arangoDB) dedupeIGPPrefix() error {
	ctx := context.TODO()

	// Query to find duplicate prefixes (same prefix, prefix_len, domain_id)
	// For each duplicate set, remove entries with higher protocol_id or r_flag: true
	dupQuery := fmt.Sprintf(`
		FOR p IN %s
		LET prefix_key = CONCAT(p.prefix, "_", p.prefix_len, "_", p.domain_id)
		COLLECT pk = prefix_key INTO duplicates = {
			_key: p._key,
			protocol_id: p.protocol_id,
			igp_router_id: p.igp_router_id,
			prefix: p.prefix,
			prefix_len: p.prefix_len,
			r_flag: p.prefix_attr_tlvs.flags.r_flag
		}
		FILTER LENGTH(duplicates) > 1
		LET originals = (
			FOR d IN duplicates
			FILTER d.r_flag == false OR d.protocol_id == 1
			RETURN d
		)
		LET readvertised = (
			FOR d IN duplicates
			FILTER d.r_flag == true OR d.protocol_id == 2
			RETURN d
		)
		FILTER LENGTH(originals) > 0 AND LENGTH(readvertised) > 0
		FOR reAdv IN readvertised
		RETURN {
			_key: reAdv._key,
			prefix: reAdv.prefix,
			prefix_len: reAdv.prefix_len,
			protocol_id: reAdv.protocol_id,
			igp_router_id: reAdv.igp_router_id,
			r_flag: reAdv.r_flag
		}
	`, "ls_prefix")

	glog.V(6).Infof("ISIS prefix deduplication query: %s", dupQuery)

	cursor, err := a.db.Query(ctx, dupQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to execute prefix deduplication query: %w", err)
	}
	defer cursor.Close()

	collection, err := a.db.Collection(ctx, "ls_prefix")
	if err != nil {
		return fmt.Errorf("failed to get ls_prefix collection: %w", err)
	}

	count := 0
	for {
		var doc map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return fmt.Errorf("error reading duplicate prefix document: %w", err)
		}

		prefixKey, _ := doc["_key"].(string)
		prefix, _ := doc["prefix"].(string)
		prefixLen, _ := doc["prefix_len"].(float64)
		protocolID, _ := doc["protocol_id"].(float64)
		routerID, _ := doc["igp_router_id"].(string)

		glog.Infof("Removing re-advertised duplicate prefix: %s/%d from router %s (protocol_id: %d)",
			prefix, int(prefixLen), routerID, int(protocolID))

		if _, err := collection.RemoveDocument(ctx, prefixKey); err != nil {
			if !driver.IsNotFound(err) {
				return fmt.Errorf("failed to remove duplicate prefix %s: %w", prefixKey, err)
			}
		}
		count++
	}

	if count > 0 {
		glog.Infof("Removed %d re-advertised prefix duplicates", count)
	}

	return nil
}

// runDeduplication runs the deduplication process and logs results
func (a *arangoDB) runDeduplication() error {
	glog.Info("Starting IGP deduplication process...")

	if err := a.dedupeIGPNode(); err != nil {
		return fmt.Errorf("node deduplication failed: %w", err)
	}

	if err := a.dedupeIGPPrefix(); err != nil {
		return fmt.Errorf("prefix deduplication failed: %w", err)
	}

	glog.Info("IGP deduplication completed successfully")
	return nil
}
