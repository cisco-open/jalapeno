package arango

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
	Latency int    `json:"Latency,omitempty"`
	// ASPath ???
	Utilization float32 `json:"Utilization,omitempty"`
}

func (l LinkEdge) GetKey() string {
	return fmt.Sprintf("%s_%s", l.FromIP, l.ToIP) // tmp
}

func (l *LinkEdge) SetKey() error {
	ret := ErrKeyInvalid
	if l.From != "" && l.To != "" && l.FromIP != "" && l.ToIP != "" {
		l.Key = l.GetKey()
		ret = nil
	}
	return ret
}

func (l LinkEdge) GetType() string {
	return linkName
}

func (l LinkEdge) GetID() string {
	return fmt.Sprintf("%s/%s", l.GetType(), l.GetKey())
}

func (l *LinkEdge) SetEdge(to DBObject, from DBObject) {
	l.To = to.GetKey()
	l.From = from.GetKey()
}
