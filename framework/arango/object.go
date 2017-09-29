package arango

import "errors"

const (
	graphName = "topology"
)

var (
	ErrKeyInvalid = errors.New("Error: Key value is invalid")
	ErrKeyChange  = errors.New("Error: Key value changed")
)

type DBObject interface {
	GetKey() string
	SetKey() error
	GetType() string
}

type EdgeObject interface {
	//To, From
	SetEdge(DBObject, DBObject)
}
