/*
	Copyright Cisco Systems 2017
	Maintained by Mike Napolitano <mnapolit@cisco.com>
	License TBD.
*/

package openbmp

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"
)

type Topic string

func (t Topic) String() string {
	return string(t)
}

type Action string

func (a Action) String() string {
	return string(a)
}

const (
	TopicCollector     = "openbmp.parsed.collector"
	TopicRouter        = "openbmp.parsed.router"
	TopicPeer          = "openbmp.parsed.peer"
	TopicBMPStat       = "openbmp.parsed.bmp_stat"
	TopicBaseAttribute = "openbmp.parsed.base_attribute"
	TopicUnicastPrefix = "openbmp.parsed.unicast_prefix"
	TopicLSNode        = "openbmp.parsed.ls_node"
	TopicLSLink        = "openbmp.parsed.ls_link"
	TopicLSPrefix      = "openbmp.parsed.ls_prefix"
	TopicL3VPN         = "openbmp.parsed.l3vpn"
	TopicEVPN          = "openbmp.parsed.evpn"
	TopicInvalid       = "invalid"
)

var topics = []string{
	string(TopicCollector), string(TopicRouter), string(TopicPeer), string(TopicBMPStat),
	string(TopicBaseAttribute), string(TopicUnicastPrefix), string(TopicLSNode),
	string(TopicLSPrefix), string(TopicL3VPN), string(TopicEVPN), string(TopicLSLink),
}

const (
	ActionAdd       = "add"
	ActionDel       = "delete"
	ActionStarted   = "started"
	ActionChange    = "change"
	ActionHeartbeat = "heartbeat"
	ActionStopped   = "stopped"
	ActionFirst     = "first"
	ActionInit      = "init"
	ActionTerm      = "term"
	ActionInvalid   = "invalid"
)

func AllTopics() []string {
	return topics
}

func Topics(t ...Topic) []string {
	var ts []string
	for _, t := range t {
		ts = append(ts, string(t))
	}
	return ts
}

type Message struct {
	Topic
	Fields map[string]interface{}
}

func parseTopic(topic string) Topic {
	for _, t := range topics {
		if strings.Contains(t, topic) {
			return Topic(t)
		}
	}
	return TopicInvalid
}

// NewMessage creates a parsed BMP message of topic `topic` with
// contents `value`
func NewMessage(topic string, value []byte) *Message {
	typ := parseTopic(topic)
	if typ == TopicInvalid {
		return nil
	}
	bs := bytes.Split(value, []byte("\n\n"))
	if len(bs) != 2 {
		return nil
	}

	fields := strings.Split(string(bs[1]), "\t")
	heads := headers[string(typ)]
	if len(fields) != len(heads) {
		return nil
	}
	message := &Message{
		Topic:  typ,
		Fields: map[string]interface{}{},
	}

	// TODO: Distinguish between empty/nil values
	for i, h := range heads {
		message.Fields[h] = strings.TrimSpace(fields[i])
	}
	return message
}

func (m Message) Action() Action {
	a, ok := m.GetString("action")
	if !ok {
		return ActionInvalid
	}
	return Action(a)
}

// IsTopic checks if message is one of the supplied topics
func (m Message) IsTopic(t ...Topic) bool {
	for _, top := range t {
		if top == m.Topic {
			return true
		}
	}
	return false
}

// IsAction checks if message is one of the supplied Actions
func (m Message) IsAction(a ...Action) bool {
	myAction := m.Action()
	for _, act := range a {
		if act == myAction {
			return true
		}
	}
	return false
}

func (m Message) Has(field string) bool {
	_, ok := m.Fields[field]
	return ok
}

func (m Message) Get(field string) (interface{}, bool) {
	field = strings.ToLower(field)
	field = strings.Replace(field, " ", "_", -1)
	v, ok := m.Fields[field]
	if !ok {
		return nil, false
	}
	return v, true
}

func (m Message) GetString(field string) (string, bool) {
	v, ok := m.Get(field)
	if ok {
		s, ok := v.(string)
		if s == "" {
			return "", false
		}
		return s, ok
	}
	return "", false
}

func (m Message) GetIP(field string) (net.IP, bool) {
	v, ok := m.GetString(field)
	if ok {
		return net.ParseIP(v), true
	}
	return net.IP{}, false
}

func (m Message) GetInt(field string) (int64, bool) {
	v, ok := m.Get(field)
	if !ok {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() < reflect.Int && rv.Kind() > reflect.Float64 {
		return 0, false
	}
	rv = rv.Convert(reflect.TypeOf(int64(0)))
	if !rv.CanInterface() {
		return 0, false
	}
	return rv.Interface().(int64), true
}

func (m Message) GetFloat(field string) (float64, bool) {
	v, ok := m.Get(field)
	if !ok {
		return 0.0, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() < reflect.Int && rv.Kind() > reflect.Float64 {
		return 0.0, false
	}
	rv = rv.Convert(reflect.TypeOf(float64(0.0)))
	if !rv.CanInterface() {
		return 0, false
	}
	return rv.Interface().(float64), true
}

func (m Message) GetBool(field string) (bool, bool) {
	v, ok := m.Get(field)
	if !ok {
		return false, false
	}
	if b, ok := v.(bool); ok {
		return b, true
	}

	if i, ok := m.GetInt(field); ok {
		return i == 1, true
	}
	return false, false
}

func (m Message) GetTimestamp() (time.Time, bool) {
	v, ok := m.GetString("timestamp")
	if !ok {
		return time.Time{}, false
	}
	//Mon Jan 2 15:04:05 MST 2006
	if t, err := time.Parse("2006-01-02 15:04:05.000000", v); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func (m Message) GetSequence() (int64, bool) {
	return m.GetInt("sequence")
}

func (m Message) GetUnsafe(field string) interface{} {
	v, _ := m.Get(field)
	return v
}

func (m Message) String() string {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s: \n", m.Topic.String())
	for k, v := range m.Fields {
		fmt.Fprintf(&buffer, "\t%s: %v\n", k, v)
	}
	fmt.Fprintln(&buffer)
	return buffer.String()
}
