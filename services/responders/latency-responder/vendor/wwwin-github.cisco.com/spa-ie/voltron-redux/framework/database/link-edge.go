package database

import "fmt"

const linkName = "LinkEdges"

type LinkEdge struct {
	From    string `json:"_from,omitempty"`
	To      string `json:"_to,omitempty"`
	Key     string `json:"_key,omitempty"`
	FromIP  string `json:"FromIP,omitempty"`
	ToIP    string `json:"ToIP,omitempty"`
	Netmask string `json:"Netmask,omitempty"`
	Label   string `json:"Label,omitempty"`
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
	return linkName
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
