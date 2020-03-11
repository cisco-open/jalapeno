package database

import (
	"fmt"
	"strings"
)

const ExternalPrefixEdgeName = "ExternalPrefixEdges"

type ExternalPrefixEdge struct {
	From            string   `json:"_from,omitempty"`
	To              string   `json:"_to,omitempty"`
	Key             string   `json:"_key,omitempty"`
        SrcRouterI      string   `json:"SrcRouter"`
	SrcRouterASN    string   `json:"SrcRouterASN,omitempty"`
	SrcIntfIP       string   `json:"EgressIntIP,omitempty"`
	DstPrefix       string   `json:"DstPrefix,omitempty"`
	DstPrefixASN    string   `json:"DstPrefixASN,omitempty"`
	DstPrefixLength int      `json:"DstPrefixLength,omitempty"`
}

func (a ExternalPrefixEdge) GetKey() (string, error) {
	if a.Key == "" {
		return a.makeKey()
	}
	return a.Key, nil
}

func (a *ExternalPrefixEdge) SetKey() error {
	k, err := a.makeKey()
	if err != nil {
		return err
	}
	a.Key = k
	return nil
}

func (a *ExternalPrefixEdge) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if a.From != "" && a.To != "" {
		ret = fmt.Sprintf("%s_%s_%s", strings.Replace(a.From, "/", "_", -1), a.SrcIntfIP, strings.Replace(a.To, "/", "_", -1))
		err = nil
	}
	return ret, err
}

func (a ExternalPrefixEdge) GetType() string {
	return ExternalPrefixEdgeName
}

