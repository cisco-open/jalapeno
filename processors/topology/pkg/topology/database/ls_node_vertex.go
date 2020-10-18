package database

import (
	"fmt"
	"strconv"

	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/bgpls"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

const LSNodeVertexName = "LSNodeVertex"

type LSNodeVertex struct {
	Key                 string                      `json:"_key,omitempty"`
	Name                string                      `json:"name,omitempty"`
	Timestamp           string                      `json:"timestamp,omitempty"`
	DomainID            int64                       `json:"domain_id"`
	IGPRouterID         string                      `json:"igp_router_id,omitempty"`
	RouterID            string                      `json:"router_id,omitempty"`
	ASN                 int32                       `json:"asn,omitempty"`
	MTID                []uint16                    `json:"mtid,omitempty"`
	OSPFAreaID          string                      `json:"ospf_area_id,omitempty"`
	ISISAreaID          string                      `json:"isis_area_id,omitempty"`
	Protocol            string                      `json:"protocol,omitempty"`
	ProtocolID          base.ProtoID                `json:"protocol_id,omitempty"`
	NodeFlags           uint8                       `json:"node_flags,omitempty"`
	SRGBStart           int                         `json:"srgb_start,omitempty"`
	SRGBRange           uint32                      `json:"srgb_range,omitempty"`
	SRCapabilityFlags   uint8                       `json:"sr_capability_flags,omitempty"`
	SRAlgorithm         []int                       `json:"sr_algorithm,omitempty"`
	SRLocalBlock        *sr.LocalBlock              `json:"sr_local_block,omitempty"`
	SRv6CapabilitiesTLV *srv6.CapabilityTLV         `json:"srv6_capabilities_tlv,omitempty"`
	NodeMSD             []*base.MSDTV               `json:"node_msd,omitempty"`
	FlexAlgoDefinition  []*bgpls.FlexAlgoDefinition `json:"flex_algo_definition,omitempty"`
}

func (r LSNodeVertex) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *LSNodeVertex) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *LSNodeVertex) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.IGPRouterID != "" {
		//ret = fmt.Sprintf("%s_%s_%s_%s", strconv.Itoa(int(r.ProtocolID)), strconv.Itoa(int(r.DomainID)), r.IGPRouterID)
		ret = fmt.Sprintf("%s_%s_%s", strconv.Itoa(int(r.ProtocolID)), strconv.Itoa(int(r.DomainID)), r.IGPRouterID)
		err = nil
	}
	return ret, err
}

func (r LSNodeVertex) GetType() string {
	return LSNodeVertexName
}
