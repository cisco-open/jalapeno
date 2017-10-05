package arango

import "fmt"

const prefixName = "Prefixes"

type Prefix struct {
	Key    string `json:"_key,omitempty"`
	Prefix string `json:"Prefix,omitempty"`
	Length int    `json:"Length,omitempty"`
}

func (p Prefix) GetKey() string {
	return fmt.Sprintf("%s_%d", p.Prefix, p.Length)
}

func (p *Prefix) SetKey() error {
	ret := ErrKeyInvalid
	if p.Prefix != "" && p.Length != 0 {
		p.Key = p.GetKey()
		ret = nil
	}
	return ret
}

func (p Prefix) GetID() string {
	return fmt.Sprintf("%s/%s", p.GetType(), p.GetKey())
}

func (p Prefix) GetType() string {
	return prefixName
}
