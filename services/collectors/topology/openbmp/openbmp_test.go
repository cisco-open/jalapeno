package openbmp

import (
	"net"
	"reflect"
	"testing"
	"time"
)

func TestTopics(t *testing.T) {
	result1 := Topics(TopicBMPStat, TopicBaseAttribute)
	expect1 := []string{TopicBMPStat, TopicBaseAttribute}
	if !reflect.DeepEqual(expect1, result1) {
		t.Errorf("topics not equal: Expected: %v. Received: %v", expect1, result1)
	}

	result2 := Topics()
	if len(result2) != 0 {
		t.Errorf("topics not equal, expected empty. Received: %v", result2)
	}
}

func TestParseTopic(t *testing.T) {
	tests := []struct {
		topic    string
		expected Topic
	}{
		{
			topic:    "openbmp.parsed.collector",
			expected: TopicCollector,
		},
		{
			topic:    "this is a bad value",
			expected: TopicInvalid,
		},
	}

	for i, test := range tests {
		result := parseTopic(test.topic)
		if test.expected != result {
			t.Errorf("Test: %d. Expected: %q. Received: %q", i, test.expected, result)
		}
	}
}

func TestNewMessage(t *testing.T) {
	tests := []struct {
		topic string
		value []byte
		mess  Message
		isErr bool
	}{
		{
			topic: "openbmp.parsed.unicast_prefix",
			value: []byte(
				`V: 1.6
        C_HASH_ID: 93e1d1f47bb0693e4d8aea398b5c4ebe
        T: peer
        L: 410
        R: 1


        add	0	2905310a01881126ae7eed46f0d2e183	92f9871845135c3f4666d8125e7371be	10.20.0.58	437464b1dd0e8d19061a0f83346348f2	fda2f2664714e3803eb046eb1706dbe9	10.1.1.1	100000	2017-09-18 16:03:41.683572	10.1.1.4	32	1	igp		0	0	10.1.1.4	0	100				10.1.1.1	0	1	10.1.1.4	0	3	1	1`,
			),
			mess: Message{
				Topic: TopicUnicastPrefix,
				Fields: map[string]interface{}{
					"timestamp":         "2017-09-18 16:03:41.683572",
					"action":            "add",
					"as_path":           "",
					"as_path_count":     "0",
					"med":               "0",
					"ext_communit_list": nil,
					"isatomicagg":       "0",
					"router_hash":       "92f9871845135c3f4666d8125e7371be",
					"peer_asn":          "100000",
					"nexthop":           "10.1.1.4",
					"aggregator":        "",
					"is_adj_rib_in":     "1",
					"hash":              "2905310a01881126ae7eed46f0d2e183",
					"router_ip":         "10.20.0.58",
					"prefix":            "10.1.1.4",
					"path_id":           "0",
					"sequence":          "0",
					"base_attr_hash":    "437464b1dd0e8d19061a0f83346348f2",
					"peer_ip":           "10.1.1.1",
					"origin":            "igp",
					"is_nexthop_ipv":    "1",
					"isprepolicy":       "1",
					"origin_as":         "0",
					"local_pref":        "100",
					"labels":            "3",
					"peer_hash":         "fda2f2664714e3803eb046eb1706dbe9",
					"community_list":    "",
					"cluster_list":      "10.1.1.1",
					"originator_id":     "10.1.1.4",
					"is_ipv":            "1",
					"prefix_len":        "32",
				},
			},
			isErr: false,
		},
		// bad message (not enough fields)
		{
			topic: "openbmp.parsed.unicast_prefix",
			value: []byte(
				`V: 1.6
        C_HASH_ID: 93e1d1f47bb0693e4d8aea398b5c4ebe
        T: peer
        L: 410
        R: 1


        add	0`,
			),
			mess:  Message{},
			isErr: true,
		},
		// bad message2 (new lines)
		{
			topic: "openbmp.parsed.unicast_prefix",
			value: []byte(
				`V: 1.6
        C_HASH_ID: 93e1d1f47bb0693e4d8aea398b5c4ebe
        T: peer
        L: 410
        R: 1
        add	0`,
			),
			mess:  Message{},
			isErr: true,
		},
		// bad message2 (bad topic)
		{
			topic: "banana",
			value: []byte(
				`V: 1.6
        C_HASH_ID: 93e1d1f47bb0693e4d8aea398b5c4ebe
        T: peer
        L: 410
        R: 1

        add	0`,
			),
			mess:  Message{},
			isErr: true,
		},
	}

	for i, test := range tests {
		result := NewMessage(test.topic, test.value)
		if result == nil {
			if !test.isErr {
				t.Errorf("Test %d Error when not expected", i)
			}
			continue
		}

		if result.Topic != test.mess.Topic {
			t.Errorf("Topic error: Expected: %q. Received: %q", result.Topic, test.mess.Topic)
		}
		for k, _ := range test.mess.Fields {
			testV, _ := test.mess.Get(k)
			resV, _ := result.Get(k)
			if testV != resV {
				t.Errorf("Message Fields not equal: Ket: %v Expected: %v Received: %v", k, testV, resV)
			} else {
				delete(test.mess.Fields, k)
				delete(result.Fields, k)
			}
		}
	}
}

func TestAction(t *testing.T) {
	tests := []struct {
		mess    Message
		expects Action
	}{
		{
			mess: Message{
				Fields: map[string]interface{}{
					"action": "add",
				}},
			expects: ActionAdd,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{},
			},
			expects: ActionInvalid,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"action": "pineapple-pen",
				},
			},
			expects: Action("pineapple-pen"),
		},
	}

	for i, test := range tests {
		result := test.mess.Action()
		if result != test.expects {
			t.Errorf("Test: %d Failed. Expected: %v. Received: %v", i, test.expects, result)
		}
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		mess      Message
		fieldName string
		expects   string
		ok        bool
	}{
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "This is an example string",
				},
			},
			fieldName: "example",
			expects:   "This is an example string",
			ok:        true,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "This is an example string",
				},
			},
			fieldName: "not found",
			expects:   "",
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"empty": "",
				},
			},
			fieldName: "empty",
			expects:   "",
			ok:        false,
		},
	}

	for i, test := range tests {
		result := test.mess.GetStr(test.fieldName)
		result2, ok := test.mess.GetString(test.fieldName)
		if test.ok {
			if result != test.expects {
				t.Errorf("Test: %d GetStr failed. Expected: %q. Received: %q", i, test.expects, result)
			}
			if result2 != test.expects || test.ok != ok {
				t.Errorf("Test: %d GetString failed. Expected: %q. Received: %q", i, test.expects, result)
			}
		}
		if test.ok != ok {
			t.Errorf("Test: %d. OK Mismatch. Expected: %t. Received: %t", i, test.ok, ok)
		}
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		mess      Message
		fieldName string
		expects   int
		ok        bool
	}{
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "0",
				},
			},
			fieldName: "example",
			expects:   0,
			ok:        true,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "not an int",
				},
			},
			fieldName: "example",
			expects:   0,
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "This is an example string",
				},
			},
			fieldName: "not found",
			expects:   0,
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"empty": "",
				},
			},
			fieldName: "empty",
			expects:   0,
			ok:        false,
		},
	}

	for i, test := range tests {
		result, ok := test.mess.GetInt(test.fieldName)
		if test.ok {
			if result != test.expects {
				t.Errorf("Test: %d GetInt failed. Expected: %d. Received: %d", i, test.expects, result)
			}
		}
		if test.ok != ok {
			t.Errorf("Test: %d. OK Mismatch. Expected: %t. Received: %t", i, test.ok, ok)
		}
	}
}

func TestGetFloat(t *testing.T) {
	tests := []struct {
		mess      Message
		fieldName string
		expects   float64
		ok        bool
	}{
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "213.34",
				},
			},
			fieldName: "example",
			expects:   213.34,
			ok:        true,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "213.sdaf",
				},
			},
			fieldName: "example",
			expects:   0.0,
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "This is an example string",
				},
			},
			fieldName: "not found",
			expects:   0,
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"empty": "",
				},
			},
			fieldName: "empty",
			expects:   0,
			ok:        false,
		},
	}

	for i, test := range tests {
		result, ok := test.mess.GetFloat(test.fieldName)
		if test.ok {
			if result != test.expects {
				t.Errorf("Test: %d GetFloat failed. Expected: %v. Received: %v", i, test.expects, result)
			}
		}
		if test.ok != ok {
			t.Errorf("Test: %d. OK Mismatch. Expected: %t. Received: %t", i, test.ok, ok)
		}
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		mess      Message
		fieldName string
		expects   bool
		ok        bool
	}{
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "True",
				},
			},
			fieldName: "example",
			expects:   true,
			ok:        true,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "TrUe",
				},
			},
			fieldName: "example",
			expects:   true,
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "213.sdaf",
				},
			},
			fieldName: "example",
			expects:   false,
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "This is an example string",
				},
			},
			fieldName: "not found",
			expects:   false,
			ok:        false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"empty": "",
				},
			},
			fieldName: "empty",
			expects:   false,
			ok:        false,
		},
	}

	for i, test := range tests {
		result, ok := test.mess.GetBool(test.fieldName)
		if test.ok {
			if result != test.expects {
				t.Errorf("Test: %d GetBool failed. Expected: %t. Received: %t", i, test.expects, result)
			}
		}
		if test.ok != ok {
			t.Errorf("Test: %d. OK Mismatch. Expected: %t. Received: %t", i, test.ok, ok)
		}
	}
}

func TestGetTimestamp(t *testing.T) {
	goodTime, err := time.Parse("2006-01-02 15:04:05.000000", "2006-01-02 15:04:05.000000")
	if err != nil {
		t.Fatal("Couldn't parse goodTime")
	}
	tests := []struct {
		mess    Message
		expects time.Time
		ok      bool
	}{
		{
			mess: Message{
				Fields: map[string]interface{}{
					"timestamp": "2006-01-02 15:04:05.000000",
				},
			},
			expects: goodTime,
			ok:      true,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"timestamp": "bad time",
				},
			},
			expects: time.Time{},
			ok:      false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"example": "This is an example string",
				},
			},
			expects: time.Time{},
			ok:      false,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"empty": "",
				},
			},
			expects: time.Time{},
			ok:      false,
		},
	}

	for i, test := range tests {
		result, ok := test.mess.GetTimestamp()
		if test.ok {
			if result != test.expects {
				t.Errorf("Test: %d GetBool failed. Expected: %t. Received: %t", i, test.expects, result)
			}
		}
		if test.ok != ok {
			t.Errorf("Test: %d. OK Mismatch. Expected: %t. Received: %t", i, test.ok, ok)
		}
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		mess      Message
		fieldName string
		expects   string
		ok        bool
	}{
		{
			mess: Message{
				Fields: map[string]interface{}{
					"ip": "127.0.0.1",
				},
			},
			fieldName: "ip",
			expects:   "127.0.0.1",
			ok:        true,
		},
		{
			mess: Message{
				Fields: map[string]interface{}{
					"ip": "banana",
				},
			},
			fieldName: "ip",
			expects:   "",
			ok:        true,
		}, {
			mess: Message{
				Fields: map[string]interface{}{
					"ip": "banana",
				},
			},
			fieldName: "badKEY",
			expects:   "",
			ok:        false,
		},
	}

	for i, test := range tests {
		result, ok := test.mess.GetIP(test.fieldName)
		if test.ok {
			if !result.Equal(net.ParseIP(test.expects)) {
				t.Errorf("Test: %d GetIP failed. Expected: %q. Received: %q", i, test.expects, result)
			}
		}
		if test.ok != ok {
			t.Errorf("Test: %d. OK Mismatch. Expected: %t. Received: %t", i, test.ok, ok)
		}
	}
}
