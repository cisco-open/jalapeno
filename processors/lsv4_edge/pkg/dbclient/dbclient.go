package dbclient

// DB defines required methods for a database client to support
type DB interface {
	ProcessMessage(msgType int, msg []byte) error
}

// Srv defines required method of a database server
type Srv interface {
	Start() error
	Stop() error
	GetInterface() DB
}
