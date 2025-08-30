package arangodb

import (
	"github.com/sbezverk/gobmp/pkg/pub"
	//	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
)

type PubArango struct {
	dbclient.Srv
}

func (p *PubArango) PublishMessage(msgType int, msgHash []byte, msg []byte) error {
	db := p.GetInterface()
	newType := dbclient.CollectionType(msgType)
	//figure out collection type from parsed message
	db.StoreMessage(newType, msg)

	return nil
}

func (p *PubArango) Stop() {
	p.Stop()
}

// NewArango returns a new instance of a publisher that uses and existing connection to ArangoDB
func NewPubArango(db dbclient.Srv) (pub.Publisher, error) {
	// should probably test that the connection is active?
	return &PubArango{db}, nil
}
