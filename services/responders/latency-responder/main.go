package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/api/v1/client"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/database"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/manager"
)

var (
	api          *client.DefaultApi
	db           database.ArangoConn
	err          error
	dbUser       *string
	dbPass       *string
	dbURL        *string
	dbName       *string
	apiURL       *string
	frameworkURL *string
)

func init() {
	dbUser = flag.String("db-user", "root", "The Arango DB User.")
	dbPass = flag.String("db-pass", "voltron", "The Arango DB Password.")
	dbURL = flag.String("db-url", "127.0.0.1:8529", "The Arango DB URL.")
	dbName = flag.String("db-name", "voltron", "The Arango database name to use.")
	apiURL = flag.String("api-url", "127.0.0.1:8080", "The URL to listen on.")
	frameworkURL = flag.String("framework-url", "127.0.0.1:8876:", "The URL of the Framework API.")
}

func main() {
	flag.Parse()

	cfg := database.NewConfig()
	cfg.Database = *dbName
	cfg.Password = *dbPass
	cfg.User = *dbUser
	cfg.URL = fmt.Sprintf("http://%s", *dbURL)
	db, err = database.NewArango(cfg)
	r := mux.NewRouter()
	api = client.NewDefaultApiWithBasePath(fmt.Sprintf("http://%s/v1", *frameworkURL))

	r.HandleFunc("/scores", GetScores)
	r.HandleFunc("/labels", GetLabels)
	svr := &http.Server{
		Handler:      r,
		Addr:         *apiURL,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Printf("Responder listening on %s\n", *apiURL)
	svr.ListenAndServe()
}

func GetScores(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/scores called")
	cols, _, err := api.GetCollectors("", "", manager.StatusRunning, "", "", "", "")
	if err != nil {
		http.Error(w, "Failed to fetch collectors", http.StatusInternalServerError)
		return
	}

	scores := make([]string, len(cols))
	for i, c := range cols {
		scores[i] = c.FieldName
	}

	if err = json.NewEncoder(w).Encode(scores); err != nil {
		fmt.Printf("Failed to encode scores")
	}
}

func GetLabels(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/labels Called")
	q := r.URL.Query()
	weight_attribute, ok := q["weight_attribute"]
	if !ok || len(weight_attribute) != 1 {
		fmt.Println("Couldn't find weight_attribute in url")
		http.Error(w, "Couldn't find weight_attribute in url", http.StatusBadRequest)
		return
	}
	router_src, ok := q["router_src"]
	if !ok || len(router_src) != 1 {
		fmt.Println("Couldn't find router_src in url")
		http.Error(w, "Couldn't find router_src in url", http.StatusBadRequest)
		return
	}
	prefix_dst, ok := q["prefix_dst"]
	if !ok || len(prefix_dst) != 1 {
		fmt.Println("Couldn't find prefix_dst in url")
		http.Error(w, "Couldn't find prefix_dst in url", http.StatusBadRequest)
		return
	}
	default_weight, ok := q["default_weight"]
	if !ok || len(default_weight) != 1 {
		fmt.Println("Couldn't find default_weight in url")
		http.Error(w, "Couldn't find default_weight in url", http.StatusBadRequest)
		return
	}
	isV6 := false
	version, ok := q["version"]
	if ok && len(version) == 1 {
		if strings.ToLower(version[0]) == "v6" {
			isV6 = true
		}
	}

	isValid, err := fieldValidForQuery(weight_attribute[0])
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isValid {
		fmt.Println("Field name not valid")
		http.Error(w, "Field name not valid", http.StatusBadRequest)
		return
	}

	labels, err := queryShortest(isV6, router_src[0], prefix_dst[0], weight_attribute[0], default_weight[0])
	if err != nil {
		fmt.Println("Failed to find shortest path")
		http.Error(w, "Failed to find shortest path", http.StatusInternalServerError)
		return
	}

	var ret []string
	for _, l := range labels {
		if l != nil {
			ret = append(ret, l.(string))
		}
	}

	if ret == nil {
		ret = make([]string, 0)
	}
	if err = json.NewEncoder(w).Encode(ret); err != nil {
		fmt.Println("Failed to encode labels")
	}
}

func queryShortest(isV6 bool, from string, to string, field string, weight string) ([]interface{}, error) {
	edge := "LinkEdgesV4"
	if isV6 {
		edge = "LinkEdgesV6"
	}

	q := fmt.Sprintf(`return FLATTEN(FOR v,e in OUTBOUND SHORTEST_PATH @from TO @to %s,PrefixEdges
		OPTIONS {weightAttribute: @attribute, defaultWeight: @default_weight} RETURN [e.Label, v.Label])`, edge)
	bind := map[string]interface{}{
		"from":           fmt.Sprintf("Routers/%s", from),
		"to":             fmt.Sprintf("Prefixes/%s", to),
		"attribute":      field,
		"default_weight": weight,
	}
	var a string
	labels, err := db.Query(q, bind, a)
	if err != nil {
		return nil, errors.New("Failed to query")
	}
	if len(labels) != 1 {
		return nil, errors.New("Failed to query")
	}

	return labels[0].([]interface{}), err
}

// should use the framework API. not calling DB directly. This should not know the structure of Collectors
func fieldValidForQuery(fieldName string) (bool, error) {
	cols, _, err := api.GetCollectors("", "", manager.StatusRunning, "", fieldName, "", "")
	if err != nil {
		return false, errors.New("Failed to fetch Collectors")
	}

	if len(cols) != 1 {
		return false, errors.New("Failed to find one running collector matching field")
	}

	return true, nil
}
