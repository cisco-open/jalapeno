package database

import "fmt"

const InternalRouterName = "InternalRouters"

type InternalRouter struct {
	Key          string `json:"_key,omitempty"`
	Name         string `json:"Name,omitempty"`
	BGPID        string `json:"BGPID,omitempty"`
	ASN          string `json:"ASN,omitempty"`
        RouterIP     string `json:"RouterIP,omitempty"`
        SRGB         string `json:"SRGB,omitempty"`
        NodeSIDIndex string `json:"NodeSIDIndex,omitempty"`
        SRNodeSID    string `json:"SRNodeSID,omitempty"`
	IGPID        string `json:"IGPID,omitempty"`
}

func (r InternalRouter) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *InternalRouter) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *InternalRouter) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterIP != "" {
		ret = fmt.Sprintf("%s", r.RouterIP)
		err = nil
	}
	return ret, err
}

func (r InternalRouter) GetType() string {
	return InternalRouterName
}
