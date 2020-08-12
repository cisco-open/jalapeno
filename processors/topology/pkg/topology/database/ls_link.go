package database

import (
	"fmt"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

const LSLinkName = "LSLink"

type LSLink struct {
        LocalRouterKey     string `json:"_from,omitempty"`
        RemoteRouterKey    string `json:"_to,omitempty"`
        Key                string `json:"_key,omitempty"`
        LocalRouterID      string `json:"LocalRouterID,omitempty"`
        LocalIGPID         string `json:"LocalIGPID,omitempty"`
        RemoteRouterID     string `json:"RemoteRouterID,omitempty"`
        RemoteIGPID        string `json:"RemoteIGPID,omitempty"`
        Protocol           string `json:"Protocol,omitempty"`
        ASN                uint32 `json:"ASN,omitempty"`
        LocalInterfaceIP   string `json:"FromInterfaceIP,omitempty"`
        RemoteInterfaceIP  string `json:"ToInterfaceIP,omitempty"`
        IGPMetric          uint32 `json:"IGPMetric,omitempty"`
        TEMetric           uint32 `json:"TEMetric,omitempty"`
        AdminGroup         uint32 `json:"AdminGroup,omitempty"`
        MaxLinkBW          uint32 `json:"MaxLinkBW,omitempty"`
        MaxResvBW          uint32 `json:"MaxResvBW,omitempty"`
        UnResvBW           []uint32 `json:"UnResvBW,omitempty"`
        LinkProtection     uint16 `json:"LinkProtection,omitempty"`
        LinkName           string `json:"LinkName,omitempty"`
        SRLG               []uint32 `json:"SRLG"`
        UnidirLinkDelay        uint32   `json:"unidir_link_delay"`
	UnidirLinkDelayMinMax  []uint32 `json:"unidir_link_delay_min_max"`
	UnidirDelayVariation   uint32   `json:"unidir_delay_variation"`
	UnidirPacketLoss       uint32   `json:"unidir_packet_loss"`
	UnidirResidualBW       uint32   `json:"unidir_residual_bw"`
	UnidirAvailableBW      uint32   `json:"unidir_available_bw"`
	UnidirBWUtilization    uint32   `json:"unidir_bw_utilization"`
        AdjacencySID       []map[string]int `json:"AdjacencySID,omitempty"`
        //SRv6BGPPeerNodeSID SRv6BGPPeerNodeSID `json:"SRv6BGPPeerNodeSID"`
	SRv6EndXSID        *srv6.ENDXSIDTLV `json:"SRv6ENDXSID"`
	Timestamp          string `json:"Timestamp"`
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

func (l *LSLink) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if l.LocalInterfaceIP != "" && l.RemoteInterfaceIP != "" {
		ret = fmt.Sprintf("%s_%s_%s_%s", l.LocalIGPID, l.LocalInterfaceIP, l.RemoteInterfaceIP, l.RemoteIGPID)
		err = nil
	}
	return ret, err
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
