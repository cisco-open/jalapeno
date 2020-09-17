package arangodb

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/tools"
	"github.com/sbezverk/topology/pkg/dbclient"
	"github.com/sbezverk/topology/pkg/locker"
)

type arangoDB struct {
	stop chan struct{}
	dbclient.DB
	*ArangoConn
	lckr locker.Locker
}

// NewDBSrvClient returns an instance of a DB server client process
func NewDBSrvClient(arangoSrv, user, pass, dbname string) (dbclient.Srv, error) {
	if err := tools.URLAddrValidation(arangoSrv); err != nil {
		return nil, err
	}
	arangoConn, err := NewArango(ArangoConfig{
		URL:      arangoSrv,
		User:     user,
		Password: pass,
		Database: dbname,
	})
	if err != nil {
		return nil, err
	}
	arango := &arangoDB{
		stop: make(chan struct{}),
		lckr: locker.NewLocker(),
	}
	arango.DB = arango
	arango.ArangoConn = arangoConn

	return arango, nil
}

func (a *arangoDB) Start() error {
	glog.Infof("Connected to arango database, starting monitor")
	go a.monitor()

	return nil
}

func (a *arangoDB) Stop() error {
	close(a.stop)

	return nil
}

func (a *arangoDB) GetInterface() dbclient.DB {
	return a.DB
}

func (a *arangoDB) GetArangoDBInterface() *ArangoConn {
	return a.ArangoConn
}

func (a *arangoDB) StoreMessage(msgType int, msg interface{}) error {
	switch msgType {
	case bmp.PeerStateChangeMsg:
		p, ok := msg.(*message.PeerStateChange)
		if !ok {
			return fmt.Errorf("malformed PeerStateChange message")
		}
		a.peerChangeHandler(p)
	case bmp.UnicastPrefixMsg:
		un, ok := msg.(*message.UnicastPrefix)
		if !ok {
			return fmt.Errorf("malformed UnicastPrefix message")
		}
		a.unicastPrefixHandler(un)
	case bmp.LSLinkMsg:
		lsl, ok := msg.(*message.LSLink)
		if !ok {
			return fmt.Errorf("malformed LSNode message")
		}
		a.lslinkHandler(lsl)
	case bmp.LSNodeMsg:
		lsn, ok := msg.(*message.LSNode)
		if !ok {
			return fmt.Errorf("malformed LSNode message")
		}
		a.lsnodeHandler(lsn)
	case bmp.LSPrefixMsg:
		lsp, ok := msg.(*message.LSPrefix)
		if !ok {
			return fmt.Errorf("malformed LSPrefix message")
		}
		a.lsprefixHandler(lsp)
	case bmp.LSSRv6SIDMsg:
		srv6sid, ok := msg.(*message.LSSRv6SID)
		if !ok {
			return fmt.Errorf("malformed LSPrefix message")
		}
		a.lsSRV6SIDHandler(srv6sid)
	case bmp.L3VPNMsg:
		l3, ok := msg.(*message.L3VPNPrefix)
		if !ok {
			return fmt.Errorf("malformed L3VPN message")
		}
		a.l3vpnHandler(l3)
	}

	return nil
}

func (a *arangoDB) monitor() {
	for {
		select {
		case <-a.stop:
			// TODO Add clean up of connection with Arango DB
			return
		}
	}
}
