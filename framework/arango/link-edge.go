package arango

const linkName = "LinkEdges"

type LinkEdge struct {
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
	Key  string `json:"_key,omitempty"`
}

func (l *LinkEdge) GetKey() string {
	return l.Key
}

func (l *LinkEdge) GetType() string {
	return linkName
}

func (l *LinkEdge) SetEdge(to string, from string) {
	l.To = to
	l.From = from
}

func (l *LinkEdge) GetEdge() (string, string) {
	return l.To, l.From
}
