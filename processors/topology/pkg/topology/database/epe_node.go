package database

import "fmt"

const EPENodeName = "Demo-EPENode"

type EPENode struct {
	Key      string   `json:"_key,omitempty"`
	Name     string   `json:"name,omitempty"`
	RouterID string   `json:"router_id,omitempty"`
	PeerIP   []string `json:"peer_ip,omitempty"`
	ASN      int32    `json:"asn,omitempty"`
}

func (r EPENode) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *EPENode) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *EPENode) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterID != "" {
		ret = fmt.Sprintf("%s", r.RouterID)
		err = nil
	}
	return ret, err
}

func (r EPENode) GetType() string {
	return EPENodeName
}
