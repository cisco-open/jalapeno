package database

import "fmt"

const UnicastPrefixName = "UnicastPrefix"

type UnicastPrefix struct {
	Key           string   `json:"_key,omitempty"`
	Prefix        string   `json:"prefix,omitempty"`
	Length        int32    `json:"length,omitempty"`
	PeerIP        string   `json:"peer_ip,omitempty"`
	RouterIP      string   `json:"router_ip,omitempty"`
	PeerASN       int32    `json:"peer_asn,omitempty"`
	Nexthop       string   `json:"nexthop,omitempty"`
	OriginAS      int32    `json:"origin_as,omitempty"`
	ASPath        []uint32 `json:"as_path,omitempty"`
	ASPathCount   int32    `json:"as_path_count,omitempty"`
	MED           uint32   `json:"med"`
	LocalPref     uint32   `json:"local_pref"`
	CommunityList string   `json:"community_list,omitempty"`
	ExtComm       []string `json:"ext_comm,omitempty"`
	IsIPv4        bool     `json:"is_ipv4,omitempty"`
	IsNexthopIPv4 bool     `json:"is_nexthop_ipv4,omitempty"`
	PathID        int32    `json:"path_id,omitempty"`
	Labels        []uint32 `json:"labels,omitempty"`
	Timestamp     string   `json:"timestamp,omitempty"`
}

func (r UnicastPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *UnicastPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *UnicastPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.Prefix != "" {
		ret = fmt.Sprintf("%s_%s_%d", r.PeerIP, r.Prefix, r.Length)
		err = nil
	}
	return ret, err
}

func (r UnicastPrefix) GetType() string {
	return UnicastPrefixName
}