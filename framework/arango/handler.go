package arango

import (
	"fmt"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka/handler"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/openbmp"
)

type ArangoHandler struct {
	fmap map[string]handler.HandlerFunc
	db   ArangoConn
}

func NewHandler(db ArangoConn) *ArangoHandler {
	a := &ArangoHandler{
		fmap: make(map[string]handler.HandlerFunc),
		db:   db,
	}
	a.fmap[openbmp.TopicPeer] = a.HandlePeer
	a.fmap[openbmp.TopicCollector] = a.HandleCollector
	a.fmap[openbmp.TopicBaseAttribute] = a.HandleBaseAttribute
	a.fmap[openbmp.TopicUnicastPrefix] = a.HandleUnicastPrefix
	a.fmap[openbmp.TopicLSNode] = a.HandleLSNode
	a.fmap[openbmp.TopicLSLink] = a.HandleLSLink
	a.fmap[openbmp.TopicLSPrefix] = a.HandleLSPrefix
	return a
}

func (a *ArangoHandler) Handle(m *openbmp.Message) {
	if f, ok := a.fmap[m.Topic.String()]; ok {
		f(m)
	} else {
		fmt.Println("uhhhh...???")
	}
}

func (a *ArangoHandler) Register(topic openbmp.Topic, f handler.HandlerFunc) {
	a.fmap[topic.String()] = f
}

func (a *ArangoHandler) RegisterDefault(f handler.HandlerFunc) {

}

func (a *ArangoHandler) HandlePeer(m *openbmp.Message) {
	fmt.Println(m)
}

func (a *ArangoHandler) HandleCollector(m *openbmp.Message) {
	if m.Action() != openbmp.ActionHeartbeat {
		fmt.Printf("Got Collector %s [seq %v] action: %v.\n", m.GetUnsafe("admin_id"), m.GetUnsafe("sequence"), m.Action())
	}
}

func (a *ArangoHandler) HandleBaseAttribute(m *openbmp.Message) {
	fmt.Println(m)
}

func (a *ArangoHandler) HandleUnicastPrefix(m *openbmp.Message) {
	fmt.Println(m)
}

func (a *ArangoHandler) HandleLSNode(m *openbmp.Message) {
	fmt.Println(m)

}

func (a *ArangoHandler) HandleLSLink(m *openbmp.Message) {
	fmt.Println(m)
}

func (a *ArangoHandler) HandleLSPrefix(m *openbmp.Message) {
	fmt.Println(m)
}

func (a *ArangoHandler) Debug() {
	fmt.Println("Debug Not Impl")
}
