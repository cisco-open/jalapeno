package arango

import "fmt"

func (a *ArangoConn) GetRouterByIP(ip string) *Router {
	r := &Router{}
	q := fmt.Sprintf("FOR r in Routers FILTER r.RouterIP == %q RETURN r", ip)
	results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
		return results[0].(*Router)
	}
	return nil
}

func (a *ArangoConn) GetRouterKeyFromInterfaceIP(ip string) string {
	if len(ip) == 0 {
		return ""
	}
	var r string
	q := fmt.Sprintf("FOR e in LinkEdges Filter e.FromIP == %q RETURN DISTINCT e._from", ip)
	results, _ := a.Query(q, nil, r)
	if len(results) == 1 {
		return results[0].(string)
	}
	return ""

}
