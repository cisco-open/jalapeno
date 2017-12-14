package database

import (
	"fmt"
	"strings"
)

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
	key := "Routers/" + ip
	col := LinkEdgeNamev4
	if strings.Contains(ip, ":") {
		col = LinkEdgeNamev6
	}
	q := fmt.Sprintf("FOR e in %s Filter e.ToIP == %q OR e._to == %q  RETURN DISTINCT e._to", col, ip, key)
	results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
		return results[len(results)-1].(string)
	}
	return ""

}
