package database

import "fmt"

const EPEExternalPrefixName = "EPEExternalPrefix"

type EPEExternalPrefix struct {
	Key         string `json:"_key,omitempty"`
        Prefix      string `json:"Prefix,omitempty"`
        Length      int    `json:"Length,omitempty"`
        RouterID    string `json:"RouterID,omitempty"`
        PeerIP      string `json:"PeerIP,omitempty"`
        PeerASN     string `json:"PeerASN,omitempty"`
	OriginAS    string `json:"OriginAS,omitempty"`
	ASPath      string `json:"ASPath,omitempty"`
}

func (r EPEExternalPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *EPEExternalPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *EPEExternalPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.Prefix != "" {
		ret = fmt.Sprintf("%s_%s", r.PeerIP, r.Prefix)
		err = nil
	}
	return ret, err
}

func (r EPEExternalPrefix) GetType() string {
	return EPEExternalPrefixName
}
