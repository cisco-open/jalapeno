package database

import (
	//"fmt"
	"strconv"
	"github.com/sbezverk/gobmp/pkg/srv6"
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/bgpls"
)

const LSLinkName = "LSLink"

type LSLink struct {
	LocalRouterKey        string           			`json:"_from,omitempty"`
	RemoteRouterKey       string           			`json:"_to,omitempty"`
	Key                   string           			`json:"_key,omitempty"`
	Timestamp       	  string           			`json:"timestamp,omitempty"`
	LocalRouterID         string           			`json:"local_router_id,omitempty"`
	RemoteRouterID        string           			`json:"remote_router_id,omitempty"`
	LocalLinkID           uint32           			`json:"local_link_id,omitempty"`
	RemoteLinkID          uint32           			`json:"remote_link_id,omitempty"`
	LocalLinkIP           []string           		`json:"local_interface_ip,omitempty"`
	RemoteLinkIP          []string           		`json:"remote_interface_ip,omitempty"`
	IsIPv4                bool                      `json:"is_ipv4"`
	IGPRouterID           string           			`json:"local_igp_id,omitempty"`
	RemoteIGPRouterID     string           			`json:"remote_igp_id,omitempty"`
	LocalNodeASN          uint32           			`json:"local_node_asn,omitempty"`
	RemoteNodeASN         uint32           			`json:"remote_node_asn,omitempty"`
	Protocol              string           			`json:"protocol,omitempty"`
	ProtocolID            base.ProtoID     			`json:"protocol_id,omitempty"`
	MTID                  []uint16         			`json:"mtid,omitempty"`
	IGPMetric             uint32           			`json:"igp_metric,omitempty"`
	AdminGroup            uint32           			`json:"admin_group,omitempty"`
	MaxLinkBW             uint32           			`json:"max_link_BW,omitempty"`
	MaxResvBW             uint32           			`json:"max_resv_bw,omitempty"`
	UnResvBW              []uint32         			`json:"unresv_bw,omitempty"`
	TEDefaultMetric       uint32           			`json:"te_metric,omitempty"`
	LinkProtection        uint16           			`json:"link_protection,omitempty"`
	MPLSProtoMask         uint8            			`json:"mpls_proto_mask,omitempty"`
	SRLG                  []uint32         			`json:"srlg"`
	LinkName              string           			`json:"link_name,omitempty"`
	LSAdjacencySID        []map[string]int 			`json:"adjacency_sid,omitempty"`
	SRv6EndXSID           []*srv6.EndXSIDTLV 		`json:"srv6_end_x_sid,omitempty"`
	LinkMSD               []*base.MSDTV    			`json:"link_msd,omitempty"`
	AppSpecLinkAttr       []*bgpls.AppSpecLinkAttr  `json:"app_spec_link_attr,omitempty"`
	UnidirLinkDelay       uint32           			`json:"unidir_link_delay"`
	UnidirLinkDelayMinMax []uint32         			`json:"unidir_link_delay_min_max"`
	UnidirDelayVariation  uint32           			`json:"unidir_delay_variation"`
	UnidirPacketLoss      uint32           			`json:"unidir_packet_loss"`
	UnidirResidualBW      uint32           			`json:"unidir_residual_bw"`
	UnidirAvailableBW     uint32           			`json:"unidir_available_bw"`
	UnidirBWUtilization   uint32           			`json:"unidir_bw_utilization"`
}

func (l LSLink) GetKey() (string, error) {
	if l.Key == "" {
		return l.makeKey()
	}
	return l.Key, nil
}

func (l *LSLink) SetKey() error {
	k, err := l.makeKey()
	if err != nil {
		return err
	}
	l.Key = k
	return nil
}

//func (l *LSLink) makeKey() (string, error) {
//	err := ErrKeyInvalid
//	ret := ""
	//if l.LocalInterfaceIP != "" && l.RemoteInterfaceIP != "" {
//	ret = fmt.Sprintf("%s_%s_%s_%s", l.IGPRouterID, l.LocalLinkIP[], l.LocalLinkID, l.RemoteLinkIP[], l.RemoteLinkID, l.RemoteIGPRouterID)
//	err = nil
	//}
//	return ret, err
//}

func (l *LSLink) makeKey() (string, error) {
	var localIP, remoteIP, localID, remoteID string
	localID = "0"
	remoteID = "0"
	switch obj.MTID {
	case 0:
		localIP = "127.0.0.1"
		remoteIP = "127.0.0.1"
	case 2:
		localIP = "::1"
		remoteIP = "::1"
	default:
		localIP = "unknown-mt-id"
		remoteIP = "unknown-mt-id"
	}
	if len(obj.LocalLinkIP) != 0 {
		localIP = obj.LocalLinkIP[0]
	}
	if len(obj.RemoteLinkIP) != 0 {
		remoteIP = obj.RemoteLinkIP[0]
	}
	localID = strconv.Itoa(int(obj.LocalLinkID))
	remoteID = strconv.Itoa(int(obj.RemoteLinkID))

	return l.IGPRouterID + "_" + localIP + "_" + localID + l.RemoteIGPRouterID + "_" + remoteIP + "_" + remoteID, nil
}

func (l LSLink) GetType() string {
	return LSLinkName
}

func (l *LSLink) SetEdge(to DBObject, from DBObject) error {
	var err error
	l.RemoteRouterID, err = GetID(to)
	if err != nil {
		return err
	}
	l.LocalRouterID, err = GetID(from)
	if err != nil {
		return err
	}
	return nil
}
