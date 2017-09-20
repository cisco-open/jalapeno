package arango

const asName = "ASEdges"

type ASEdge struct {
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
	Key  string `json:"_key,omitempty"`
}

func (a *ASEdge) GetKey() string {
	return a.Key
}

func (a *ASEdge) GetType() string {
	return asName
}

func (a *ASEdge) SetEdge(to string, from string) {
	a.To = to
	a.From = from
}

func (a *ASEdge) GetEdge() (string, string) {
	return a.To, a.From
}
