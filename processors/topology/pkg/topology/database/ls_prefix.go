package database

import (
        "fmt"
)

const LSPrefixName = "LSPrefix"

type LSPrefix struct {
        Key          string   `json:"_key,omitempty"`
        IGPRouterID  string   `json:"IGPRouterID,omitempty"`
        Prefix       string   `json:"Prefix,omitempty"`
        Length       int32    `json:"Length,omitempty"`
        Protocol     string   `json:"Protocol,omitempty"`
        SRFlags      []string `json:"SRFlags,omitempty"`
        Algorithm    uint8    `json:"Algorithm"`
        Timestamp    string   `json:"Timestamp,omitempty"`
        SIDIndex     []int    `json:"SIDIndex,omitempty"`
}

func (r LSPrefix) GetKey() (string, error) {
        if r.Key == "" {
                return r.makeKey()
        }
        return r.Key, nil
}

func (r *LSPrefix) SetKey() error {
        k, err := r.makeKey()
        if err != nil {
                return err
        }
        r.Key = k
        return nil
}

func (r *LSPrefix) makeKey() (string, error) {
        err := ErrKeyInvalid
        ret := ""
        if r.IGPRouterID != "" {
                ret = fmt.Sprintf("%s_%s", r.IGPRouterID, r.Prefix)
                err = nil
        }
        return ret, err
}

func (r LSPrefix) GetType() string {
        return LSPrefixName
}





