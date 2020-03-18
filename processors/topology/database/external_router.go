package database

import "fmt"

const ExternalRouterName = "ExternalRouters"

type ExternalRouter struct {
	Key       string `json:"_key,omitempty"`
	Name      string `json:"Name,omitempty"`
	BGPID     string `json:"BGPID,omitempty"`
	ASN       string `json:"ASN,omitempty"`
        RouterIP  string `json:"RouterIP,omitempty"`
        PeeringType  string `json:"PeerType,omitempty"`
}

func (r ExternalRouter) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *ExternalRouter) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *ExternalRouter) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterIP != "" {
		ret = fmt.Sprintf("%s", r.RouterIP)
		err = nil
	}
	return ret, err
}

func (r ExternalRouter) GetType() string {
	return ExternalRouterName
}
