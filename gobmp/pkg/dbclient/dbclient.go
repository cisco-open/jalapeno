package dbclient

import "github.com/sbezverk/gobmp/pkg/bmp"

// DB defines required methods for a database client to support
type DB interface {
	StoreMessage(msgType CollectionType, msg []byte) error
}

// Srv defines required method of a database server
type Srv interface {
	Start() error
	Stop() error
	GetInterface() DB
}

type CollectionType int

const (
	PeerStateChange CollectionType = bmp.PeerStateChangeMsg
	LSLink          CollectionType = bmp.LSLinkMsg
	LSNode          CollectionType = bmp.LSNodeMsg
	LSPrefix        CollectionType = bmp.LSPrefixMsg
	LSSRv6SID       CollectionType = bmp.LSSRv6SIDMsg
	L3VPN           CollectionType = bmp.L3VPNMsg
	L3VPNV4         CollectionType = bmp.L3VPNV4Msg
	L3VPNV6         CollectionType = bmp.L3VPNV6Msg
	UnicastPrefix   CollectionType = bmp.UnicastPrefixMsg
	UnicastPrefixV4 CollectionType = bmp.UnicastPrefixV4Msg
	UnicastPrefixV6 CollectionType = bmp.UnicastPrefixV6Msg
)
