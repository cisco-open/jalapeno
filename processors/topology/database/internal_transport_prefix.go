package database

import "fmt"

const InternalTransportPrefixName = "InternalTransportPrefixes"

type InternalTransportPrefix struct {
	Key            string `json:"_key,omitempty"`
	Name           string `json:"Name,omitempty"`
	RouterIP       string `json:"RouterIP,omitempty"`
	BGPID          string `json:"BGPID,omitempty"`
	ASN            string `json:"ASN,omitempty"`
        SRGB           string `json:"SRGB,omitempty"`
        PrefixSIDIndex string `json:"PrefixSIDIndex,omitempty"`
        SRPrefixSID    string `json:"SRPrefixSID,omitempty"`
}

func (r InternalTransportPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *InternalTransportPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *InternalTransportPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterIP != "" {
		ret = fmt.Sprintf("InternalTransportPrefix:%s", r.RouterIP)
		err = nil
	}
	return ret, err
}

func (r InternalTransportPrefix) GetType() string {
	return InternalTransportPrefixName
}
