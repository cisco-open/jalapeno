package bmp

import (
	"encoding/binary"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bgp"
	"github.com/sbezverk/gobmp/pkg/tools"
)

// PeerUpMessage defines BMPPeerUpMessage per rfc7854
type PeerUpMessage struct {
	LocalAddress []byte
	LocalPort    uint16
	RemotePort   uint16
	SentOpen     *bgp.OpenMessage
	ReceivedOpen *bgp.OpenMessage
	Information  []InformationalTLV
}

// UnmarshalPeerUpMessage processes Peer Up message and returns BMPPeerUpMessage object
func UnmarshalPeerUpMessage(b []byte) (*PeerUpMessage, error) {
	glog.V(6).Infof("BMP Peer Up Message Raw: %s", tools.MessageHex(b))
	var err error
	pu := &PeerUpMessage{
		LocalAddress: make([]byte, 16),
		SentOpen:     &bgp.OpenMessage{},
		ReceivedOpen: &bgp.OpenMessage{},
		Information:  make([]InformationalTLV, 0),
	}
	p := 0
	copy(pu.LocalAddress, b[:16])
	p += 16
	pu.LocalPort = binary.BigEndian.Uint16(b[p : p+2])
	p += 2
	pu.RemotePort = binary.BigEndian.Uint16(b[p : p+2])
	p += 2
	// Skip first marker 16 bytes
	p += 16
	l1 := int16(binary.BigEndian.Uint16(b[p : p+2]))
	pu.SentOpen, err = bgp.UnmarshalBGPOpenMessage(b[p : p+int(l1-16)])
	if err != nil {
		return nil, err
	}
	// Moving pointer to the next marker
	p += int(l1) - 16
	// Skip second marker
	p += 16
	l2 := int16(binary.BigEndian.Uint16(b[p : p+2]))
	pu.ReceivedOpen, err = bgp.UnmarshalBGPOpenMessage(b[p : p+int(l2-16)])
	if err != nil {
		return nil, err
	}
	p += int(l2) - 16
	// Last part is optional Informational TLVs
	if len(b) > int(p) {
		// Since pointer p does not point to the end of buffer,
		// then processing Informational TLVs
		tlvs, err := UnmarshalTLV(b[p : len(b)-int(p)])
		if err != nil {
			return nil, err
		}
		pu.Information = tlvs
	}
	return pu, nil
}
