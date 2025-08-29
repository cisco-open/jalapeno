package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

// Find the igp_node that corresponds to the ls_srv6_sid and add or append the srv6 sid to the igp_node document
func (a *arangoDB) processLSSRv6SID(ctx context.Context, key, id string, e *message.LSSRv6SID) error {
	query := fmt.Sprintf(`
		FOR l IN %s
		FILTER l.igp_router_id == @routerId
		FILTER l.domain_id == @domainId
		RETURN l`,
		a.igpNode.Name(),
	)

	bindVars := map[string]interface{}{
		"routerId": e.IGPRouterID,
		"domainId": e.DomainID,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to execute IGP node query: %w", err)
	}
	defer cursor.Close()

	var igpNode igpNode
	meta, err := cursor.ReadDocument(ctx, &igpNode)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			return nil
		}
		return fmt.Errorf("error reading IGP node document: %w", err)
	}

	glog.V(5).Infof("Processing IGP node %s with SRv6 SID %s", meta.Key, e.SRv6SID)

	newsid := SID{
		SRv6SID:              e.SRv6SID,
		SRv6EndpointBehavior: e.SRv6EndpointBehavior,
		SRv6BGPPeerNodeSID:   e.SRv6BGPPeerNodeSID,
		SRv6SIDStructure:     e.SRv6SIDStructure,
	}

	// Check if SID already exists by comparing SRv6SID field
	for _, existingSID := range igpNode.SIDS {
		if existingSID.SRv6SID == newsid.SRv6SID {
			glog.V(5).Infof("SRv6 SID %s already exists in IGP node document", e.SRv6SID)
			return nil
		}
	}

	// Append new SID and update document
	igpNode.SIDS = append(igpNode.SIDS, newsid)
	if _, err := a.igpNode.UpdateDocument(ctx, meta.Key, igpNode); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to update IGP node: %w", err)
		}
		glog.V(5).Infof("Conflict while updating IGP node %s", meta.Key)
	}

	return nil
}

// Find the ls_srv6_sid that corresponds to the igp_node and add the srv6 sid to the igp_node document
func (a *arangoDB) findSrv6SID(ctx context.Context, key string, e *message.LSNode) error {
	query := fmt.Sprintf(`
		FOR l IN %s
		FILTER l.igp_router_id == @routerId
		FILTER l.srv6_sid != null
		RETURN l`,
		a.lssrv6sid.Name(),
	)

	bindVars := map[string]interface{}{
		"routerId": e.IGPRouterID,
	}

	ncursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to execute SRv6 SID query: %w", err)
	}
	defer ncursor.Close()

	var lp message.LSSRv6SID
	_, err = ncursor.ReadDocument(ctx, &lp)
	if err != nil {
		if !driver.IsNoMoreDocuments(err) {
			return fmt.Errorf("error reading SRv6 SID document: %w", err)
		}
		return nil // No matching documents found
	}

	newsid := SID{
		SRv6SID:              lp.SRv6SID,
		SRv6EndpointBehavior: lp.SRv6EndpointBehavior,
		SRv6BGPPeerNodeSID:   lp.SRv6BGPPeerNodeSID,
		SRv6SIDStructure:     lp.SRv6SIDStructure,
	}

	// Get current IGP node document to check existing SIDs
	var igpNode igpNode
	_, err = a.igpNode.ReadDocument(ctx, e.Key, &igpNode)
	if err != nil {
		return fmt.Errorf("failed to read IGP node document: %w", err)
	}

	// Check if SID already exists by comparing SRv6SID field
	for _, existingSID := range igpNode.SIDS {
		if existingSID.SRv6SID == newsid.SRv6SID {
			glog.V(5).Infof("SRv6 SID %s already exists in IGP node document", lp.SRv6SID)
			return nil
		}
	}

	// Append new SID and update document
	igpNode.SIDS = append(igpNode.SIDS, newsid)
	if _, err := a.igpNode.UpdateDocument(ctx, e.Key, igpNode); err != nil {
		return fmt.Errorf("failed to update IGP node with SRv6 SID: %w", err)
	}

	return a.dedupeIgpNode()
}
