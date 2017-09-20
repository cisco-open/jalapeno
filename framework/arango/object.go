package arango

const (
	graphName = "topology"
)

type DBObject interface {
	GetKey() string
	GetType() string
}

type EdgeObject interface {
	//To string, From string
	SetEdge(string, string)
	//To string, From string
	GetEdge() (string, string)
}
