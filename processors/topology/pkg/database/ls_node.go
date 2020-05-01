package database

import (
        "fmt"
        "github.com/sbezverk/gobmp/pkg/sr"
)

const LSNodeName = "LSNode"

type LSNode struct {
        Key          string `json:"_key,omitempty"`
        Name         string `json:"Name,omitempty"`
        RouterID     string `json:"RouterID,omitempty"`
        //BGPID      string `json:"BGPID,omitempty"`
        ASN          int32 `json:"ASN,omitempty"`
        SRGBStart    int `json:"SRStart,omitempty"`
        SRGBRange    uint32 `json:"SRRange,omitempty"`
        SRGB         *sr.Capability `json:"SRGB,omitempty"`
        SIDIndex     int `json:"SIDIndex,omitempty"`
        PrefixSID    int `json:"PrefixSID,omitempty"`
        IGPID        string `json:"IGPID,omitempty"`
        SRv6Capabilities string `json:"SRv6Capabilities,omitempty"`
        SRAlgorithm  []int `json:"SRAlgorithm,omitempty"`
        SRLocalBlock string `json:"SRLocalBlock,omitempty"`
        NodeMaxSIDDepth string `json:"NodeMaxSIDDepth,omitempty"`
        AreaID       string `json:"AreaID,omitempty"`
        Protocol     string `json:"Protocol,omitempty"`
}

func (r LSNode) GetKey() (string, error) {
        if r.Key == "" {
                return r.makeKey()
        }
        return r.Key, nil
}

func (r *LSNode) SetKey() error {
        k, err := r.makeKey()
        if err != nil {
                return err
        }
        r.Key = k
        return nil
}

func (r *LSNode) makeKey() (string, error) {
        err := ErrKeyInvalid
        ret := ""
        if r.RouterID != "" {
                ret = fmt.Sprintf("%s", r.RouterID)
                //              ret = fmt.Sprintf("Router/%s", r.RouterIP)
                err = nil
        }
        return ret, err
}

func (r LSNode) GetType() string {
        return LSNodeName
}

