package database

import "fmt"

const L3VPN_PrefixName = "L3VPN_Prefixes"

type L3VPN_Prefix struct {
        Key             string `json:"_key,omitempty"`
        RD              string `json:"RD,omitempty"`
        Prefix          string `json:"Prefix,omitempty"`
        Length          int    `json:"Length,omitempty"`
        RouterIP        string `json:"RouterIP,omitempty"`
        ASN             string `json:"ASN,omitempty"`
        AdvertisingPeer string `json:"AdvertisingPeer,omitempty"`
        VPN_Label       string `json:"VPN_Label,omitempty"`
        ExtComm         string `json:"ExtComm,omitempty"`
}

func (p L3VPN_Prefix) GetKey() (string, error) {
	if p.Key == "" {
		return p.makeKey()
	}
	return p.Key, nil
}

func (p *L3VPN_Prefix) SetKey() error {
	k, err := p.makeKey()
	if err != nil {
		return err
	}
	p.Key = k
	return nil
}

func (p *L3VPN_Prefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if p.RD != "" && p.Prefix != "" && p.Length != 0 {
		ret = fmt.Sprintf("%s:%s:%d", p.RD, p.Prefix, p.Length)
		err = nil
	}
	return ret, err
}

func (p L3VPN_Prefix) GetType() string {
	return L3VPN_PrefixName
}
