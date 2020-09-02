package database

import (
	"fmt"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/bgpls"
)

const LSPrefixName = "LSPrefix"

type LSPrefix struct {
	Key             string              `json:"_key,omitempty"`
	IGPRouterID     string              `json:"IGPRouterID,omitempty"`
	Prefix          string              `json:"Prefix,omitempty"`
	Length          int32               `json:"Length,omitempty"`
	Protocol        string              `json:"Protocol,omitempty"`
	Timestamp       string              `json:"Timestamp,omitempty"`
	LSPrefixSID     []*sr.PrefixSIDTLV  `json:"PrefixSID,omitempty"`
	PrefixAttrFlags uint8 `json:"PrefixAttrFlags,omitempty"`
        FlexAlgoPrefixMetric *bgpls.FlexAlgoPrefixMetric `json:"FlexAlgoPrefixMetric,omitempty"`
}

func (r LSPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *LSPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *LSPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.IGPRouterID != "" {
		ret = fmt.Sprintf("%s_%s", r.IGPRouterID, r.Prefix)
		err = nil
	}
	return ret, err
}

func (r LSPrefix) GetType() string {
	return LSPrefixName
}
