package handler

import (
	"fmt"
	"time"

	"wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/database"
	"wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/log"
	"wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/openbmp"
)

type ArangoHandler struct {
	fmap map[string]HandlerFunc
	db database.ArangoConn
	asn string
	direct_peer_asns string
	transit_provider_asns string
}

func NewArango(db database.ArangoConn, localASN string, directPeerASNS string, transitProviderASNS string) *ArangoHandler {
	a := &ArangoHandler{
		fmap: make(map[string]HandlerFunc),
		db:   db,
		asn:  localASN,
		direct_peer_asns: directPeerASNS,
		transit_provider_asns: transitProviderASNS,
	}
	a.fmap[openbmp.TopicPeer] = a.HandlePeer
	a.fmap[openbmp.TopicRouter] = a.HandleRouter
	a.fmap[openbmp.TopicCollector] = a.HandleCollector
	a.fmap[openbmp.TopicBaseAttribute] = a.HandleBaseAttribute
	a.fmap[openbmp.TopicBMPStat] = a.HandleBMPStat
	a.fmap[openbmp.TopicUnicastPrefix] = a.HandleUnicastPrefix
	a.fmap[openbmp.TopicLSNode] = a.HandleLSNode
	a.fmap[openbmp.TopicLSLink] = a.HandleLSLink
	a.fmap[openbmp.TopicLSPrefix] = a.HandleLSPrefix
	a.fmap[openbmp.TopicL3VPN] = a.HandleL3VPN
	return a
}

func (a *ArangoHandler) Handle(m *openbmp.Message) {
	ts, ok := m.GetTimestamp()
	t := time.Date(2017, 11, 16, 1, 0, 0, 0, time.UTC)
	if !ok || ts.Before(t) {
		return
	}

	if f, ok := a.fmap[m.Topic.String()]; ok {
		f(m)
	} else {
		log.WithField("Topic", m.Topic.String()).Warn("Unknown topic")
	}
}

func (a *ArangoHandler) Register(topic openbmp.Topic, f HandlerFunc) {
	a.fmap[topic.String()] = f
}

func (a *ArangoHandler) RegisterDefault(f HandlerFunc) {
	log.Debugf("Register Default")
}

func (a *ArangoHandler) HandleCollector(m *openbmp.Message) {
	fmt.Println("Handling Collector OpenBMP message")
	fmt.Println(m)
	collector(a, m)
}

func (a *ArangoHandler) HandleRouter(m *openbmp.Message) {
	fmt.Println("No need to handle Router OpenBMP message -- skipping")
	//fmt.Println(m)
	//router(a, m)
}

func (a *ArangoHandler) HandlePeer(m *openbmp.Message) {
	fmt.Println("Handling Peer OpenBMP message")
	fmt.Println(m)
	peer(a, m)
}

func (a *ArangoHandler) HandleUnicastPrefix(m *openbmp.Message) {
	fmt.Println("Handling UnicastPrefix OpenBMP message")
	//fmt.Println(m)
	//unicast_prefix(a, m)
}

func (a *ArangoHandler) HandleLSNode(m *openbmp.Message) {
	fmt.Println("Handling LSNode OpenBMP message")
	fmt.Println(m)
	ls_node(a, m)
}

func (a *ArangoHandler) HandleLSLink(m *openbmp.Message) {
	fmt.Println("Handling LSLink OpenBMP message")
	fmt.Println(m)
	ls_link(a, m)
}

func (a *ArangoHandler) HandleLSPrefix(m *openbmp.Message) {
	fmt.Println("Handling LSPrefix OpenBMP message")
	fmt.Println(m)
	ls_prefix(a, m)
}

func (a *ArangoHandler) HandleL3VPN(m *openbmp.Message) {
	fmt.Println("Handling L3VPN OpenBMP message")
	fmt.Println(m)
	l3vpn(a, m)
}

func (a *ArangoHandler) HandleBMPStat(m *openbmp.Message) {
	fmt.Println("Handling BMPStat OpenBMP message")
	//fmt.Println(m)
}

func (a *ArangoHandler) HandleBaseAttribute(m *openbmp.Message) {
	fmt.Println("Handling Base Attribute OpenBMP message")
	//fmt.Println(m)
}
