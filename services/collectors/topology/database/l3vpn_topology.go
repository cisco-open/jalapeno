package database

import (
	"fmt"
)

const L3VPN_TopologyName = "L3VPN_Topology"

type L3VPN_Topology struct {
        Key             string `json:"_key,omitempty"`
        SrcIP           string `json:"_from,omitempty"`
        DstIP           string `json:"_to,omitempty"`
        Source          string `json:"Source,omitempty"`
        Destination     string `json:"Destination,omitempty"`
	RD              string `json:"RD,omitempty"`
        Label           int    `json:"Label,omitempty"`
        PE_Nexthop      string `json:"PE_Nexthop,omitempty"`
}

func (l L3VPN_Topology) GetKey() (string, error) {
	if l.Key == "" {
		return l.makeKey()
	}
	return l.Key, nil
}

func (l *L3VPN_Topology) SetKey() error {
	k, err := l.makeKey()
	if err != nil {
		return err
	}
	l.Key = k
	return nil
}

func (l *L3VPN_Topology) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	ret = fmt.Sprintf("%s_%s_%s", l.Source, l.RD, l.Destination)
	err = nil
	return ret, err
}

func (l L3VPN_Topology) GetType() string {
	return L3VPN_TopologyName
}

func (l *L3VPN_Topology) SetEdge(to DBObject, from DBObject) error {
	var err error
	l.DstIP, err = GetID(to)
	if err != nil {
		return err
	}
	l.SrcIP, err = GetID(from)
	if err != nil {
		return err
	}
	return nil
}

