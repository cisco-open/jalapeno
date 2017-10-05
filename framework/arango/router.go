package arango

import "fmt"

const routerName = "Routers"

type Router struct {
	Key      string `json:"_key,omitempty"`
	Name     string `json:"_name,omitempty"`
	RouterIP string `json:"RouterIP,omitempty"`
	BGPID    string `json:"BGPID,omitempty"`
	IsLocal  bool   `json:"IsLocal,omitempty"`
	ASN      string `json:"ASN,omitempty"`
}

func (r Router) GetKey() string {
	return fmt.Sprintf("%s_%s", r.BGPID, r.ASN)
}

func (r Router) GetID() string {
	return fmt.Sprintf("%s/%s", r.GetType(), r.GetKey())
}

func (r *Router) SetKey() error {
	ret := ErrKeyInvalid
	if r.BGPID != "" && r.ASN != "" {
		r.Key = r.GetKey()
		ret = nil
	}
	return ret
}

func (r Router) GetType() string {
	return routerName
}
