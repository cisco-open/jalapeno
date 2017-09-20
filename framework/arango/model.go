package arango

const (
	prefixName = "Prefix"
	routerName = "Router"
	asName     = "ASEdge"
	linkName   = "LinkEdge"
	graphName  = "topology"
)

type Prefix struct {
	IP  string `json:"_ip,omitempty"`
	Key string `json:"_key,omitempty"`
}

type Router struct {
	Name string `json:"_name,omitempty"`
	Key  string `json:"_key,omitempty"`
}

type ASEdge struct {
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
	Name string `json:"_name,omitempty"`
	Key  string `json:"_key,omitempty"`
}

type LinkEdge struct {
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
	Name string `json:"_name,omitempty"`
	Key  string `json:"_key,omitempty"`
}
