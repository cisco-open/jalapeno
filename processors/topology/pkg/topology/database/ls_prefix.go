package database

import (
	"fmt"

	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/bgpls"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

const LSPrefixName = "LSPrefixDemo"

type LSPrefix struct {
	Key                  string                        `json:"_key,omitempty"`
	Timestamp            string                        `json:"timestamp,omitempty"`
	IGPRouterID          string                        `json:"igp_router_id,omitempty"`
	RouterID             string                        `json:"router_id,omitempty"`
	Prefix               string                        `json:"prefix,omitempty"`
	Length               int32                         `json:"length,omitempty"`
	Protocol             string                        `json:"protocol,omitempty"`
	ProtocolID           base.ProtoID                  `json:"protocol_id,omitempty"`
	MTID                 uint16                        `json:"mtid,omitempty"`
	OSPFRouteType        uint8                         `json:"ospf_route_type,omitempty"`
	IGPFlags             uint8                         `json:"igp_flags,omitempty"`
	IGPRouteTag          []uint32                      `json:"route_tag,omitempty"`
	IGPExtRouteTag       []uint64                      `json:"ext_route_tag,omitempty"`
	OSPFFwdAddr          string                        `json:"ospf_fwd_addr,omitempty"`
	IGPMetric            uint32                        `json:"igp_metric,omitempty"`
	PrefixSID            []*sr.PrefixSIDTLV            `json:"prefix_sid,omitempty"`
	PrefixAttrFlags      uint8                         `json:"prefix_attr_flags,omitempty"`
	SRv6Locator          []*srv6.LocatorTLV            `json:"srv6_locator,omitempty"`
	FlexAlgoPrefixMetric []*bgpls.FlexAlgoPrefixMetric `json:"flex_algo_prefix_metric,omitempty"`
}

func (r LSPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *LSPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *LSPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.IGPRouterID != "" {
		ret = fmt.Sprintf("%s_%s_%d", r.IGPRouterID, r.Prefix, r.Length)
		err = nil
	}
	return ret, err
}

func (r LSPrefix) GetType() string {
	return LSPrefixName
}
