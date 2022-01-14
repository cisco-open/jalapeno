package arangodb

import (
	"github.com/sbezverk/gobmp/pkg/message"
)

type flowspecArangoMessage struct {
	*message.Flowspec
}

func (fs *flowspecArangoMessage) MakeKey() string {
	return fs.SpecHash
}

func (fs *flowspecArangoMessage) UnmarshalJSON(b []byte) error {
	o := flowspecArangoMessage{}
	m := message.Flowspec{}
	if err := m.UnmarshalJSON(b); err != nil {
		return err
	}
	o.Flowspec = &m
	*fs = o

	return nil
}
