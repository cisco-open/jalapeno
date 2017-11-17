package database

import (
	"fmt"
	"strings"
)

const linkEdgeNamev4 = "LinkEdgesV4"
const linkEdgeNamev6 = "LinkEdgesV6"

type LinkEdge struct {
	From    string `json:"_from,omitempty"`
	To      string `json:"_to,omitempty"`
	Key     string `json:"_key,omitempty"`
	FromIP  string `json:"FromIP,omitempty"`
	ToIP    string `json:"ToIP,omitempty"`
	Netmask string `json:"Netmask,omitempty"`
	Label   string `json:"Label,omitempty"`
	V6      bool   `json:"-"`
}

func (l LinkEdge) GetKey() (string, error) {
	if l.Key == "" {
		return l.makeKey()
	}
	return l.Key, nil
}

func (l *LinkEdge) SetKey() error {
	k, err := l.makeKey()
	if err != nil {
		return err
	}
	l.Key = k
	return nil
}

func (l *LinkEdge) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if l.FromIP != "" && l.ToIP != "" {
		//ret = fmt.Sprintf("%s_%s_%s_%s", strings.TrimPrefix(l.From, "Routers/"), l.FromIP, l.ToIP, strings.TrimPrefix(l.To, "Routers/")) // tmp
		ret = fmt.Sprintf("%s_%s", l.FromIP, l.ToIP)
		err = nil
	}
	return ret, err
}

func (l LinkEdge) GetType() string {
	if l.V6 || strings.Contains(l.FromIP, ":") || strings.Contains(l.ToIP, ":") {
		l.V6 = true
		return linkEdgeNamev6
	}
	return linkEdgeNamev4
}

func (l *LinkEdge) SetEdge(to DBObject, from DBObject) error {
	var err error
	l.To, err = GetID(to)
	if err != nil {
		return err
	}
	l.From, err = GetID(from)
	if err != nil {
		return err
	}
	return nil
}

func (l *LinkEdge) IsV6() bool {
	return l.V6 || strings.Contains(l.FromIP, ":") || strings.Contains(l.ToIP, ":")
}
