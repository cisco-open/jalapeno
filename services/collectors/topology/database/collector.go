package database

const CollectorName = "Collectors"

type Collector struct {
	Key           string `json:"_key,omitempty"`
	Name          string `json:"Name, omitempty"`
	Description   string `json:"Description, omitempty"`
	Status        string `json:"Status,omitempty"`
	EdgeType      string `json:"EdgeType,omitempty"`
	FieldName     string `json:"FieldName,omitempty"`
	Timeout       string `json:"Timeout, omitempty"`
	LastHeartbeat string `json:"LastHeartbeat, omitempty"`
}

func (p Collector) GetKey() (string, error) {
	if p.Key == "" {
		return p.makeKey()
	}
	return p.Key, nil
}

func (p *Collector) SetKey() error {
	k, err := p.makeKey()
	if err != nil {
		return err
	}
	p.Key = k
	return nil
}

func (p *Collector) makeKey() (string, error) {
	err := ErrKeyInvalid
	if p.Name != "" {
		err = nil
	}
	return p.Name, err
}

func (p Collector) GetType() string {
	return CollectorName
}
