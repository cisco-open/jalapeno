package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

// Find the igp_node that corresponds to the ls_prefix and add the prefix sid to the igp_node document
func (a *arangoDB) processPrefixSID(ctx context.Context, key, id string, e message.LSPrefix) error {
	query := fmt.Sprintf(`
		FOR l IN %s
		FILTER l.igp_router_id == @routerId
		RETURN l`,
		a.igpNode.Name(),
	)

	bindVars := map[string]interface{}{
		"routerId": e.IGPRouterID,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to execute IGP node query: %w", err)
	}
	defer cursor.Close()

	obj := srObject{
		PrefixAttrTLVs: e.PrefixAttrTLVs,
	}

	for {
		var ln igpNode
		meta, err := cursor.ReadDocument(ctx, &ln)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return fmt.Errorf("error reading IGP node document: %w", err)
		}

		glog.V(6).Infof("Updating IGP node: %s with prefix SID %v", meta.Key, e.PrefixAttrTLVs.LSPrefixSID)

		if _, err := a.igpNode.UpdateDocument(ctx, meta.Key, &obj); err != nil {
			if !driver.IsConflict(err) {
				return fmt.Errorf("failed to update IGP node: %w", err)
			}
			glog.V(5).Infof("Conflict while updating IGP node %s", meta.Key)
		}
	}

	return nil
}

// Find the ls_prefix that corresponds to the igp_node and add the prefix sid to the igp_node document
func (a *arangoDB) findPrefixSID(ctx context.Context, key string, e *message.LSNode) error {
	query := fmt.Sprintf(`
		FOR l IN %s
		FILTER l.igp_router_id == @routerId
		FILTER l.prefix_attr_tlvs.ls_prefix_sid != null
		RETURN l`,
		a.lsprefix.Name(),
	)

	bindVars := map[string]interface{}{
		"routerId": e.IGPRouterID,
	}

	ncursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to execute prefix SID query: %w", err)
	}
	defer ncursor.Close()

	var lp message.LSPrefix
	_, err = ncursor.ReadDocument(ctx, &lp)
	if err != nil {
		if !driver.IsNoMoreDocuments(err) {
			return fmt.Errorf("error reading prefix SID document: %w", err)
		}
		return nil // No matching documents found
	}

	obj := srObject{
		PrefixAttrTLVs: lp.PrefixAttrTLVs,
	}

	if _, err := a.igpNode.UpdateDocument(ctx, e.Key, &obj); err != nil {
		return fmt.Errorf("failed to update IGP node with prefix SID: %w", err)
	}

	return a.dedupeIgpNode()
}
