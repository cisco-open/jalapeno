package arango

const prefixName = "Prefixes"

type Prefix struct {
	Key string `json:"_key,omitempty"`
}

func (p *Prefix) GetKey() string {
	return p.Key
}

func (p *Prefix) GetType() string {
	return prefixName
}
