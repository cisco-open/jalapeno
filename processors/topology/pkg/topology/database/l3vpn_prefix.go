package database

import "fmt"
import "github.com/sbezverk/gobmp/pkg/srv6"

const L3VPNPrefixName = "L3VPNPrefix"

type L3VPNPrefix struct {
        Key             string   `json:"_key,omitempty"`
        RD              string   `json:"RD,omitempty"`
        Prefix          string   `json:"Prefix,omitempty"`
        Length          int32    `json:"Length,omitempty"`
        RouterID        string   `json:"RouterID,omitempty"`
        ControlPlaneID  string   `json:"ControlPlaneID,omitempty"`
        ASN             int32    `json:"ASN,omitempty"`
        VPN_Label       uint32   `json:"VPN_Label,omitempty"`
        SRv6_SID        []srv6.SubTLV   `json:SRv6_SID,omitempty"`
	ExtComm         []string `json:"ExtComm,omitempty"`
        IPv4            bool     `json:"IPv4"`
}

func (p L3VPNPrefix) GetKey() (string, error) {
	if p.Key == "" {
		return p.makeKey()
	}
	return p.Key, nil
}

func (p *L3VPNPrefix) SetKey() error {
	k, err := p.makeKey()
	if err != nil {
		return err
	}
	p.Key = k
	return nil
}

func (p *L3VPNPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if p.RD != "" && p.Prefix != "" && p.Length != 0 {
		ret = fmt.Sprintf("%s_%s_%d", p.RD, p.Prefix, p.Length)
		err = nil
	}
	return ret, err
}

func (p L3VPNPrefix) GetType() string {
	return L3VPNPrefixName
}
