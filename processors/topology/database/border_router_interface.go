package database

import "fmt"

const BorderRouterInterfaceName = "BorderRouterInterfaces"

type BorderRouterInterface struct {
	Key               string `json:"_key,omitempty"`
	BGPID             string `json:"BGPID,omitempty"`
	RouterASN         string `json:"ASN,omitempty"`
        RouterIP          string `json:"RouterIP,omitempty"`
	RouterInterfaceIP string `json:"RouterInterfaceIP,omitempty"`
        EPELabel          string `json:"EPELabel,omitempty"`
}

func (r BorderRouterInterface) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *BorderRouterInterface) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *BorderRouterInterface) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterIP != "" {
		ret = fmt.Sprintf("%s_%s", r.RouterIP, r.RouterInterfaceIP)
		err = nil
	}
	return ret, err
}

func (r BorderRouterInterface) GetType() string {
	return BorderRouterInterfaceName
}
