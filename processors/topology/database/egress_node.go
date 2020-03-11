package database

import "fmt"

const EgressNodeName = "EgressNodes"

type EgressNode struct {
	Key         string `json:"_key,omitempty"`
	Name        string `json:"Name,omitempty"`
	BGPID       string `json:"BGPID,omitempty"`
        SRNodeSID   string `json:"SRNodeSID,omitempty"`
	IntfIP      string `json:"InterfaceIP,omitempty"`
        EPELabel    string `json:"EPELabel,omitempty"`
	NeighborIP  string `json:"NeighborIP,omitempty"`
	NeighborASN string `json:"NeighborASN,omitempty"`
}

func (r EgressNode) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *EgressNode) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *EgressNode) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.BGPID != "" {
		ret = fmt.Sprintf("%s", r.BGPID)
		err = nil
	}
	return ret, err
}

func (r EgressNode) GetType() string {
	return EgressNodeName
}
