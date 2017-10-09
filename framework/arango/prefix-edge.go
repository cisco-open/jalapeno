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
	Labels      []string `json:"Labels,omitempty"`
	BGPPolicy   string   `json:"BGPPolicy,omitempty"`
	Latency     int      `json:"Latency,omitempty"`
	Utilization float32  `json:"Utilization"`
}

func (a PrefixEdge) GetKey() (string, error) {
	if a.Key == "" {
		return a.makeKey()
	}
	return a.Key, nil
}

func (a *PrefixEdge) SetKey() error {
	k, err := a.makeKey()
	if err != nil {
		return err
	}
	a.Key = k
	return nil
}

func (a *PrefixEdge) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if a.From != "" && a.To != "" {
		ret = fmt.Sprintf("%s_%s", strings.Replace(a.From, "/", "_", -1), strings.Replace(a.To, "/", "_", -1))
		err = nil
	}
	return ret, err
}

func (a PrefixEdge) GetType() string {
	return asName
}

func (a *PrefixEdge) SetEdge(to DBObject, from DBObject) error {
	var err error
	a.To, err = GetID(to)
	if err != nil {
		return err
	}
	a.From, err = GetID(from)
	if err != nil {
		return err
	}
	return nil
}
