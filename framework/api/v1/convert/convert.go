package convert

import (
	"encoding/json"
	"net/http"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/api/v1/client"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/database"
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
		Name:          apiCol.Name,
		Description:   apiCol.Description,
		Status:        apiCol.Status,
		EdgeType:      apiCol.EdgeType,
		FieldName:     apiCol.FieldName,
		FieldType:     apiCol.FieldType,
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
	return apiCol, nil
}

func DbCol2ApiCol(dbCol database.Collector) (apiCol client.Collector) {
	return client.Collector{
		Name:          dbCol.Name,
		Description:   dbCol.Description,
		Status:        dbCol.Status,
		EdgeType:      dbCol.EdgeType,
		FieldName:     dbCol.FieldName,
		FieldType:     dbCol.FieldType,
		Timeout:       dbCol.Timeout,
		LastHeartbeat: dbCol.LastHeartbeat,
	}
}
