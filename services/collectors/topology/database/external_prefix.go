package database

import "fmt"

const ExternalPrefixName = "ExternalPrefixes"

type ExternalPrefix struct {
	Key         string `json:"_key,omitempty"`
        Prefix      string `json:"Prefix,omitempty"`
        Length      int    `json:"Length,omitempty"`
	Name        string `json:"Name,omitempty"`
	ASN         string `json:"ASN,omitempty"`
        ASPathCount string `json:"ASPathCount,omitempty"`
}

func (r ExternalPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *ExternalPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *ExternalPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.Prefix != "" {
		ret = fmt.Sprintf("%s", r.Prefix)
		err = nil
	}
	return ret, err
}

func (r ExternalPrefix) GetType() string {
	return ExternalPrefixName
}
