package database

import (
	"fmt"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/srv6"
	"github.com/sbezverk/gobmp/pkg/bgpls"
)

const LSNodeName = "LSNode"

type LSNode struct {
	Key                  string                    `json:"_key,omitempty"`
	Name                 string                    `json:"Name,omitempty"`
	Timestamp       	 string                    `json:"timestamp,omitempty"`
	IGPRouterID          string                    `json:"igp_router_id,omitempty"`
	RouterID             string                    `json:"router_id,omitempty"`
	ASN                  int32                     `json:"asn,omitempty"`
	MTID                 []uint16			       `json:"mtid,omitempty"`
	OSPFAreaID           string                    `json:"ospf_area_id,omitempty"`
	ISISAreaID           string                    `json:"isis_area_id,omitempty"`
	Protocol             string                    `json:"protocol,omitempty"`
	ProtocolID           base.ProtoID              `json:"protocol_id,omitempty"`
	Flags                uint8                     `json:"flags,omitempty"`			
	SRGBStart            int                       `json:"srgb_start,omitempty"`
	SRGBRange            uint32                    `json:"srgb_range,omitempty"`
	SRCapabilityFlags    uint8                     `json:"sr_capability_flags,omitempty"`
	SRAlgorithm          []int                     `json:"sr_algorithm,omitempty"`
	SRLocalBlock         *sr.LocalBlock            `json:"sr_localBlock,omitempty"`
	SRv6CapabilitiesTLV  *srv6.CapabilityTLV       `json:"srv6_capabilities_tlv,omitempty"`	
	NodeMSD              []*base.MSDTV        	   `json:"node_msd,omitempty"`
	FlexAlgoDefinition   *bgpls.FlexAlgoDefinition `json:"flex_algo_definition,omitempty"`
}

func (r LSNode) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *LSNode) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *LSNode) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.IGPRouterID != "" {
		ret = fmt.Sprintf("%s", r.IGPRouterID)
		err = nil
	}
	return ret, err
}

func (r LSNode) GetType() string {
	return LSNodeName
}
