package handler

import (
	"fmt"
	"sort"
	"strings"

	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/log"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
)

/* all the garbage in here will be removed soon. This is for testing. */

type HandlerFunc func(*openbmp.Message)

var index, seqN int

type Handler interface {
	Handle(*openbmp.Message)
	Register(topic openbmp.Topic, f HandlerFunc)
	RegisterDefault(f HandlerFunc)
}

type DefaultHandler struct {
	fmap     map[openbmp.Topic]HandlerFunc
	def      HandlerFunc
	errthing map[string]map[string][]*openbmp.Message
}

func NewDefault() *DefaultHandler {
	return &DefaultHandler{
		fmap:     make(map[openbmp.Topic]HandlerFunc),
		def:      nil,
		errthing: make(map[string]map[string][]*openbmp.Message),
	}
}

func (h *DefaultHandler) Handle(m *openbmp.Message) {
	log.WithField("Mess", m).Debug("Handle")
	index++
	seqN, _ = m.GetInt("sequence")
}

func (h *DefaultHandler) Register(topic openbmp.Topic, f HandlerFunc) {
	h.fmap[topic] = f
}

func (h *DefaultHandler) RegisterDefault(f HandlerFunc) {
	h.def = f
}

func (h *DefaultHandler) Debug() {
	log.Debugf("Index: %v, SeqN: %v\n", index, seqN)
}

func (h *DefaultHandler) TestPrint() {
	for k, v := range h.errthing {
		if k == "sequence" || strings.Contains(k, "hash") || k == "timestamp" {
			continue
		}
		var keys []string
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		j := 0
		tmp := ""
		if len(keys) > 100 {
			for _, k := range keys {
				if len(tmp) > 80 {
					keys[j] = tmp
					j++
					tmp = ""
					continue
				}
				tmp += k + ", "
			}
			keys = keys[0:j]
		}

		fmt.Println(k)
		for _, k := range keys {
			fmt.Printf("\t%v\n", k)
		}
		fmt.Println()
	}
}

func (h *DefaultHandler) Def(m *openbmp.Message) {
	log.WithField("Message", m).Debug()
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
