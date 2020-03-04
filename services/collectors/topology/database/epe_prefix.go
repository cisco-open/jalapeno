package database

import "fmt"

const EPEPrefixName = "EPEPrefix"

type EPEPrefix struct {
	Key         string `json:"_key,omitempty"`
        Prefix      string `json:"Prefix,omitempty"`
        Length      int    `json:"Length,omitempty"`
	Name        string `json:"Name,omitempty"`
        RouterID    string `json:"RouterID,omitempty"`
        PeerIP      string `json:"PeerIP,omitempty"`
        PeerASN   string `json:"PeerASN,omitempty"`
	RemoteASN   string `json:"RemoteASN,omitempty"`
	ASPath      string `json:"ASPath,omitempty"`
}

func (r EPEPrefix) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *EPEPrefix) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *EPEPrefix) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.Prefix != "" {
		ret = fmt.Sprintf("%s_%s", r.PeerIP, r.Prefix)
		err = nil
	}
	return ret, err
}

func (r EPEPrefix) GetType() string {
	return EPEPrefixName
}
