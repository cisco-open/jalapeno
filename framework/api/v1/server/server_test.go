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

	q := "FOR c in Collectors RETURN c"
	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Query(q, nil, database.Collector{}).Return(test.collectors, test.dbErr)
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

func TestAddCollectors(t *testing.T) {
	tests := []struct {
		collector          client.Collector
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
				FieldType: "integer",
			},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusCreated,
		},
		// collector bad edgeType
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "BAD EDGE",
				FieldName: "Field1",
				FieldType: "integer",
			},
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
				FieldType: "integer",
			},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		// collector bad FieldType
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
				FieldType: "BAD TYPE",
			},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
				FieldType: "integer",
			},
			dbErr:              errors.New("DB Problems"),
			fieldErr:           nil,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
				FieldType: "integer",
			},
			dbErr:              nil,
			fieldErr:           errors.New(""),
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Insert(colMatcher{convert.ApiCol2DbCol(test.collector)}).Return(test.dbErr)
		setFieldExists(test.collector, dbMock, test.fieldErr)
		//TODO stop ignoring the error
		resp, _ := client.AddCollector(test.collector)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestDeleteCollectors(t *testing.T) {
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
		dbMock.EXPECT().Delete(colMatcher{database.Collector{Name: test.collectorName}}).Return(test.dbErr)
		//TODO stop ignoring the error
		resp, _ := client.DeleteCollector(test.collectorName)

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
				FieldType: "integer",
			},
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
				FieldType: "integer",
			},
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
				FieldType: "integer",
			},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		// collector bad FieldType
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
				FieldType: "BAD TYPE",
			},
			dbErr:              nil,
			fieldErr:           nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
				FieldType: "integer",
			},
			dbErr:              errors.New("DB Problems"),
			fieldErr:           nil,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			collector: client.Collector{
				Name:      "Add1",
				EdgeType:  "PrefixEdges",
				FieldName: "Field1",
				FieldType: "integer",
			},
			dbErr:              nil,
			fieldErr:           errors.New(""),
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for index, test := range tests {
		dbMock, client := setUp(t)
		dbMock.EXPECT().Update(colMatcher{convert.ApiCol2DbCol(test.collector)}).Return(test.dbErr)
		setFieldExists(test.collector, dbMock, test.fieldErr)
		//TODO stop ignoring the error
		resp, _ := client.UpdateCollector(test.collector.Name, test.collector)

		if resp != nil && resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test %d: \n\tExpected: %v \n\tReceived: %v", index, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func setFieldExists(col client.Collector, dbMock *mock_database.MockDatabase, err error) {
	q := "FOR c in Collectors FILTER c.FieldName == @name RETURN c"
	// returning the DB object isn't important, returning an error will indicate a match was found
	dbMock.EXPECT().Query(q, map[string]interface{}{"name": col.FieldName}, database.Collector{}).Return([]interface{}{}, err)
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
