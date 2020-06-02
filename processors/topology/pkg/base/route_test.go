package base

import (
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestGetPrefixIPReachability(t *testing.T) {
	tests := []struct {
		name   string
		input  *PrefixDescriptor
		expect string

		ipv4 bool
	}{
		{
			name: "ipv4",
			input: &PrefixDescriptor{
				PrefixTLV: map[uint16]TLV{
					265: {
						Type:   265,
						Length: 5,
						Value:  []byte{32, 192, 168, 8, 8},
					},
				},
			},
			expect: "192.168.8.8",
			ipv4:   true,
		},
		{
			name: "ipv6",
			input: &PrefixDescriptor{
				PrefixTLV: map[uint16]TLV{
					265: {
						Type:   265,
						Length: 16,
						Value:  []byte{120, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
			},
			expect: "10::",
			ipv4:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := tt.input.GetPrefixIPReachability(tt.ipv4)
			var rs string
			if route == nil {
				t.Errorf("failed, no routes reterned")
			}
			if len(route.Prefix) == 16 {
				rs = net.IP(route.Prefix).To16().String()
			} else {
				rs = net.IP(route.Prefix).To4().String()
			}
			if strings.Compare(tt.expect, rs) != 0 {
				t.Errorf("failed, expected %s route got %s route", tt.expect, rs)
			}
		})
	}
}

func TestUnmarshalBaseNLRI(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		expect []Route
	}{
		{
			name:  "Default prefix",
			input: []byte{0x0},
			expect: []Route{
				{
					Length: 0x0,
					Prefix: []byte{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRoutes(tt.input)
			if err != nil {
				t.Fatalf("test failed with error: %+v", err)
			}
			if !reflect.DeepEqual(tt.expect, got) {
				t.Fatalf("test failed as expected nlri %+v does not match actual nlri %+v", tt.expect, got)
			}
		})
	}
}
