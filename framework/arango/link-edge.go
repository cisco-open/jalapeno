package arango

import (
	"fmt"
	"strings"
)

const linkName = "LinkEdges"

type LinkEdge struct {
	From        string  `json:"_from,omitempty"`
	To          string  `json:"_to,omitempty"`
	Key         string  `json:"_key,omitempty"`
	FromIP      string  `json:"FromIP"`
	ToIP        string  `json:"ToIP"`
	Netmask     string  `json:"Netmask"`
	Label       string  `json:"Label"`
	Latency     int     `json:"Latency"`
	Utilization float32 `json:"Utilization"`
}

func (l LinkEdge) GetKey() string {
	return l.Key
}

func (l *LinkEdge) SetKey() error {
	ret := ErrKeyInvalid
	if l.From != "" && l.To != "" && l.FromIP != "" && l.ToIP != "" {
		l.Key = fmt.Sprintf("%s_%s_%s_%s", strings.Replace(l.From, "/", "", -1), strings.Replace(l.To, "/", "", -1), l.FromIP, l.ToIP) // tmp
		ret = nil
	}
	return ret
}

func (l LinkEdge) GetType() string {
	return linkName
}

func (l *LinkEdge) SetEdge(to DBObject, from DBObject) {
	l.To = to.GetKey()
	l.From = from.GetKey()
}
