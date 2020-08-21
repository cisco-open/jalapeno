package database

import "fmt"
//import "github.com/sbezverk/gobmp/pkg/srv6"

const L3VPNRTName = "L3VPNRT"

type L3VPNRT struct {
	ID       string            `json:"_id,omitempty"`
	Key      string            `json:"_key,omitempty"`
	RT       string            `json:"RT,omitempty"`
	Prefixes map[string]string `json:"Prefixes,omitempty"`
}

func (p L3VPNRT) GetKey() (string, error) {
	if p.Key == "" {
		return p.makeKey()
	}
	return p.Key, nil
}

func (p *L3VPNRT) SetKey() error {
	k, err := p.makeKey()
	if err != nil {
		return err
	}
	p.Key = k
	return nil
}

func (p *L3VPNRT) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if p.RT != "" {
		ret = fmt.Sprintf("%s",p.RT)
		err = nil
	}
	return ret, err
}

func (p L3VPNRT) GetType() string {
	return L3VPNRTName
}
