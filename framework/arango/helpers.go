package arango

import "fmt"

func (a *ArangoConn) GetRouterByIP(ip string) *Router {
	r := &Router{}
	q := fmt.Sprintf("FOR r in Routers FILTER r.RouterIP == %q RETURN r", ip)
	results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
		return results[len(results)-1].(*Router)
	}
	return nil
}

func (a *ArangoConn) GetRouterKeyFromInterfaceIP(ip string) string {
	if len(ip) == 0 {
		return ""
	}
	var r string
	key := "Routers/" + ip + "%"
	q := fmt.Sprintf("FOR e in LinkEdges Filter e.ToIP == %q OR e._to LIKE %q  RETURN DISTINCT e._to", ip, key)
	results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
		return results[len(results)-1].(string)
	}
	return ""

}
