package database

import "fmt"

const ExternalRouterInterfaceName = "ExternalRouterInterfaces"

type ExternalRouterInterface struct {
	Key               string `json:"_key,omitempty"`
	BGPID             string `json:"BGPID,omitempty"`
	RouterASN         string `json:"ASN,omitempty"`
        RouterIP          string `json:"RouterIP,omitempty"`
	RouterInterfaceIP string `json:"RouterInterfaceIP,omitempty"`
}

func (r ExternalRouterInterface) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *ExternalRouterInterface) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *ExternalRouterInterface) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.RouterIP != "" {
		ret = fmt.Sprintf("%s_%s", r.RouterIP, r.RouterInterfaceIP)
		err = nil
	}
	return ret, err
}

func (r ExternalRouterInterface) GetType() string {
	return ExternalRouterInterfaceName
}
