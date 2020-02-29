package database

import (
	"fmt"
)

const EPETopologyName = "EPETopology"

type EPETopology struct {
	EPENodeKey    string `json:"_from,omitempty"`
        ExtPrefixKey  string `json:"_to,omitempty"`
        Key	      string `json:"_key,omitempty"`
	RouterID      string `json:"RouterID,omitempty"`
	ASN           string `json:"ASN,omitempty"`
        PeerIP        string `json:"PeerIP,omitempty"`
        PeerASN       string `json:"PeerASN,omitempty"`
	LocalInterfaceIP  string `json:"LocalInterfaceIP,omitempty"`
	RemoteInterfaceIP  string `json:"RemoteInterfaceIP,omitempty"`
	NextHop       string `json:"NextHop,omitempty"`
        PrefixASN       string `json:"PrefixASN,omitempty"`
        ASPath        string `json:"ASPath,omitempty"`
}

func (l EPETopology) GetKey() (string, error) {
	if l.Key == "" {
		return l.makeKey()
	}
	return l.Key, nil
}

func (l *EPETopology) SetKey() error {
	k, err := l.makeKey()
	if err != nil {
		return err
	}
	l.Key = k
	return nil
}

func (l *EPETopology) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if l.LocalInterfaceIP != "" && l.RemoteInterfaceIP != "" {
		ret = fmt.Sprintf("%s_%s_%s_%s", l.RouterID, l.LocalInterfaceIP, l.RemoteInterfaceIP, l.PeerIP)
		err = nil
	}
	return ret, err
}

func (l EPETopology) GetType() string {
	return EPETopologyName
}

func (l *EPETopology) SetEdge(to DBObject, from DBObject) error {
	var err error
	l.PeerIP, err = GetID(to)
	if err != nil {
		return err
	}
	l.RouterID, err = GetID(from)
	if err != nil {
		return err
	}
	return nil
}

