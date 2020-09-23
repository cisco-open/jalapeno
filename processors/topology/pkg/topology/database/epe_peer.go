package database

import "fmt"

const EPEPeerName = "EPEPeer"

type EPEPeer struct {
	Key           string   `json:"_key,omitempty"`
	PeerIP        string   `json:"peer_ip,omitempty"`
	RouterIP      string   `json:"router_ip,omitempty"`
	PeerASN       int32    `json:"peer_asn,omitempty"`
	Nexthop       string   `json:"nexthop,omitempty"`
	IsNexthopIPv4 bool     `json:"is_nexthop_ipv4,omitempty"`
	Labels        []uint32 `json:"labels,omitempty"`
	Timestamp     string   `json:"timestamp,omitempty"`
}

func (r EPEPeer) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *EPEPeer) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *EPEPeer) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.PeerIP != "" {
		ret = fmt.Sprintf("%s", r.PeerIP)
		err = nil
	}
	return ret, err
}

func (r EPEPeer) GetType() string {
	return EPEPeerName
}
