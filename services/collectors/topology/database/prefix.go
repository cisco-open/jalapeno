package database

import "fmt"

const PrefixName = "Prefixes"

type Prefix struct {
	Key    string `json:"_key,omitempty"`
	Prefix string `json:"Prefix,omitempty"`
	Length int    `json:"Length,omitempty"`
        ASN    string `json:"ASN,omitempty"`
}

func (p Prefix) GetKey() (string, error) {
	if p.Key == "" {
		return p.makeKey()
	}
	return p.Key, nil
}

func (p *Prefix) SetKey() error {
	k, err := p.makeKey()
	if err != nil {
		return err
	}
	p.Key = k
	return nil
}

func (p *Prefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if p.Prefix != "" && p.Length != 0 {
		ret = fmt.Sprintf("%s", p.Prefix)
		err = nil
	}
	return ret, err
}

func (p Prefix) GetType() string {
	return PrefixName
}
