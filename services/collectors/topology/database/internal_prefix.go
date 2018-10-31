package database

import "fmt"

const InternalPrefixName = "InternalPrefixes"

type InternalPrefix struct {
	Key         string `json:"_key,omitempty"`
        Prefix      string `json:"Prefix,omitempty"`
        Length      int    `json:"Length,omitempty"`
	Name        string `json:"Name,omitempty"`
	ASN         string `json:"ASN,omitempty"`
        ASPathCount string `json:"ASPathCount,omitempty"`
        SRLabel     string `json:"SRLabel,omitempty"`
}

func (r InternalPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *InternalPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *InternalPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.Prefix != "" {
                ret = fmt.Sprintf("%s", r.Prefix)
		err = nil
	}
	return ret, err
}

func (r InternalPrefix) GetType() string {
	return InternalPrefixName
}
