package arango

const routerName = "Routers"

type Router struct {
	Key  string `json:"_key,omitempty"`
	Name string `json:"_name,omitempty"`
}

func (r *Router) GetKey() string {
	return r.Key
}

func (r *Router) GetType() string {
	return routerName
}
