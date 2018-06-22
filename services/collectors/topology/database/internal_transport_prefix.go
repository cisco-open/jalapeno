package database

import "fmt"

const InternalTransportPrefixName = "InternalTransportPrefix"

type InternalTransportPrefix struct {
	Key      string `json:"_key,omitempty"`
	Name     string `json:"Name,omitempty"`
	RouterIP string `json:"RouterIP,omitempty"`
	BGPID    string `json:"BGPID,omitempty"`
	IsLocal  bool   `json:"IsLocal"`
	ASN      string `json:"ASN,omitempty"`
        SRGB     string `json:"SRGB,omitempty"`
        SRNodeSID string `json:"SRNodeSID,omitempty"`
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
	if r.BGPID != "" {
		ret = fmt.Sprintf("%s", r.BGPID)
		err = nil
	}
	return ret, err
}

func (r InternalTransportPrefix) GetType() string {
	return InternalTransportPrefixName
}
