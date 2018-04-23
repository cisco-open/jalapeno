package convert

import (
	"encoding/json"
	"net/http"

	"wwwin-github.cisco.com/spa-ie/voltron/services/framework/api/v1/client"
	"wwwin-github.cisco.com/spa-ie/voltron/services/framework/database"
)

func ApiReq2DbCol(request *http.Request) (database.Collector, error) {
	var apiCol client.Collector
	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()
	if err := decoder.Decode(&apiCol); err != nil {
		return database.Collector{}, err
	}

	dbCol := ApiCol2DbCol(apiCol)
	return dbCol, dbCol.SetKey()
}

func ApiCol2DbCol(apiCol client.Collector) (dbCol database.Collector) {
	return database.Collector{
		Key:           apiCol.Name,
		Name:          apiCol.Name,
		Description:   apiCol.Description,
		Status:        apiCol.Status,
		EdgeType:      apiCol.EdgeType,
		FieldName:     apiCol.FieldName,
		Timeout:       apiCol.Timeout,
		LastHeartbeat: apiCol.LastHeartbeat,
	}
}

func DbReq2ApiCol(request *http.Request) (client.Collector, error) {
	var dbCol database.Collector
	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()
	if err := decoder.Decode(&dbCol); err != nil {
		return client.Collector{}, err
	}

	apiCol := DbCol2ApiCol(dbCol)
	return apiCol.(client.Collector), nil
}

type ConvertFunc func(interface{}) interface{}

func DbCol2ApiCol(i interface{}) interface{} {
	dbCol := i.(*database.Collector)
	return client.Collector{
		Name:          dbCol.Name,
		Description:   dbCol.Description,
		Status:        dbCol.Status,
		EdgeType:      dbCol.EdgeType,
		FieldName:     dbCol.FieldName,
		Timeout:       dbCol.Timeout,
		LastHeartbeat: dbCol.LastHeartbeat,
	}
}

func DBLinkE2ApiLinkE(i interface{}) interface{} {
	dbLinkE := i.(*database.LinkEdge)
	return client.LinkEdge{
		Key:     dbLinkE.Key,
		To:      dbLinkE.To,
		From:    dbLinkE.From,
		ToIP:    dbLinkE.ToIP,
		FromIP:  dbLinkE.FromIP,
		Netmask: dbLinkE.Netmask,
		Label:   dbLinkE.Label,
		V6:      dbLinkE.V6,
	}
}

func DBPrefixE2ApiPrefixE(i interface{}) interface{} {
	dbPrefixE := i.(*database.PrefixEdge)
	return client.PrefixEdge{
		Key:         dbPrefixE.Key,
		To:          dbPrefixE.To,
		From:        dbPrefixE.From,
		NextHop:     dbPrefixE.NextHop,
		InterfaceIP: dbPrefixE.InterfaceIP,
		ASPath:      dbPrefixE.ASPath,
		Labels:      dbPrefixE.Labels,
		BGPPolicy:   dbPrefixE.BGPPolicy,
	}
}

func DBPrefix2ApiPrefix(i interface{}) interface{} {
	dbPrefix := i.(*database.Prefix)
	return client.Prefix{
		Key:    dbPrefix.Key,
		Prefix: dbPrefix.Prefix,
		Length: int32(dbPrefix.Length),
	}
}

func DbRouter2ApiRouter(i interface{}) interface{} {
	dbRouter := i.(*database.Router)
	return client.Router{
		Key:      dbRouter.Key,
		Name:     dbRouter.Name,
		RouterIP: dbRouter.RouterIP,
		BGPID:    dbRouter.BGPID,
		IsLocal:  dbRouter.IsLocal,
		ASN:      dbRouter.ASN,
	}
}
