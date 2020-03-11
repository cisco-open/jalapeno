package database

import (
	"fmt"
	"strings"
)

const EPEEdgeName = "EPEEdges"

type EPEEdge struct {
	From        string   `json:"_from,omitempty"`
	To          string   `json:"_to,omitempty"`
	Key         string   `json:"_key,omitempty"`
	ASPath      []string `json:"ASPath,omitempty"`
        EgressIntIP string   `json:"InterfaceIP,omitempty"`
        SRPrefixSID string   `json:"SRPrefixSID,omitempty"`
	EPELabel    string `json:"EPELabel,omitempty"`
}

func (a EPEEdge) GetKey() (string, error) {
	if a.Key == "" {
		return a.makeKey()
	}
	return a.Key, nil
}

func (a *EPEEdge) SetKey() error {
	k, err := a.makeKey()
	if err != nil {
		return err
	}
	a.Key = k
	return nil
}

func (a *EPEEdge) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if a.From != "" && a.To != "" {
		ret = fmt.Sprintf("%s_%s_%s", strings.Replace(a.From, "/", "_", -1), a.EgressIntIP, strings.Replace(a.To, "/", "_", -1))
		err = nil
	}
	return ret, err
}

func (a EPEEdge) GetType() string {
	return EPEEdgeName
}

func (a *EPEEdge) SetEdge(to DBObject, from DBObject) error {
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
