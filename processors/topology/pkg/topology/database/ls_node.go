package database

import (
	"fmt"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

const LSNodeName = "LSNode"

type LSNode struct {
	Key                  string               `json:"_key,omitempty"`
	Name                 string               `json:"Name,omitempty"`
	RouterID             string               `json:"RouterID,omitempty"`
	ASN                  int32                `json:"ASN,omitempty"`
	SRGBStart            int                  `json:"SRGBStart,omitempty"`
	SRGBRange            uint32               `json:"SRGBRange,omitempty"`
	SRCapabilityFlags    uint8                `json:"SRCapabilityFlags,omitempty"`
	IGPID                string               `json:"IGPID,omitempty"`
	SRv6CapabilitiesTLV  *srv6.CapabilityTLV  `json:"SRv6CapabilitiesTLV,omitempty"`
	SRAlgorithm          []int                `json:"SRAlgorithm,omitempty"`
	SRLocalBlock         *sr.LocalBlock       `json:"SRLocalBlock,omitempty"`
	NodeMSD              []*base.MSDTV        `json:"NodeMSD,omitempty"`
	AreaID               string               `json:"AreaID,omitempty"`
	Protocol             string               `json:"Protocol,omitempty"`
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
	if r.IGPID != "" {
		ret = fmt.Sprintf("%s", r.IGPID)
		err = nil
	}
	return ret, err
}

func (r LSNode) GetType() string {
	return LSNodeName
}
