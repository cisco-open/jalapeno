package mockdb

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/topology/pkg/dbclient"
)

type mockDB struct {
	stop chan struct{}
	dbclient.DB
}

// NewDBSrvClient returns an instance of a mock DB server client process
func NewDBSrvClient(addr string) (dbclient.Srv, error) {
	m := &mockDB{
		stop: make(chan struct{}),
	}
	m.DB = m

	return m, nil
}

func (m *mockDB) Start() error {
	glog.Info("Starting Mock DB Client")
	return nil
}

func (m *mockDB) Stop() error {
	close(m.stop)

	return nil
}

func (m *mockDB) GetInterface() dbclient.DB {
	return m.DB
}

func (m *mockDB) StoreMessage(msgType int, msg interface{}) error {
	switch msgType {
	case bmp.PeerStateChangeMsg:
		p, ok := msg.(*message.PeerStateChange)
		if !ok {
			return fmt.Errorf("malformed PeerStateChange message")
		}
		m.peerChangeHandler(p)
	case bmp.UnicastPrefixMsg:
		un, ok := msg.(*message.UnicastPrefix)
		if !ok {
			return fmt.Errorf("malformed UnicastPrefix message")
		}
		m.unicastPrefixHandler(un)
	case bmp.LSNodeMsg:
		ln, ok := msg.(*message.LSNode)
		if !ok {
			return fmt.Errorf("malformed LSNode message")
		}
		m.lsNodeHandler(ln)
	case bmp.LSLinkMsg:
		ll, ok := msg.(*message.LSLink)
		if !ok {
			return fmt.Errorf("malformed LSLink message")
		}
		m.lsLinkHandler(ll)
	case bmp.L3VPNMsg:
		l3, ok := msg.(*message.L3VPNPrefix)
		if !ok {
			return fmt.Errorf("malformed L3VPN message")
		}
		m.l3vpnPrefixHandler(l3)
	}

	return nil
}

func (m *mockDB) peerChangeHandler(obj *message.PeerStateChange) {
	glog.V(5).Infof("peer change handler")
}

func (m *mockDB) unicastPrefixHandler(obj *message.UnicastPrefix) {
	glog.V(5).Infof("unicast prefix handler")
}

func (m *mockDB) lsNodeHandler(obj *message.LSNode) {
	glog.V(5).Infof("LS Node handler")
}

func (m *mockDB) lsLinkHandler(obj *message.LSLink) {
	glog.V(5).Infof("LS Link handler")
}

func (m *mockDB) l3vpnPrefixHandler(obj *message.L3VPNPrefix) {
	glog.V(5).Infof("L3VPNPrefix handler")
}
