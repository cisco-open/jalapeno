package arango

import "fmt"

const prefixName = "Prefixes"

type Prefix struct {
	Key    string `json:"_key,omitempty"`
	Prefix string `json:"Prefix,omitempty"`
	Length int    `json:"Length,omitempty"`
}

func (p Prefix) GetKey() string {
	return p.Key
}

func (p *Prefix) SetKey() error {
	ret := ErrKeyInvalid
	if p.Prefix != "" && p.Length != 0 {
		p.Key = fmt.Sprintf("%s_%d", p.Prefix, p.Length)
		ret = nil
	}
	return ret
}

func (p Prefix) GetType() string {
	return prefixName
}
