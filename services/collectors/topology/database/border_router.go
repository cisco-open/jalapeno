package database

import "fmt"

const BorderRouterName = "BorderRouters"

type BorderRouter struct {
	Key          string `json:"_key,omitempty"`
	BGPID        string `json:"BGPID,omitempty"`
	ASN          string `json:"ASN,omitempty"`
        RouterIP     string `json:"RouterIP,omitempty"`
}

func (r BorderRouter) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *BorderRouter) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *BorderRouter) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterIP != "" {
		ret = fmt.Sprintf("%s", r.RouterIP)
		err = nil
	}
	return ret, err
}

func (r BorderRouter) GetType() string {
	return BorderRouterName
}
