package database

import "fmt"

const EPEPrefixName = "EPEPrefix"

type EPEPrefix struct {
        Key           string   `json:"_key,omitempty"`
        Prefix        string   `json:"Prefix,omitempty"`
        Length        int32    `json:"Length,omitempty"`
        PeerIP        string   `json:"PeerIP,omitempty"`
        PeerASN       int32    `json:"PeerASN,omitempty"`
        Nexthop       string   `json:"Nexthop,omitempty"`
        OriginASN     string   `json:"OriginASN,omitempty"`
        ASPath        []uint32 `json:"ASPath,omitempty"`
        ASPathCount   int32    `json:"ASPathCount,omitempty"`
        MED           uint32   `json:"MED"`
        LocalPref     uint32   `json:"LocalPref"`
        CommunityList string   `json:"CommunityList,omitempty"`
        ExtComm       string   `json:"ExtComm,omitempty"`
        IsIPv4        bool     `json:"IsIPv4,omitempty"`
        IsNexthopIPv4 bool     `json:"IsNexthopIPv4,omitempty"`
        Labels        []uint32 `json:"Labels,omitempty"`
        Timestamp     string   `json:"Timestamp,omitempty"`
}

func (r EPEPrefix) GetKey() (string, error) {
        if r.Key == "" {
                return r.makeKey()
        }
        return r.Key, nil
}

func (r *EPEPrefix) SetKey() error {
        k, err := r.makeKey()
        if err != nil {
                return err
        }
        r.Key = k
        return nil
}

func (r *EPEPrefix) makeKey() (string, error) {
        err := ErrKeyInvalid
        ret := ""
        if r.Prefix != "" {
                ret = fmt.Sprintf("%s_%s_%d", r.PeerIP, r.Prefix, r.Length)
                err = nil
        }
        return ret, err
}

func (r EPEPrefix) GetType() string {
        return EPEPrefixName
}

