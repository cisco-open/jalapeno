// Server is the api for the framework. It tracks collectors and lets responders know about collectors.
package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/api/v1/client"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/api/v1/convert"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/database"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
)

var (
	ErrNoCert = errors.New("Config needs a certificate.")
	ErrNoKey  = errors.New("Config needs a key.")
)

// Config is the basic configuration required to start an API server
type Config struct {
	Url  string `desc:"The IP:Port of the REST API. 127.0.0.1:8876"`
	Cert string `desc:"Path to certificate file for HTTPS"`
	Key  string `desc:"Path to PEM file for HTTPS"`
}

// Server interal struct that holds a database connection and connects routes to handlers
type Server struct {
	cfg       Config
	router    *Router
	quit      chan bool
	tlsConfig tls.Config
	secure    bool
	db        database.Database
}

// NewConfig returns a server.Config
func NewConfig() Config {
	//88766 == vtrn
	return Config{Url: "127.0.0.1:8876"}
}

// New instantiates a new agent API server
func New(cfg Config, db database.Database) (*Server, error) {
	quit := make(chan bool)
	s := &Server{cfg, newRouter(), quit, tls.Config{}, false, db}
	if (cfg.Cert != "") || (cfg.Key != "") {
		if _, err := os.Stat(cfg.Cert); os.IsNotExist(err) {
			return nil, ErrNoCert
		}
		if _, err := os.Stat(cfg.Key); os.IsNotExist(err) {
			return nil, ErrNoKey
		}
		// Try loading the key pair
		cer, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
		if err != nil {
			return nil, err
		}
		// This may need more options
		// (http://www.levigross.com/2015/11/21/mutual-tls-authentication-in-go/)
		s.tlsConfig.Certificates = []tls.Certificate{cer}
		s.secure = true
	}

	s.router.InitRoutes(s)
	return s, nil
}

// ListenAndServe starts the API on the specified URL
func (s *Server) ListenAndServe(errs chan error) {
	server := &http.Server{Addr: s.cfg.Url, Handler: s.router}
	server.SetKeepAlivesEnabled(false)
	var err error
	var listener net.Listener
	if !s.secure {
		listener, err = net.Listen("tcp", s.cfg.Url)
	} else {
		listener, err = tls.Listen("tcp", s.cfg.Url, &s.tlsConfig)
	}

	if err != nil {
		log.Error(err)
		errs <- err
		return
	}

	if err = server.Serve(listener); err != nil {
		log.Error(err)
		errs <- err
	}
}

// Start is the service interface implementation that lets the main routine start/monitor/shutdown the server
func (s *Server) Start() error {
	var err error
	errs := make(chan error)
	go s.ListenAndServe(errs)
	proto := "HTTP"
	if s.secure {
		proto = "HTTPS"
	}
	log.Infof("Voltron Framework %s API On: %s", proto, s.cfg.Url)

	select {
	case err = <-errs:
		log.WithError(err).Error("Voltron Framework API ListenAndServe")
	case <-s.quit:
		log.Infof("Voltron Framework API Server Quit Received.")
	}
	log.Info("Voltron Framework API Server Shutdown")
	return err
}

// Stop is the service interface implementation that lets the main routine to shutdown the server
func (s *Server) Stop() {
	close(s.quit)
}

// AddCollector responds to POSTS and parses the body and adds valid collectors to the db
func (s *Server) AddCollector(w http.ResponseWriter, r *http.Request) {
	collector, err := convert.ApiReq2DbCol(r)
	if err != nil {
		returnErr(w, "parsing collector", err, http.StatusBadRequest)
		return
	}
	defer log.Infof("AddCollector: %+v", collector)

	// validate collector edges and types and things exist
	// Fieldname is not overlapping
	err = s.validateCollector(collector)
	if err != nil {
		returnErr(w, "validating collector", err, http.StatusBadRequest)
		return
	}

	readDB := database.Collector{Name: collector.Name}
	s.db.Read(&readDB)

	now := time.Now().Format(time.RFC3339)
	collector.LastHeartbeat = now
	readDB.LastHeartbeat = now

	if reflect.DeepEqual(readDB, collector) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusAlreadyReported)
		return
	}

	err = s.db.Insert(&collector)
	if err != nil {
		if strings.Contains(err.Error(), "_key") {
			err = errors.New("collector name already exists")
		}
		returnErr(w, "inserting collector", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
}

// DeleteCollector IDs collectors by name in the path and removes that collector if it exists
func (s *Server) DeleteCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["collector-name"]
	if !ok || name == "" {
		returnErr(w, "parsing collector-name from path", nil, http.StatusBadRequest)
		return
	}
	dbObj := database.Collector{Name: name}
	err := s.db.Read(&dbObj)
	if err != nil {
		returnErr(w, "getting collector", err, http.StatusInternalServerError)
		return
	}
	defer log.Infof("DeleteCollector: %+v", convert.DbCol2ApiCol(dbObj))

	err, code := s.removeAllFields(dbObj.EdgeType, dbObj.FieldName)
	if err != nil {
		returnErr(w, "failed to remove fields", err, code)
		return
	}

	err = s.db.Delete(&dbObj)
	if err != nil {
		returnErr(w, "deleting document", err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetCollector takes in the name and updates a collector based on the body provided
func (s *Server) GetCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["collector-name"]
	if !ok || name == "" {
		returnErr(w, "parsing collector-name from path", nil, http.StatusBadRequest)
		return
	}
	dbObj := database.Collector{Name: name}
	err := s.db.Read(&dbObj)
	if err != nil {
		returnErr(w, "reading document", err, http.StatusInternalServerError)
		return
	}
	defer log.Infof("GetCollector: %+v", convert.DbCol2ApiCol(dbObj))

	if err = json.NewEncoder(w).Encode(convert.DbCol2ApiCol(dbObj)); err != nil {
		log.Errorf("Unable to Encode Sites: %v", err)
	}
}

// GetCollectors returns all known collectors
func (s *Server) GetCollectors(w http.ResponseWriter, r *http.Request) {
	defer log.Infof("GetCollectors")
	var col database.Collector
	q, bind := prepareQuery(r.URL.Query())
	dbCols, err := s.db.Query(q, bind, col)
	if err != nil {
		returnErr(w, "fetching colectors", err, http.StatusInternalServerError)
		return
	}

	var apiCols []client.Collector
	for _, c := range dbCols {
		apiCols = append(apiCols, convert.DbCol2ApiCol(*c.(*database.Collector)))
	}

	if len(apiCols) == 0 {
		apiCols = make([]client.Collector, 0)
	}
	if err = json.NewEncoder(w).Encode(apiCols); err != nil {
		log.Errorf("Unable to Encode Sites: %v", err)
	}
}

func prepareQuery(vals url.Values) (string, map[string]interface{}) {
	baseQuery := "FOR c in Collectors %s RETURN c"
	filterInsert := "FILTER %s"
	var insert bytes.Buffer
	bind := make(map[string]interface{})
	var q string
	count := 0
	for k, v := range vals {
		if len(v) == 1 && v[0] != "" {
			if count == 0 {
				insert.WriteString(fmt.Sprintf("c.%s == @%s", k, k))
			} else {
				insert.WriteString(fmt.Sprintf(" and c.%s == @%s", k, k))
			}
			bind[k] = v[0]
			count++
		}
	}
	if len(bind) != 0 {
		q = fmt.Sprintf(baseQuery, fmt.Sprintf(filterInsert, insert.String()))
	} else {
		q = "FOR c in Collectors RETURN c"
	}
	return q, bind
}

// GetHealthz is the health endpoint to be used for kubernetes
func (s *Server) GetHealthz(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Healthz")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetLiveness is the liveness endpoint for kubernetes
func (s *Server) GetLiveness(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Liveness")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetMetrics is the prometheus endpoint
func (s *Server) GetMetrics(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Metrics")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetEdge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	edgeType, ok := vars["edge-type"]
	if !ok || edgeType == "" {
		returnErr(w, "parsing edgeType from path", nil, http.StatusBadRequest)
		return
	}
	fieldName, ok := vars["field-name"]
	if !ok || fieldName == "" {
		returnErr(w, "parsing fieldName from path", nil, http.StatusBadRequest)
		return
	}
	fieldValue, ok := vars["field-value"]
	if !ok || fieldValue == "" {
		returnErr(w, "parsing fieldValue from path", nil, http.StatusBadRequest)
		return
	}

	obj, err := getCollection(edgeType)
	if err != nil {
		returnErr(w, "Invalid edgeType", err, http.StatusBadRequest)
		return
	}

	q := fmt.Sprintf("FOR e IN %s FILTER e.@field == @val RETURN e", edgeType)
	fields := map[string]interface{}{
		"field": fieldName,
		"val":   fieldValue,
	}

	res, err := s.db.Query(q, fields, obj)
	if err != nil || len(res) == 0 {
		returnErr(w, "failed to get edge", err, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res[0]); err != nil {
		log.Warning(err, res)
	}
}

// HeatbeatCollector is a heartbeat endpoint that all collectors periodically call, to notify us that they are alive
func (s *Server) HeartbeatCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["collector-name"]
	if !ok || name == "" {
		returnErr(w, "parsing collector-name from path", nil, http.StatusBadRequest)
		return
	}

	dbCol := database.Collector{Name: name}
	err := s.db.Read(&dbCol)
	if err != nil {
		returnErr(w, "finding collector", err, http.StatusInternalServerError)
		return
	}

	dbCol.LastHeartbeat = time.Now().Format(time.RFC3339)
	err = s.db.Update(&dbCol)
	if err != nil {
		returnErr(w, "updating heartbeat", err, http.StatusInternalServerError)
		return
	}
	defer log.Infof("Heartbeat collector: %+v", dbCol)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) RemoveAllFields(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	edgeType, ok := vars["edge-type"]
	if !ok || edgeType == "" {
		returnErr(w, "parsing edge-type from path", nil, http.StatusBadRequest)
		return
	}

	fieldName, ok := vars["field-name"]
	if !ok || fieldName == "" {
		returnErr(w, "parsing field-name from path", nil, http.StatusBadRequest)
		return
	}
	defer log.Infof("RemoveAllFields: %v:%v", edgeType, fieldName)

	if err, code := s.removeAllFields(edgeType, fieldName); err != nil {
		returnErr(w, "Failed to remove fields", err, code)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) removeAllFields(edgeType string, fieldName string) (error, int) {
	obj, err := getCollection(edgeType)
	if err != nil {
		return err, http.StatusBadRequest
	}

	q := fmt.Sprintf("FOR e IN %s UPDATE e WITH {@field: null} IN %s OPTIONS { keepNull: false }",
		edgeType, edgeType)
	fields := map[string]interface{}{
		"field": fieldName,
	}

	_, err = s.db.Query(q, fields, obj)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (s *Server) RemoveField(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	edgeType, ok := vars["edge-type"]
	if !ok || edgeType == "" {
		returnErr(w, "parsing edge-type from path", nil, http.StatusBadRequest)
		return
	}

	edgeKey, ok := vars["edge-key"]
	if !ok || edgeKey == "" {
		returnErr(w, "parsing edge-key from path", nil, http.StatusBadRequest)
		return
	}

	fieldName, ok := vars["field-name"]
	if !ok || fieldName == "" {
		returnErr(w, "parsing field-name from path", nil, http.StatusBadRequest)
		return
	}

	defer log.Infof("RemoveField: %v:%v on key: %v", edgeType, fieldName, edgeKey)

	obj, err := getCollection(edgeType)
	if err != nil {
		returnErr(w, "Invalid edgeType", err, http.StatusBadRequest)
		return
	}

	q := fmt.Sprintf("FOR e IN %s FILTER e._key == @key UPDATE e WITH {@field: null} IN %s OPTIONS { keepNull: false }",
		edgeType, edgeType)
	fields := map[string]interface{}{
		"field": fieldName,
		"key":   edgeKey,
	}

	_, err = s.db.Query(q, fields, obj)
	if err != nil {
		returnErr(w, "failed to remove field from edges", err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// UpdateCollector updates the specified collector
func (s *Server) UpdateCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["collector-name"]
	if !ok || name == "" {
		returnErr(w, "parsing collector-name from path", nil, http.StatusBadRequest)
		return
	}
	collector, err := convert.ApiReq2DbCol(r)
	if err != nil {
		returnErr(w, "parsing collector", err, http.StatusBadRequest)
		return
	}
	if collector.Name != name {
		returnErr(w, "name mismath in object and path", nil, http.StatusBadRequest)
		return
	}

	// validate collector edges and types and things exist
	// Fieldname is not overlapping
	err = s.validateCollector(collector)
	if err != nil {
		returnErr(w, "validating collector", err, http.StatusBadRequest)
		return
	}

	collector.LastHeartbeat = time.Now().Format(time.RFC3339)
	err = s.db.Update(&collector)
	if err != nil {
		returnErr(w, "updating document", err, http.StatusInternalServerError)
		return
	}
	defer log.Infof("Updated collector: %+v", collector)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) UpsertField(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	edgeType, ok := vars["edge-type"]
	if !ok || edgeType == "" {
		returnErr(w, "parsing edge-type from path", nil, http.StatusBadRequest)
		return
	}

	fieldName, ok := vars["field-name"]
	if !ok || fieldName == "" {
		returnErr(w, "parsing field-name from path", nil, http.StatusBadRequest)
		return
	}

	if !s.validateUpsert(fieldName, edgeType) {
		returnErr(w, "field-name not registered", nil, http.StatusBadRequest)
		return
	}

	var eScore client.EdgeScore
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&eScore); err != nil {
		returnErr(w, "decoding json error", err, http.StatusBadRequest)
		return
	}

	if eScore.Key == "" && (eScore.To == "" || eScore.From == "") {
		returnErr(w, "empty key value", nil, http.StatusBadRequest)
		return
	}
	defer log.Infof("UpsertField to %v:%v on key: %v to %v", edgeType, fieldName, eScore.Key, eScore.Value)

	obj, err := getCollection(edgeType)
	if err != nil {
		returnErr(w, "Invalid edgeType", err, http.StatusBadRequest)
		return
	}
	var q string
	var fields map[string]interface{}
	if eScore.Key != "" {
		q = fmt.Sprintf("FOR e IN %s FILTER e._key == @key UPDATE e WITH { @field: @val } IN %s",
			edgeType, edgeType)
		fields = map[string]interface{}{
			"key":   eScore.Key,
			"val":   eScore.Value,
			"field": fieldName,
		}
	} else {
		q = fmt.Sprintf("FOR e IN %s FILTER e._to == @to AND e._from == @frm UPDATE e WITH { @field: @val } IN %s",
			edgeType, edgeType)
		fields = map[string]interface{}{
			"to":    eScore.To,
			"frm":   eScore.From,
			"val":   eScore.Value,
			"field": fieldName,
		}
	}

	_, err = s.db.Query(q, fields, obj)
	if err != nil {
		returnErr(w, "failed to update edge", err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func returnErr(w http.ResponseWriter, mess string, err error, status int) {
	if err == database.ErrNotFound {
		http.Error(w, "document not found", http.StatusNotFound)
		return
	}
	http.Error(w, fmt.Sprintf("%s: err: %v", mess, err), status)
	return
}

func (s *Server) validateCollector(dbCol database.Collector) error {
	pEdge := database.PrefixEdge{}
	lEdge := database.LinkEdge{}
	if dbCol.EdgeType != pEdge.GetType() && dbCol.EdgeType != lEdge.GetType() && dbCol.EdgeType != "" {
		return fmt.Errorf("Invalid Edge Type. Recieved: %q. Must be %q or %q", dbCol.EdgeType, pEdge.GetType(), lEdge.GetType())
	}

	if dbCol.FieldName == "" {
		return errors.New("Empty field name is not valid")
	}

	if dbCol.Status != "" {
		return errors.New("Status is a read only field")
	}

	return s.fieldExists(dbCol)
}

func (s *Server) fieldExists(dbCol database.Collector) error {
	var col database.Collector
	dbCols, err := s.db.Query("FOR c in Collectors FILTER c.FieldName == @name RETURN c", map[string]interface{}{"name": dbCol.FieldName}, col)
	if err != nil {
		return errors.New("Failed to fetch collectors")
	}

	// If the field name is someone else not my own update
	if len(dbCols) == 1 && dbCol.Name != dbCols[0].(*database.Collector).Name {
		return fmt.Errorf("Field Name %q already exists", dbCol.FieldName)
	} else if len(dbCols) > 1 {
		return fmt.Errorf("More than 1 field name %q. Collections Corrupted!!", dbCol.FieldName)
	}

	return nil
}

func (s *Server) validateUpsert(fieldName string, edgeType string) bool {
	var col database.Collector
	dbCols, err := s.db.Query("FOR c in Collectors FILTER c.FieldName == @name AND c.EdgeType == @edgeType RETURN c",
		map[string]interface{}{"name": fieldName, "edgeType": edgeType}, col)
	if err != nil {
		return false
	}
	// If the field name is someone else not my own update
	if len(dbCols) == 1 {
		return true
	}
	return false
}

func getCollection(col string) (interface{}, error) {
	pEdge := database.PrefixEdge{}
	lEdge := database.LinkEdge{}
	var obj interface{}
	if col == pEdge.GetType() {
		obj = database.PrefixEdge{}
	} else if col == lEdge.GetType() {
		obj = database.LinkEdge{}
	} else {
		return obj, errors.New("Invalid EdgeType")
	}
	return obj, nil
}
