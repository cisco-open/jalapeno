package database

import "fmt"

const PeeringRouterName = "PeeringRouters"

type PeeringRouter struct {
	Key          string `json:"_key,omitempty"`
	BGPID        string `json:"BGPID,omitempty"`
	ASN          string `json:"ASN,omitempty"`
        RouterIP     string `json:"RouterIP,omitempty"`
}

func (r PeeringRouter) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *PeeringRouter) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *PeeringRouter) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterIP != "" {
		ret = fmt.Sprintf("%s", r.RouterIP)
		err = nil
	}
	return ret, err
}

func (r PeeringRouter) GetType() string {
	return PeeringRouterName
}
