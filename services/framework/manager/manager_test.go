package manager

import(
  "errors"
  "fmt"
  "reflect"
  "testing"
  "time"

  "github.com/golang/mock/gomock"
  "wwwin-github.cisco.com/spa-ie/voltron/services/framework/database"
  mock_database "wwwin-github.cisco.com/spa-ie/voltron/services/framework/database/mock"
)

func setUp(t *testing.T) (*mock_database.MockDatabase, *Manager) {
	ctrl := gomock.NewController(t)
	dbMock := mock_database.NewMockDatabase(ctrl)
  m, err := NewManager(Config{Interval: "10ms"}, dbMock)
  if err != nil {
    t.Fatal("Failed to create manager")
  }

	return dbMock, m
}

func TestProcessControllers(t *testing.T){

  tests := []struct {
    collectors []interface{}
    updateCol database.Collector
    dbErr error
    expectedErr error
  }{
    // Empty Collector ret
    {
      collectors: []interface{}{},
      dbErr: nil,
      expectedErr: nil,
    },
    // DB error
    {
      collectors: []interface{}{},
      dbErr: errors.New("error"),
      expectedErr: errors.New("error"),
    },
    // Transition to Running with no status
    {
      collectors: []interface{}{
        &database.Collector{
          Timeout: "1s",
          LastHeartbeat: time.Now().Add(1 * time.Second).Format(time.RFC3339),
        },
      },
      updateCol: database.Collector{
        Timeout: "1s",
        Status: StatusRunning,
      },
      dbErr: nil,
      expectedErr: nil,
    },
    // Transition to Down with no status
    {
      collectors: []interface{}{
        &database.Collector{
          Timeout: "1s",
          LastHeartbeat: time.Now().Add(-10 * time.Second).Format(time.RFC3339),
        },
      },
      updateCol: database.Collector{
        Timeout: "1s",
        Status: StatusDown,
      },
      dbErr: nil,
      expectedErr: nil,
    },
  }


  for index, test := range tests {
    dbMock, m := setUp(t)
    setProcessCtrl(dbMock, test.expectedErr, test.collectors)
    if test.updateCol.Timeout != "" {
      dbMock.EXPECT().Update(colMatcher{col: test.updateCol})
    }
    err := m.ProcessControllers()
    if err != test.expectedErr {
      t.Errorf("Test %d: Expected: %v. Received: %v", index, test.expectedErr, err)
    }
  }
}

func setProcessCtrl(dbMock *mock_database.MockDatabase, err error, ret []interface{}){
  q := "FOR c in Collectors return c"
  dbMock.EXPECT().Query(q, nil, database.Collector{}).Return(ret, err)
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
