package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/api/v1/client"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/api/v1/convert"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/database"
	mock_database "wwwin-github.cisco.com/spa-ie/voltron-redux/framework/database/mock"
)

func setUp(t *testing.T) (*mock_database.MockDatabase, *client.DefaultApi) {
	ctrl := gomock.NewController(t)
	dbMock := mock_database.NewMockDatabase(ctrl)
	svr, err := New(NewConfig(), dbMock)
	if err != nil {
		t.Fatalf("Could not create server. err: %v", err)
	}

	htest := httptest.NewServer(svr.router)
	client := client.NewDefaultApiWithBasePath(fmt.Sprintf("%v/v1", htest.URL))
	return dbMock, client
}

func TestGetCollectors(t *testing.T) {
	tests := []struct {
		collectors         []interface{}
		dbErr              error
		expectedStatusCode int
	}{
		// empty collectors returned
		{
			collectors:         []interface{}{&database.Collector{}},
			dbErr:              nil,
			expectedStatusCode: http.StatusOK,
		},
		// collectors returned
		{
			collectors: []interface{}{
				&database.Collector{
					Name: "Test1",
				},
				&database.Collector{
					Name: "Test2",
				},
			},
			dbErr:              nil,
			expectedStatusCode: http.StatusOK,
		},
		// db error returned
		{
			collectors: []interface{}{
				&database.Collector{
					Name: "Test1",
				},
				&database.Collector{
					Name: "Test2",
				},
			},
			dbErr:              errors.New("error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		setGetCollector(test.collectors, test.dbErr, dbMock)
		//TODO stop ignoring the error
		cols, resp, _ := client.GetCollectors()
		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
		for i, c := range cols {
			if !reflect.DeepEqual(c, convert.DbCol2ApiCol(*test.collectors[i].(*database.Collector))) {
				t.Errorf("Test %d: Returned objects :%d do not match expected.", index, i)
			}
		}
	}
}

func setGetCollector(cols []interface{}, err error, dbMock *mock_database.MockDatabase) {
	q := "FOR c in Collectors RETURN c"
	dbMock.EXPECT().Query(q, nil, database.Collector{}).Return(cols, err)
}

func TestAddCollectors(t *testing.T) {
	tests := []struct {
		collector          client.Collector
		retCols            []interface{}
		dbErr              error
		readErr            error
		fieldErr           error
		expectedStatusCode int
		findMatch          bool
	}{
		// collector Added
		{
			collector: client.Collector{
				Name:      "Add0",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			readErr:            nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusCreated,
		},
		// collector bad edgeType
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "BAD EDGE",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			readErr:            nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		// collector bad fieldName
		{
			collector: client.Collector{
				Name:      "Add2",
				EdgeType:  "PrefixEdges",
				FieldName: "",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			readErr:            nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			collector: client.Collector{
				Name:      "Add3",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              errors.New("DB Problems"),
			readErr:            nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			collector: client.Collector{
				Name:      "Add4",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			readErr:            nil,
			fieldErr:           errors.New(""),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			collector: client.Collector{
				Name:      "Add5",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			readErr:            nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusAlreadyReported,
			findMatch:          true,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Insert(colMatcher{convert.ApiCol2DbCol(test.collector)}).Return(test.dbErr)
		if test.findMatch {
			dbMock.EXPECT().Read(colMatcher{database.Collector{Name: test.collector.Name}}).Return(test.readErr).SetArg(0, convert.ApiCol2DbCol(test.collector))
		} else {
			dbMock.EXPECT().Read(colMatcher{database.Collector{Name: test.collector.Name}}).Return(test.readErr)
		}
		setFieldExists(test.retCols, test.collector.FieldName, test.fieldErr, dbMock)
		//TODO stop ignoring the error
		resp, _ := client.AddCollector(test.collector)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func setFieldExists(cols []interface{}, fieldName string, err error, dbMock *mock_database.MockDatabase) {
	q := "FOR c in Collectors FILTER c.FieldName == @name RETURN c"
	bind := map[string]interface{}{"name": fieldName}
	setQuery(q, bind, database.Collector{}, cols, err, dbMock)
}

func TestDeleteCollectors(t *testing.T) {
	tests := []struct {
		collector          database.Collector
		readErr            error
		fieldErr           error
		delErr             error
		expectedStatusCode int
	}{
		{
			collector: database.Collector{
				Name:      "test0",
				EdgeType:  "PrefixEdges",
				FieldName: "test0",
			},
			readErr:            nil,
			fieldErr:           nil,
			delErr:             nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			collector: database.Collector{
				Name:      "test1",
				EdgeType:  "PrefixEdges",
				FieldName: "test1",
			},
			readErr:            nil,
			fieldErr:           nil,
			delErr:             errors.New("Not found"),
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			collector: database.Collector{
				Name:      "test2",
				EdgeType:  "PrefixEdges",
				FieldName: "test2",
			},
			readErr:            nil,
			fieldErr:           errors.New("Not found"),
			delErr:             nil,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			collector: database.Collector{
				Name:      "test3",
				EdgeType:  "PrefixEdges",
				FieldName: "test3",
			},
			readErr:            errors.New("Not found"),
			fieldErr:           nil,
			delErr:             nil,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Delete(colMatcher{test.collector}).Return(test.delErr)
		dbMock.EXPECT().Read(colMatcher{database.Collector{Name: test.collector.Name}}).Return(test.readErr).SetArg(0, test.collector)
		setRemoveAll(test.collector.EdgeType, test.collector.FieldName, test.fieldErr, dbMock)
		//TODO stop ignoring the error
		resp, _ := client.DeleteCollector(test.collector.Name)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestGetCollector(t *testing.T) {
	tests := []struct {
		collectorName      string
		dbErr              error
		expectedStatusCode int
	}{
		{
			collectorName:      "test",
			dbErr:              nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			collectorName:      "test",
			dbErr:              errors.New("Not found"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Read(colMatcher{database.Collector{Name: test.collectorName}}).Return(test.dbErr)
		//TODO stop ignoring the error
		_, resp, _ := client.GetCollector(test.collectorName)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestHeartbeatCollector(t *testing.T) {
	tests := []struct {
		collectorName      string
		readErr            error
		upErr              error
		expectedStatusCode int
	}{
		{
			collectorName:      "test",
			readErr:            nil,
			upErr:              nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			collectorName:      "test",
			readErr:            errors.New("Not found"),
			upErr:              nil,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			collectorName:      "test",
			readErr:            nil,
			upErr:              errors.New("error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Read(colMatcher{database.Collector{Name: test.collectorName}}).Return(test.readErr)
		dbMock.EXPECT().Update(colMatcher{database.Collector{Name: test.collectorName}}).Return(test.upErr)
		//TODO stop ignoring the error
		_, resp, _ := client.HeartbeatCollector(test.collectorName)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestUpdateCollectors(t *testing.T) {
	tests := []struct {
		collector          client.Collector
		retCols            []interface{}
		dbErr              error
		fieldErr           error
		expectedStatusCode int
	}{
		// collector Added
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusOK,
		},
		// collector bad edgeType
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "BAD EDGE",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		// collector bad fieldName
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              errors.New("DB Problems"),
			fieldErr:           nil,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
			},
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           errors.New(""),
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Update(colMatcher{convert.ApiCol2DbCol(test.collector)}).Return(test.dbErr)
		setFieldExists(test.retCols, test.collector.FieldName, test.fieldErr, dbMock)
		//TODO stop ignoring the error
		resp, _ := client.UpdateCollector(test.collector.Name, test.collector)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRemoveAllFields(t *testing.T) {
	tests := []struct {
		edgeType           string
		fieldName          string
		dbErr              error
		expectedStatusCode int
	}{
		// good test
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			fieldName:          "test0",
			dbErr:              nil,
			expectedStatusCode: 200,
		},
		// bad edge type
		{
			edgeType:           "badType",
			fieldName:          "test1",
			dbErr:              nil,
			expectedStatusCode: 400,
		},
		// bad edge type
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			fieldName:          "",
			dbErr:              nil,
			expectedStatusCode: 404,
		},
		// empty edge type
		{
			edgeType:           "",
			fieldName:          "test3",
			dbErr:              nil,
			expectedStatusCode: 301,
		},
		// empty field name
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			fieldName:          "",
			dbErr:              nil,
			expectedStatusCode: 404,
		},
		// DB error
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			fieldName:          "test5",
			dbErr:              errors.New("Non nill error"),
			expectedStatusCode: 500,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		setRemoveAll(test.edgeType, test.fieldName, test.dbErr, dbMock)
		//TODO stop ignoring the error
		resp, _ := client.RemoveAllFields(test.edgeType, test.fieldName)
		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func setRemoveAll(edgeType string, fieldName string, err error, dbMock *mock_database.MockDatabase) {
	q := fmt.Sprintf("FOR e IN %s UPDATE e WITH {@field: null} IN %s OPTIONS { keepNull: false }",
		edgeType, edgeType)
	bind := map[string]interface{}{
		"field": fieldName,
	}
	obj, _ := getCollection(edgeType)
	setQuery(q, bind, obj, []interface{}{}, err, dbMock)
}

func TestRemoveField(t *testing.T) {
	tests := []struct {
		edgeType           string
		edgeKey            string
		fieldName          string
		dbErr              error
		expectedStatusCode int
	}{
		// good test
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			edgeKey:            "key0",
			fieldName:          "test0",
			dbErr:              nil,
			expectedStatusCode: 200,
		},
		// bad edge type
		{
			edgeType:           "badType",
			edgeKey:            "key1",
			fieldName:          "test1",
			dbErr:              nil,
			expectedStatusCode: 400,
		},
		// bad edge type
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			edgeKey:            "key2",
			fieldName:          "",
			dbErr:              nil,
			expectedStatusCode: 404,
		},
		// empty edge type
		{
			edgeType:           "",
			edgeKey:            "key3",
			fieldName:          "test3",
			dbErr:              nil,
			expectedStatusCode: 301,
		},
		// empty field name
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			edgeKey:            "key4",
			fieldName:          "",
			dbErr:              nil,
			expectedStatusCode: 404,
		},
		// DB error
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			edgeKey:            "key5",
			fieldName:          "test5",
			dbErr:              errors.New("Non nill error"),
			expectedStatusCode: 500,
		},
		// Empty key
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			edgeKey:            "",
			fieldName:          "test6",
			dbErr:              nil,
			expectedStatusCode: 301,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		setRemove(test.edgeType, test.edgeKey, test.fieldName, test.dbErr, dbMock)
		//TODO stop ignoring the error
		resp, _ := client.RemoveField(test.edgeType, test.edgeKey, test.fieldName)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func setRemove(edgeType string, edgeKey string, fieldName string, err error, dbMock *mock_database.MockDatabase) {
	q := fmt.Sprintf("FOR e IN %s FILTER e._key == @key UPDATE e WITH {@field: null} IN %s OPTIONS { keepNull: false }",
		edgeType, edgeType)
	bind := map[string]interface{}{
		"key":   edgeKey,
		"field": fieldName,
	}
	obj, _ := getCollection(edgeType)

	setQuery(q, bind, obj, []interface{}{}, err, dbMock)
}

func TestUpsertField(t *testing.T) {
	tests := []struct {
		edgeType           string
		fieldName          string
		fieldVal           int32
		key                string
		retCols            []interface{}
		dbErr              error
		fieldErr           error
		expectedStatusCode int
	}{
		// good test with an existing field
		{
			edgeType:  database.PrefixEdge{}.GetType(),
			fieldName: "test0",
			fieldVal:  0,
			key:       "key0",
			retCols: []interface{}{
				database.Collector{
					FieldName: "test0",
				},
			},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: 200,
		},
		// bad edge type
		{
			edgeType:           "badType",
			fieldName:          "test1",
			fieldVal:           1,
			key:                "key1",
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: 400,
		},
		// bad edge type
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			fieldName:          "",
			fieldVal:           2,
			key:                "key2",
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: 404,
		},
		// empty edge type
		{
			edgeType:           "",
			fieldName:          "test3",
			fieldVal:           3,
			key:                "key3",
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: 301,
		},
		// empty field name
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			fieldName:          "",
			fieldVal:           4,
			key:                "key4",
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: 404,
		},
		// DB error
		{
			edgeType:  database.PrefixEdge{}.GetType(),
			fieldName: "test5",
			fieldVal:  5,
			key:       "key5",
			retCols: []interface{}{
				database.Collector{
					FieldName: "test5",
					EdgeType:  database.PrefixEdge{}.GetType(),
				},
			},
			dbErr:              errors.New("Non nill error"),
			fieldErr:           nil,
			expectedStatusCode: 500,
		},
		// Empty key
		{
			edgeType:  database.PrefixEdge{}.GetType(),
			fieldName: "test6",
			fieldVal:  6,
			key:       "",
			retCols: []interface{}{
				database.Collector{
					FieldName: "test6",
					EdgeType:  database.PrefixEdge{}.GetType(),
				},
			},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: 400,
		},
		// Field err
		{
			edgeType:  database.PrefixEdge{}.GetType(),
			fieldName: "test7",
			fieldVal:  7,
			key:       "key7",
			retCols: []interface{}{
				database.Collector{
					FieldName: "test7",
					EdgeType:  database.PrefixEdge{}.GetType(),
				},
			},
			dbErr:              nil,
			fieldErr:           errors.New("Non nill error"),
			expectedStatusCode: 400,
		},
		// Field does not exist
		{
			edgeType:           database.PrefixEdge{}.GetType(),
			fieldName:          "test8",
			fieldVal:           8,
			key:                "key8",
			retCols:            []interface{}{},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: 400,
		},
	}

	for index, test := range tests {
		dbMock, clientConn := setUp(t)
		setUpsert(test.edgeType, test.key, test.fieldVal, test.fieldName, test.dbErr, dbMock)
		setValidateUpsert(test.retCols, test.fieldName, test.edgeType, test.fieldErr, dbMock)

		//TODO stop ignoring the error
		resp, _ := clientConn.UpsertField(test.edgeType, test.fieldName,
			client.EdgeScore{
				Key:   test.key,
				Value: test.fieldVal,
			})

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func setUpsert(edgeType string, edgeKey string, fieldVal int32, fieldName string, err error, dbMock *mock_database.MockDatabase) {
	q := fmt.Sprintf("FOR e IN %s FILTER e._key == @key UPDATE e WITH { @field: @val } IN %s",
		edgeType, edgeType)
	bind := map[string]interface{}{
		"key":   edgeKey,
		"val":   fieldVal,
		"field": fieldName,
	}
	obj, _ := getCollection(edgeType)
	setQuery(q, bind, obj, []interface{}{}, err, dbMock)
}

func setValidateUpsert(cols []interface{}, fieldName string, edgeType string, err error, dbMock *mock_database.MockDatabase) {
	q := "FOR c in Collectors FILTER c.FieldName == @name AND c.EdgeType == @edgeType RETURN c"
	bind := map[string]interface{}{"name": fieldName, "edgeType": edgeType}
	setQuery(q, bind, database.Collector{}, cols, err, dbMock)
}

func setQuery(q string, bind map[string]interface{}, obj interface{}, retObj []interface{}, err error, dbMock *mock_database.MockDatabase) {
	dbMock.EXPECT().Query(q, bind, obj).Return(retObj, err)
}

type colMatcher struct {
	col database.Collector
}

func (cm colMatcher) Matches(x interface{}) bool {
	if col, ok := x.(database.Collector); ok {
		return CollectorsEqual(cm.col, col)
	}
	if col, ok := x.(*database.Collector); ok {
		return CollectorsEqual(cm.col, *col)
	}
	return false
}

func (cm colMatcher) String() string {
	return fmt.Sprintf("collector :%v", cm.col)
}

func CollectorsEqual(c1 database.Collector, c2 database.Collector) bool {
	//make times equal
	//make keys equal (not set by caller)
	c1.Key, _ = c1.GetKey()
	c2.Key, _ = c2.GetKey()
	t := time.Now().Format(time.RFC3339)
	c1.LastHeartbeat = t
	c2.LastHeartbeat = t
	return reflect.DeepEqual(c1, c2)
}
