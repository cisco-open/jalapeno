package arango

import (
	"fmt"
	"strings"
)

const asName = "PrefixEdges"

type PrefixEdge struct {
	From        string   `json:"_from,omitempty"`
	To          string   `json:"_to,omitempty"`
	Key         string   `json:"_key,omitempty"`
	NextHop     string   `json:"NextHop,omitempty"`
	InterfaceIP string   `json:"InterfaceIP,omitempty"`
	ASPath      []string `json:"ASPath,omitempty"`
	Label       string   `json:"Label,omitempty"`
	BGPPolicy   string   `json:"BGPPolicy,omitempty"`
	Latency     int      `json:"Latency,omitempty"`
}

func (a PrefixEdge) GetKey() string {
	return a.Key
}

func (a *PrefixEdge) SetKey() error {
	ret := ErrKeyInvalid
	if a.From != "" && a.To != "" {
		a.Key = fmt.Sprintf("%s_%s", strings.Replace(a.From, "/", "_", -1), strings.Replace(a.To, "/", "_", -1)) // tmp
		ret = nil
	}
	return ret
}

func (a PrefixEdge) GetType() string {
	return asName
}

func (a *PrefixEdge) SetEdge(to DBObject, from DBObject) {
	a.To = fmt.Sprintf("%s/%s", to.GetType(), to.GetKey())
	a.From = fmt.Sprintf("%s/%s", from.GetType(), from.GetKey())
}
