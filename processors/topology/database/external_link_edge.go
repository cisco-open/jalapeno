package database

import (
	"fmt"
)

const ExternalLinkEdgeName = "ExternalLinkEdges"

type ExternalLinkEdge struct {
        SrcIP           string `json:"_from,omitempty"`
        DstIP           string `json:"_to,omitempty"`
        Source          string `json:"Source,omitempty"`
        Destination     string `json:"Destination,omitempty"`
        Key	        string `json:"_key,omitempty"`
        SrcInterfaceIP  string `json:"FromInterfaceIP,omitempty"`
        DstInterfaceIP  string `json:"ToInterfaceIP,omitempty"`
	Protocol        string `json:"Protocol,omitempty"`
        Label           string `json:"Label,omitempty"`
}

func (l ExternalLinkEdge) GetKey() (string, error) {
	if l.Key == "" {
		return l.makeKey()
	}
	return l.Key, nil
}

func (l *ExternalLinkEdge) SetKey() error {
	k, err := l.makeKey()
	if err != nil {
		return err
	}
	l.Key = k
	return nil
}

func (l *ExternalLinkEdge) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if l.SrcInterfaceIP != "" && l.DstInterfaceIP != "" {
		ret = fmt.Sprintf("%s_%s_%s_%s", l.SrcIP, l.SrcInterfaceIP, l.DstInterfaceIP, l.DstIP)
		err = nil
	}
	return ret, err
}

func (l ExternalLinkEdge) GetType() string {
	return ExternalLinkEdgeName
}

func (l *ExternalLinkEdge) SetEdge(to DBObject, from DBObject) error {
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

