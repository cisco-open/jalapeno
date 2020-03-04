package database

import (
	"fmt"
)

const EPELinkName = "EPELink"

type EPELink struct {
	LocalRouterKey   string `json:"_from,omitempty"`
        RemoteRouterKey  string `json:"_to,omitempty"`
        Key	           string `json:"_key,omitempty"`
	LocalRouterID      string `json:"LocalRouterID,omitempty"`
        RemoteRouterID     string `json:"RemoteRouterID,omitempty"`
	LocalInterfaceIP   string `json:"FromInterfaceIP,omitempty"`
        RemoteInterfaceIP  string `json:"ToInterfaceIP,omitempty"`
	LocalASN                string `json:"LocalASN,omitempty"`
	RemoteASN          string `json:"RemoteASN,omitempty"`
	Protocol        string `json:"Protocol,omitempty"`
        Label           string `json:"Label,omitempty"`

}

func (l EPELink) GetKey() (string, error) {
	if l.Key == "" {
		return l.makeKey()
	}
	return l.Key, nil
}

func (l *EPELink) SetKey() error {
	k, err := l.makeKey()
	if err != nil {
		return err
	}
	l.Key = k
	return nil
}

func (l *EPELink) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if l.LocalInterfaceIP != "" && l.RemoteInterfaceIP != "" {
		ret = fmt.Sprintf("%s_%s_%s_%s", l.LocalRouterID, l.LocalInterfaceIP, l.RemoteInterfaceIP, l.RemoteRouterID)
		err = nil
	}
	return ret, err
}

func (l EPELink) GetType() string {
	return EPELinkName
}

func (l *EPELink) SetEdge(to DBObject, from DBObject) error {
	var err error
	l.RemoteRouterID, err = GetID(to)
	if err != nil {
		return err
	}
	l.LocalRouterID, err = GetID(from)
	if err != nil {
		return err
	}
	return nil
}

