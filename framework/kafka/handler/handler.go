package handler

import (
	"fmt"
	"strings"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/openbmp"
)

/* all the garbage in here will be removed soon. This is for testing. */

type HandlerFunc func(*openbmp.Message)

type Handler interface {
	Handle(*openbmp.Message)
	Register(topic openbmp.Topic, f HandlerFunc)
	RegisterDefault(f HandlerFunc)
}

type DefaultHandler struct {
	fmap map[openbmp.Topic]HandlerFunc
	def  HandlerFunc

	router_ips map[string][]*openbmp.Message
	peer_ips   map[string][]*openbmp.Message
	names      map[string][]*openbmp.Message
	prefixes   map[string][]*openbmp.Message
	next_hops  map[string][]*openbmp.Message
	intf_ips   map[string][]*openbmp.Message
	nei_ips    map[string][]*openbmp.Message
	origin     map[string][]*openbmp.Message

	errthing map[string]map[string][]*openbmp.Message
}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{
		fmap:       make(map[openbmp.Topic]HandlerFunc),
		def:        nil,
		router_ips: make(map[string][]*openbmp.Message),
		peer_ips:   make(map[string][]*openbmp.Message),
		names:      make(map[string][]*openbmp.Message),
		prefixes:   make(map[string][]*openbmp.Message),
		next_hops:  make(map[string][]*openbmp.Message),
		intf_ips:   make(map[string][]*openbmp.Message),
		nei_ips:    make(map[string][]*openbmp.Message),
		origin:     make(map[string][]*openbmp.Message),
		errthing:   make(map[string]map[string][]*openbmp.Message),
	}
}

func (h *DefaultHandler) Handle(m *openbmp.Message) {
	if f, ok := h.fmap[m.Topic]; ok {
		f(m)
		return
	}
	if h.def != nil {
		h.def(m)
		return
	}
	h.Def(m)
}

func (h *DefaultHandler) Register(topic openbmp.Topic, f HandlerFunc) {
	h.fmap[topic] = f
}

func (h *DefaultHandler) RegisterDefault(f HandlerFunc) {
	h.def = f
}

func (h *DefaultHandler) TestPrint() {
	vs := []map[string][]*openbmp.Message{h.router_ips, h.peer_ips, h.names, h.prefixes, h.next_hops, h.intf_ips, h.nei_ips}
	ss := []string{"router_ips:", "peer_ips:", "names:", "prefixes:", "next_hops:", "intf_ip:", "nei_ip:", "origin:"}
	for i, v := range vs {
		fmt.Println(ss[i])
		for k := range v {
			fmt.Printf("\t%v\n", k)
		}
		fmt.Println()
	}
}

func (h *DefaultHandler) TestPrint2() {
	for k, v := range h.errthing {
		if k == "sequence" || strings.Contains(k, "hash") || k == "timestamp" {
			continue
		}
		fmt.Println(k)
		for k := range v {
			fmt.Printf("\t%v\n", k)
		}
		fmt.Println()
	}
}

func (h *DefaultHandler) Def(m *openbmp.Message) {
	if f, ok := m.Get("router_ip"); ok {
		ip := fmt.Sprintf("%v", f)
		if _, ok = h.router_ips[ip]; !ok {
			h.router_ips[ip] = []*openbmp.Message{}
		}
		h.router_ips[ip] = append(h.router_ips[ip], m)
	}

	if f, ok := m.Get("peer_ip"); ok {
		ip := fmt.Sprintf("%v", f)
		if _, ok = h.peer_ips[ip]; !ok {
			h.peer_ips[ip] = []*openbmp.Message{}
		}
		h.peer_ips[ip] = append(h.peer_ips[ip], m)
	}

	if f, ok := m.Get("name"); ok {
		name := fmt.Sprintf("%v", f)
		if _, ok = h.names[name]; !ok {
			h.names[name] = []*openbmp.Message{}
		}
		h.names[name] = append(h.names[name], m)
	}

	if f, ok := m.Get("prefix"); ok {
		name := fmt.Sprintf("%v", f)
		if _, ok = h.prefixes[name]; !ok {
			h.prefixes[name] = []*openbmp.Message{}
		}
		h.prefixes[name] = append(h.prefixes[name], m)
	}
	if f, ok := m.Get("intf_ip"); ok {
		name := fmt.Sprintf("%v", f)
		if _, ok = h.intf_ips[name]; !ok {
			h.intf_ips[name] = []*openbmp.Message{}
		}
		h.intf_ips[name] = append(h.intf_ips[name], m)
	}
	if f, ok := m.Get("nexthop"); ok {
		name := fmt.Sprintf("%v", f)
		if _, ok = h.next_hops[name]; !ok {
			h.next_hops[name] = []*openbmp.Message{}
		}
		h.next_hops[name] = append(h.next_hops[name], m)
	}
	if f, ok := m.Get("nei_ip"); ok {
		name := fmt.Sprintf("%v", f)
		if _, ok = h.nei_ips[name]; !ok {
			h.nei_ips[name] = []*openbmp.Message{}
		}
		h.nei_ips[name] = append(h.nei_ips[name], m)
	}
	if f, ok := m.Get("origin"); ok {
		name := fmt.Sprintf("%v", f)
		if _, ok = h.origin[name]; !ok {
			h.origin[name] = []*openbmp.Message{}
		}
		h.origin[name] = append(h.origin[name], m)
	}

	for k, v := range m.Fields {
		if _, ok := h.errthing[k]; !ok {
			h.errthing[k] = map[string][]*openbmp.Message{}
		}
		name := fmt.Sprintf("%v", v)
		if _, ok := h.errthing[k][name]; !ok {
			h.errthing[k][name] = []*openbmp.Message{}
		}
		h.errthing[k][name] = append(h.errthing[k][name], m)
	}

}
