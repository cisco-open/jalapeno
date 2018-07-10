package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Middleware wraps handlers
type Middleware func(http.Handler) http.Handler

type Router struct {
	*mux.Router
}

// Route holds route information. Handler/Middleware are net/http package compatable
// handlers.
type Route struct {
	Name       string
	Method     string
	Pattern    string
	Handler    http.HandlerFunc
	Middleware []Middleware
}

type Routes []Route

func newRouter() *Router {
	return &Router{mux.NewRouter().StrictSlash(true)}
}

func (r *Router) InitRoutes(s *Server) {
	var routes = Routes{

		Route{
			Name:       "AddCollector",
			Method:     "POST",
			Pattern:    "/v1/collectors",
			Handler:    s.AddCollector,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "DeleteCollector",
			Method:     "DELETE",
			Pattern:    "/v1/collectors/{collector-name}",
			Handler:    s.DeleteCollector,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "GetCollector",
			Method:     "GET",
			Pattern:    "/v1/collectors/{collector-name}",
			Handler:    s.GetCollector,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "GetCollectors",
			Method:     "GET",
			Pattern:    "/v1/collectors",
			Handler:    s.GetCollectors,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "GetHealthz",
			Method:     "GET",
			Pattern:    "/v1/healthz",
			Handler:    s.GetHealthz,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "GetLiveness",
			Method:     "GET",
			Pattern:    "/v1/liveness",
			Handler:    s.GetLiveness,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "GetMetrics",
			Method:     "GET",
			Pattern:    "/v1/metrics",
			Handler:    s.GetMetrics,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "GetEdge",
			Method:     "GET",
			Pattern:    "/v1/edges/{edge-type}/filter/{field-name}/{field-value}",
			Handler:    s.GetEdge,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "HeartbeatCollector",
			Method:     "GET",
			Pattern:    "/v1/collectors/{collector-name}/heartbeat",
			Handler:    s.HeartbeatCollector,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "QueryArango",
			Method:     "GET",
			Pattern:    "/v1/query/{Collection}",
			Handler:    s.QueryArango,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "RemoveAllFields",
			Method:     "DELETE",
			Pattern:    "/v1/edges/{edge-type}/names/{field-name}",
			Handler:    s.RemoveAllFields,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "RemoveField",
			Method:     "DELETE",
			Pattern:    "/v1/edges/{edge-type}/key/{edge-key}/names/{field-name}",
			Handler:    s.RemoveField,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "UpdateCollector",
			Method:     "POST",
			Pattern:    "/v1/collectors/{collector-name}",
			Handler:    s.UpdateCollector,
			Middleware: []Middleware{},
		},

		Route{
			Name:       "UpsertField",
			Method:     "PUT",
			Pattern:    "/v1/edges/{edge-type}/names/{field-name}",
			Handler:    s.UpsertField,
			Middleware: []Middleware{},
		},
	}
	for _, route := range routes {
		r.AddRoute(route)
	}

}

// AddRoute adds routes and wraps the handlers in the provided middleware.
func (r *Router) AddRoute(route Route) {
	entry := r.
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name)

	if route.Handler != nil {
		handler := route.Handler
		for _, mw := range route.Middleware {
			handler = mw(handler).(http.HandlerFunc)
		}
		entry.Handler(handler)
	}
}

// MW is a convenience function to construct Middleware arrays (shortcut for Route field construction.)
func MW(mw ...Middleware) []Middleware {
	var mwArray []Middleware
	for _, m := range mw {
		mwArray = append(mwArray, m)
	}
	return mwArray
}
