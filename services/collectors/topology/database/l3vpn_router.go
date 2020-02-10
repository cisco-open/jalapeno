package database

import "fmt"

const L3VPN_RouterName = "L3VPN_Routers"

type L3VPN_Router struct {
        Key              string `json:"_key,omitempty"`
        RD               []string `json:"RD,omitempty"`
        RouterIP         string `json:"RouterIP,omitempty"`
        ASN              string `json:"ASN,omitempty"`
        AdvertisingPeer  string `json:"AdvertisingPeer,omitempty"`
        Prefix_SID       string `json:"Prefix_SID,omitempty"`
        ExtComm          string `json:"ExtComm,omitempty"`
}

func (r L3VPN_Router) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *L3VPN_Router) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *L3VPN_Router) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if (r.RouterIP != "") {
		ret = fmt.Sprintf("%s", r.RouterIP)
		err = nil
	}
	return ret, err
}

func (r L3VPN_Router) GetType() string {
	return L3VPN_RouterName
}
