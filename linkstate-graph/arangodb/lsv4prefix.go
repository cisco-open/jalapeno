package arangodb

import (
	"context"
	"fmt"
	"strconv"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/message"
)

// processLSPrefixEdge processes a single ls_prefix entry which is connected to a node
func (a *arangoDB) processLSPrefixEdge(ctx context.Context, key string, p *message.LSPrefix) error {
	//glog.V(9).Infof("processEdge processing lsprefix: %s", l.ID)

	if p.MTID != nil {
		return a.processigpv6PrefixEdge(ctx, key, p)
	}
	// filter out IPv6, ls link, and loopback prefixes
	if p.PrefixLen == 30 || p.PrefixLen == 31 || p.PrefixLen == 32 {
		return nil
	}

	// get remote node from ls_link entry
	lsnode, err := a.getigpv4Node(ctx, p, false)
	if err != nil {
		glog.Errorf("processEdge failed to get remote lsnode %s for link: %s with error: %+v", p.IGPRouterID, p.ID, err)
		return err
	}
	if err := a.createigpv4PrefixEdgeObject(ctx, p, lsnode); err != nil {
		glog.Errorf("processEdge failed to create Edge object with error: %+v", err)
		return err
	}
	//glog.V(9).Infof("processEdge completed processing lsprefix: %s for ls nodes: %s - %s", l.ID, ln.ID, rn.ID)

	return nil
}

// processPrefixRemoval removes a record from Node's graph collection
func (a *arangoDB) processPrefixRemoval(ctx context.Context, key string, action string) error {
	if _, err := a.graphv4.RemoveDocument(ctx, key); err != nil {
		if !driver.IsNotFound(err) {
			return err
		}
		return nil
	}

	return nil
}

func (a *arangoDB) getigpv4Node(ctx context.Context, p *message.LSPrefix, local bool) (*message.LSNode, error) {
	// Need to find ls_node object matching ls_prefix's IGP Router ID
	query := "FOR d IN " + a.igpNode.Name()
	query += " filter d.igp_router_id == " + "\"" + p.IGPRouterID + "\""
	query += " filter d.domain_id == " + strconv.Itoa(int(p.DomainID))

	// If OSPFv2 or OSPFv3, then query must include AreaID
	if p.ProtocolID == base.OSPFv2 || p.ProtocolID == base.OSPFv3 {
		query += " filter d.area_id == " + "\"" + p.AreaID + "\""
	}
	query += " return d"
	//glog.Infof("query: %s", query)
	lcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer lcursor.Close()
	var ln message.LSNode
	i := 0
	for ; ; i++ {
		_, err := lcursor.ReadDocument(ctx, &ln)
		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return nil, err
			}
			break
		}
	}
	if i == 0 {
		return nil, fmt.Errorf("query %s returned 0 results", query)
	}
	if i > 1 {
		return nil, fmt.Errorf("query %s returned more than 1 result", query)
	}

	return &ln, nil
}

func (a *arangoDB) createigpv4PrefixEdgeObject(ctx context.Context, l *message.LSPrefix, ln *message.LSNode) error {
	mtid := 0
	if l.MTID != nil {
		mtid = int(l.MTID.MTID)
	}
	// Node to Prefix direction
	nodeToPrefix := lsGraphObject{
		Key:            l.Key,
		From:           ln.ID,
		To:             l.ID,
		Link:           l.Key,
		ProtocolID:     l.ProtocolID,
		DomainID:       l.DomainID,
		MTID:           uint16(mtid),
		AreaID:         l.AreaID,
		Protocol:       l.Protocol,
		LocalNodeASN:   ln.ASN,
		Prefix:         l.Prefix,
		PrefixLen:      l.PrefixLen,
		PrefixMetric:   l.PrefixMetric,
		PrefixAttrTLVs: l.PrefixAttrTLVs,
	}

	// Prefix to Node direction
	glog.Infof("creating prefix edge object from prefix: %s to node: %s ", l.Key, ln.Key)
	prefixToNode := lsGraphObject{
		Key:            l.Key + "_to_" + ln.Key, // Changed key format
		From:           l.ID,
		To:             ln.ID,
		Link:           l.Key,
		ProtocolID:     l.ProtocolID,
		DomainID:       l.DomainID,
		MTID:           uint16(mtid),
		AreaID:         l.AreaID,
		Protocol:       l.Protocol,
		LocalNodeASN:   ln.ASN,
		Prefix:         l.Prefix,
		PrefixLen:      l.PrefixLen,
		PrefixMetric:   l.PrefixMetric,
		PrefixAttrTLVs: l.PrefixAttrTLVs,
	}

	// Create/Update both directions
	for _, edge := range []*lsGraphObject{&nodeToPrefix, &prefixToNode} {
		if _, err := a.graphv4.CreateDocument(ctx, edge); err != nil {
			if !driver.IsConflict(err) {
				return err
			}
			// The document already exists, updating it with the latest info
			if _, err := a.graphv4.UpdateDocument(ctx, edge.Key, edge); err != nil {
				return err
			}
		}
	}

	return nil
}
