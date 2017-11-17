package database

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

var (
	goodCfg = ArangoConfig{
		URL:      "http://127.0.0.1:8529",
		User:     "root",
		Password: "voltron",
		Database: "testDB",
	}
	collections = []string{prefixName, routerName, asName, linkEdgeNamev4, linkEdgeNamev6}
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg.Database != "" || cfg.Password != "" || cfg.URL != "" || cfg.User != "" {
		t.Errorf("NewConfig() did not return expected empty default: %+v", cfg)
	}
}

// Also tests ensure calls
func TestNew(t *testing.T) {
	tests := []struct {
		cfg     ArangoConfig
		err     string
		cleanUp bool
	}{
		//empty
		{
			NewConfig(),
			"ArangoDB Config has an empty field",
			true,
		},
		// Bad user
		{
			ArangoConfig{
				URL:      "http://127.0.0.1:8529",
				User:     "pen-pineapple-apple-pen",
				Password: "what",
				Database: "testNew",
			},
			"ArangoError: Code 401, ErrorNum 0",
			true,
		},
		// Bad password
		{
			ArangoConfig{
				URL:      "http://127.0.0.1:8529",
				User:     "root",
				Password: "pen-pineapple-apple-pen",
				Database: "testNew",
			},
			"ArangoError: Code 401, ErrorNum 0",
			true,
		},
		// Bad URL
		{
			ArangoConfig{
				URL:      "http://garbageMcGarbageFace",
				User:     "root",
				Password: "pen-pineapple-apple-pen",
				Database: "testNew",
			},
			"Get http://garbageMcGarbageFace/_db/testNew/_api/database/current: dial tcp: lookup garbageMcGarbageFace: no such host",
			true,
		},
		// Good
		{
			ArangoConfig{
				URL:      "http://127.0.0.1:8529",
				User:     "root",
				Password: "voltron",
				Database: "test",
			},
			"",
			false,
		},
		//Test existing
		{
			ArangoConfig{
				URL:      "http://127.0.0.1:8529",
				User:     "root",
				Password: "voltron",
				Database: "test",
			},
			"",
			true,
		},
	}

	for i, test := range tests {
		conn, err := NewArango(test.cfg)

		if err != nil {
			if err.Error() != test.err {
				t.Errorf("New Test %d: Failed, Expected: %q Received: %q", i, test.err, err.Error())
			}
		} else {
			// equivalent of nil for now
			if test.err != "" {
				t.Errorf("New Test %d: Failed, Expected %q, but we saw success", i, test.err)
			}
		}

		// Things work, lets make sure
		if err == nil {
			if test.cleanUp {
				defer cleanUp(conn)
			}

			for _, col := range collections {
				b, err := conn.db.CollectionExists(context.Background(), col)
				if err != nil || !b {
					t.Errorf("Could not find the collection %q", col)
				}

				b, err = conn.db.GraphExists(context.Background(), graphName)
				if err != nil || !b {
					t.Errorf("Could not find the graph %q", graphName)
				}
			}
		}
	}
}

func TestFindCollection(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	tests := []struct {
		colName string
		err     error
	}{
		{
			"",
			ErrCollectionNotFound,
		},
		{
			prefixName,
			nil,
		},
		{
			routerName,
			nil,
		},
		{
			linkEdgeNamev4,
			nil,
		},
		{
			linkEdgeNamev6,
			nil,
		},
		{
			asName,
			nil,
		},
	}

	for i, test := range tests {
		_, err := conn.findCollection(test.colName)
		if err != test.err {
			t.Errorf("FindCollection Test %d: Failed. Expected: %v. Received: %v", i, test.err, err)
		}
	}
}

func TestInsert(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	tests := []struct {
		obj DBObject
		err string
	}{
		// Good
		{
			&Router{
				BGPID: "test",
				ASN:   "1",
			},
			"",
		},
		//Empty
		{
			&Router{},
			ErrKeyInvalid.Error(),
		},
		//Duplicate
		{
			&Router{
				BGPID: "test",
				ASN:   "1",
			},
			"unique constraint violated - in index 0 of type primary over [\"_key\"]",
		},
		// Key changed
		{
			&Router{
				Key:   "change",
				BGPID: "test",
				ASN:   "1",
			},
			ErrKeyChange.Error(),
		},
	}

	for i, test := range tests {
		err := conn.Insert(test.obj)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("Insert Test %d: Failed, Expected: %q Received: %q", i, test.err, err.Error())
			}
		} else {
			// equivalent of nil for now
			if test.err != "" {
				t.Errorf("Insert Test %d: Failed, Expected %q, but we saw success", i, test.err)
			}
		}
	}
}

func TestRead(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	tests := []struct {
		obj      DBObject
		fetchKey string
		err      string
	}{
		// Valid
		{
			&Router{
				BGPID: "test0",
			},
			"test0",
			"",
		},
		//No Key
		{
			&Router{
				BGPID: "test1",
			},
			"",
			ErrKeyInvalid.Error(),
		},
		//Not found key
		{
			&Router{
				BGPID: "test3",
			},
			"garbage",
			ErrNotFound.Error(),
		},
	}

	for i, test := range tests {
		err := conn.Insert(test.obj)
		if err != nil {
			t.Fatalf("Read Test %d: Failed on insert: %v", i, err)
		}

		retObj := &Router{Key: test.fetchKey}
		err = conn.Read(retObj)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("Read Test %d: Failed, Expected: %q Received: %q", i, test.err, err.Error())
			}
		} else {
			// equivalent of nil for now
			if test.err != "" {
				t.Errorf("Read Test %d: Failed, Expected %q, but we saw success", i, test.err)
			} else {
				if !reflect.DeepEqual(retObj, test.obj.(*Router)) {
					t.Errorf("Read Test %d: Failed on match. \nOriginal: %+v\nReturned: %+v\n", i, test.obj, retObj)
				}
			}
		}
	}
}

func TestDelete(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	tests := []struct {
		obj      DBObject
		fetchKey string
		delKey   string
		err      string
	}{
		// Valid
		{
			&Router{
				BGPID: "test0",
			},
			"test0",
			"test0",
			"",
		},
		//No Key to delete
		{
			&Router{
				BGPID: "test1",
			},
			"test1",
			"",
			ErrKeyInvalid.Error(),
		},
		//Delete a key not found
		{
			&Router{
				BGPID: "test2",
			},
			"test2",
			"garbage",
			ErrNotFound.Error(),
		},
	}

	for i, test := range tests {
		err := conn.Insert(test.obj)
		if err != nil {
			t.Fatalf("Delete Test %d: Failed on insert: %v", i, err)
		}

		retObj := &Router{Key: test.fetchKey}
		err = conn.Read(retObj)
		if err != nil {
			t.Fatalf("Delete Test %d: Failed on initial read: %v", i, err)
		}

		delObj := &Router{Key: test.delKey}
		err = conn.Delete(delObj)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("Delete Test %d: Failed, Expected: %q Received: %q", i, test.err, err.Error())
			}
		} else {
			// equivalent of nil for now
			if test.err != "" {
				t.Errorf("Delete Test %d: Failed, Expected %q, but we saw success", i, test.err)
			}
		}
		ret2Obj := &Router{Key: test.fetchKey}

		err = conn.Read(ret2Obj)
		if err != nil && err.Error() != ErrNotFound.Error() {
			t.Errorf("Delete Test %d: Failed, Expected document not found. Err is: %v", i, err)
		}
	}
}

func TestUpdate(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	tests := []struct {
		obj        DBObject
		updatedObj DBObject
		fetchKey   string
		err        string
	}{
		// Valid
		{
			&Router{
				BGPID: "test0",
				Name:  "original",
			},
			&Router{
				BGPID: "test0",
				Name:  "updated",
			},
			"test0",
			"",
		},
		//No Key to update
		{
			&Router{
				BGPID: "test2",
			},
			&Router{},
			"test2",
			ErrKeyInvalid.Error(),
		},
		//New document on update
		{
			&Router{
				BGPID: "test3",
			},
			&Router{
				BGPID: "test4",
			},
			"test3",
			ErrNotFound.Error(),
		},
		// Key change
		{
			&Router{
				BGPID: "test5",
			},
			&Router{
				Key:   "change",
				BGPID: "testChange",
			},
			"test5",
			ErrKeyChange.Error(),
		},
	}

	for i, test := range tests {
		err := conn.Insert(test.obj)
		if err != nil {
			t.Fatalf("Update Test %d: Failed on insert: %v", i, err)
		}

		retObj := &Router{Key: test.fetchKey}
		err = conn.Read(retObj)
		if err != nil {
			t.Errorf("Update Test %d: Failed to fetch original object %v", i, err)
		}
		if !reflect.DeepEqual(retObj, test.obj.(*Router)) {
			t.Errorf("Update Test %d: Failed on match. \nOriginal: %+v\nReturned: %+v\n", i, test.obj, retObj)
		}

		err = conn.Update(test.updatedObj)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("Update Test %d: Failed, Expected: %q Received: %q", i, test.err, err.Error())
			}
		} else {
			// equivalent of nil for now
			if test.err != "" {
				t.Errorf("Update Test %d: Failed, Expected %q, but we saw success", i, test.err)
			}

			ret2Obj := &Router{Key: test.fetchKey}
			err = conn.Read(ret2Obj)
			if err != nil {
				t.Errorf("Update Test %d: Failed to fetch updated object %v", i, err)
			}
			if !reflect.DeepEqual(ret2Obj, test.updatedObj.(*Router)) {
				t.Errorf("Update Test %d: Failed on updated match. \nOriginal: %+v\nReturned: %+v\n", i, test.updatedObj, ret2Obj)
			}
		}
	}
}

func TestUpsert(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	tests := []struct {
		obj       DBObject
		upsertObj DBObject
		fetchKey  string
		fetch2Key string
		err       string
	}{
		// Valid
		{
			&Router{
				BGPID: "test0",
				Name:  "original",
			},
			&Router{
				BGPID: "test0",
				Name:  "updated",
			},
			"test0",
			"test0",
			"",
		},
		//No Key to update
		{
			&Router{
				BGPID: "test2",
			},
			&Router{},
			"test2",
			"test2",
			ErrKeyInvalid.Error(),
		},
		//New document on update
		{
			&Router{
				BGPID: "test3",
			},
			&Router{
				BGPID: "test4",
			},
			"test3",
			"test4",
			"",
		},
		// Key change
		{
			&Router{
				BGPID: "test5",
			},
			&Router{
				Key:   "change",
				BGPID: "test5",
			},
			"test5",
			"test6",
			ErrKeyChange.Error(),
		},
	}

	for i, test := range tests {
		err := conn.Insert(test.obj)
		if err != nil {
			t.Fatalf("Upsert Test %d: Failed on insert: %v", i, err)
		}

		retObj := &Router{Key: test.fetchKey}
		err = conn.Read(retObj)
		if err != nil {
			t.Errorf("Upsert Test %d: Failed to fetch original object %v", i, err)
		}
		if !reflect.DeepEqual(retObj, test.obj.(*Router)) {
			t.Errorf("Upsert Test %d: Failed on match. \nOriginal: %+v\nReturned: %+v\n", i, test.obj, retObj)
		}

		err = conn.Upsert(test.upsertObj)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("Upsert Test %d: Failed, Expected: %q Received: %q", i, test.err, err.Error())
			}
		} else {
			// equivalent of nil for now
			if test.err != "" {
				t.Errorf("Upsert Test %d: Failed, Expected %q, but we saw success", i, test.err)
			}

			ret2Obj := &Router{Key: test.fetch2Key}
			err = conn.Read(ret2Obj)
			if err != nil {
				t.Errorf("Upsert Test %d: Failed to fetch updated object %v", i, err)
			}
			if !reflect.DeepEqual(ret2Obj, test.upsertObj.(*Router)) {
				t.Errorf("Upsert Test %d: Failed on updated match. \nOriginal: %+v\nReturned: %+v\n", i, test.upsertObj, ret2Obj)
			}
		}
	}
}

func TestUpsertSafe(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	tests := []struct {
		obj         DBObject
		upsertObj   DBObject
		expectedObj DBObject
		fetchKey    string
		fetch2Key   string
		err         string
	}{
		// Valid keeping an old field
		{
			&Router{
				BGPID: "test1",
				Name:  "original",
			},
			&Router{
				BGPID:    "test1",
				RouterIP: "haha",
			},
			&Router{
				BGPID:    "test1",
				RouterIP: "haha",
				Name:     "original",
			},
			"test1",
			"test1",
			"",
		},
		// Valid keeping an old field
		{
			&Router{
				BGPID: "test11",
				Name:  "original",
			},
			&Router{
				BGPID: "test11",
				Name:  "updated",
			},
			&Router{
				BGPID: "test11",
				Name:  "updated",
			},
			"test11",
			"test11",
			"",
		},
		//No Key to update
		{
			&Router{
				BGPID: "test2",
			},
			&Router{},
			&Router{},
			"test2",
			"test2",
			ErrKeyInvalid.Error(),
		},
		//New document on update
		{
			&Router{
				BGPID: "test3",
			},
			&Router{
				BGPID: "test4",
			},
			&Router{
				BGPID: "test4",
			},
			"test3",
			"test4",
			"",
		},
		// Key change
		{
			&Router{
				BGPID: "test5",
			},
			&Router{
				Key:   "change",
				BGPID: "test6",
			},
			&Router{},
			"test5",
			"test6",
			ErrKeyChange.Error(),
		},
	}

	for i, test := range tests {
		err := conn.Insert(test.obj)
		if err != nil {
			t.Fatalf("UpsertSafe Test %d: Failed on insert: %v", i, err)
		}

		retObj := &Router{Key: test.fetchKey}
		err = conn.Read(retObj)
		if err != nil {
			t.Errorf("UpsertSafe Test %d: Failed to fetch original object %v", i, err)
		}
		if !reflect.DeepEqual(retObj, test.obj.(*Router)) {
			t.Errorf("UpsertSafe Test %d: Failed on match. \nOriginal: %+v\nReturned: %+v\n", i, test.obj, retObj)
		}

		err = conn.UpsertSafe(test.upsertObj)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("UpsertSafe Test %d: Failed, Expected: %q Received: %q", i, test.err, err.Error())
			}
		} else {
			// equivalent of nil for now
			if test.err != "" {
				t.Errorf("UpsertSafe Test %d: Failed, Expected %q, but we saw success", i, test.err)
			}

			ret2Obj := &Router{Key: test.fetch2Key}
			err = conn.Read(ret2Obj)
			if err != nil {
				t.Errorf("UpsertSafe Test %d: Failed to fetch updated object %v", i, err)
			}
			if !reflect.DeepEqual(ret2Obj, test.upsertObj.(*Router)) {
				t.Errorf("UpsertSafe Test %d: Failed on updated match. \nOriginal: %+v\nReturned: %+v\n", i, test.upsertObj, ret2Obj)
			}
		}
	}
}

func TestQuery(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}

	rs := []*Router{
		&Router{BGPID: "test1", ASN: "1"},
		&Router{BGPID: "test2", ASN: "2"},
		&Router{BGPID: "test3", ASN: "3"},
	}

	for _, r := range rs {
		err := conn.Insert(r)
		if err != nil {
			t.Fatalf("Error Inserting router: %+v. Err: %v", r, err)
		}
	}

	ls := []*LinkEdge{
		&LinkEdge{From: "Routers/test1", To: "Routers/test2", FromIP: "1.1.1.1", ToIP: "2.2.2.2"},
		&LinkEdge{From: "Routers/test1", To: "Routers/test3", FromIP: "1.1.1.1", ToIP: "3.3.3.3"},
	}

	for _, l := range ls {
		err := conn.Insert(l)
		if err != nil {
			t.Fatalf("Error Inserting Link Edge: %v", err)
		}
	}

	// Test a single type return
	r := &Router{}
	s, err := conn.Query("FOR r in Routers RETURN r", nil, r)
	if err != nil {
		t.Fatalf("Error fetching all routers: %v", err)
	}

	if !exactMatch(rs, s) {
		t.Errorf("Test Query Failed: All router query output did not match expected.")
	}

	// Test string returns
	asnString := new(string)
	expectedAsn := make([]string, 0)
	for _, i := range rs {
		expectedAsn = append(expectedAsn, i.ASN)
	}

	asns, err := conn.Query("FOR r in Routers RETURN r.ASN", nil, asnString)
	if err != nil {
		t.Fatalf("Error fetching router._asn: %v", err)
	}

	if !exactMatch(expectedAsn, asns) {
		t.Errorf("Test Query Failed: All router.ASNs did not match expected")
	}

	// Test string returns with bind Vars
	asnString2 := new(string)
	expectedAsn2 := make([]string, 0)
	expectedAsn2 = append(expectedAsn2, "2")

	asns2, err := conn.Query("FOR r in Routers FILTER r.ASN == @asn RETURN r.ASN", map[string]interface{}{"asn": "2"}, asnString2)
	if err != nil {
		t.Fatalf("Error fetching router.ASN == 2: %v", err)
	}

	if !exactMatch(expectedAsn2, asns2) {
		t.Errorf("Test Query Failed: ASM == 2 router did not match expected")
	}

	// Test mixed object return
	type Mixed struct {
		Router *Router
		Edge   *LinkEdge
	}
	mixed := Mixed{}
	expectedRouters := []*Router{rs[1], rs[2]}
	expectedLinks := []*LinkEdge{ls[0], ls[1]}
	mixedOutput, err := conn.Query("FOR router, edge in outbound 'Routers/test1' LinkEdges return {router, edge}", nil, mixed)
	if err != nil {
		t.Fatalf("Error querying: %v", err)
	}

	var rOutputs []interface{}
	var lOutputs []interface{}
	for _, obj := range mixedOutput {
		actual := obj.(*Mixed)
		rOutputs = append(rOutputs, actual.Router)
		lOutputs = append(lOutputs, actual.Edge)
	}

	if !exactMatch(expectedRouters, rOutputs) || !exactMatch(expectedLinks, lOutputs) {
		t.Errorf("Test Query Failed: Outbound Routers & Links from test_1 did not match expected")
	}

	// Empty Query return
	emptyString := new(string)
	empty, err := conn.Query("FOR r in Routers FILTER r.ASN == \"banana\" RETURN r.ASN", nil, emptyString)
	if err != nil {
		t.Fatalf("Error fetching steve: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("Test Query Failed: Expected an empty return")
	}

	// Bad Query return
	badQuery := new(string)
	_, err = conn.Query("FOR r in BAD QUERY RETURN pineapple-pen", nil, badQuery)
	if err == nil {
		t.Fatalf("Test Query: Failed, expected invalid query, but saw success")
	}
}

func TestQueryOnObject(t *testing.T) {
	conn, err := NewArango(goodCfg)
	defer cleanUp(conn)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}

	rs := []*Router{
		&Router{BGPID: "test1", ASN: "1", RouterIP: "127.0.0.1"},
		&Router{BGPID: "test2", ASN: "2", RouterIP: "127.0.0.2"},
		&Router{BGPID: "test3", ASN: "3", RouterIP: "127.0.0.3"},
	}

	for _, r := range rs {
		err := conn.Insert(r)
		if err != nil {
			t.Fatalf("Error Inserting router: %+v. Err: %v", r, err)
		}
	}

	ls := []*LinkEdge{
		&LinkEdge{From: "Routers/test1", To: "Routers/test2", FromIP: "1.1.1.1", ToIP: "2.2.2.2"},
		&LinkEdge{From: "Routers/test1", To: "Routers/test3", FromIP: "1.1.1.1", ToIP: "3.3.3.3"},
		&LinkEdge{From: "Routers/test2", To: "Routers/test3", FromIP: "2.2.2.2", ToIP: "3.3.3.3"},
	}

	for _, l := range ls {
		err := conn.Insert(l)
		if err != nil {
			t.Fatalf("Error Inserting Link Edge: %v", err)
		}
	}

	// Test with a router that returns one object
	queryRouter := &Router{ASN: "2"}
	r2 := &Router{}
	expectedRouters := []*Router{rs[1]}

	routerList, err := conn.QueryOnObject(queryRouter, r2, map[string]string{})
	if err != nil {
		t.Errorf("Failed to Query on Object Router")
	}
	if !exactMatch(expectedRouters, routerList) {
		t.Errorf("Test QueryOnObj Router failed. Expected: %v. Recieved: %v", expectedRouters, routerList)
	}

	// Test with a router and operator to return 2 objects
	queryRouter2 := &Router{ASN: "2"}
	operators := make(map[string]string)
	operators["ASN"] = "!="
	r3 := &Router{}
	expectedRouters2 := []*Router{rs[0], rs[2]}

	routerList2, err := conn.QueryOnObject(queryRouter2, r3, operators)
	if err != nil {
		t.Errorf("Failed to Query on Object Router with operators")
	}
	if !exactMatch(expectedRouters2, routerList2) {
		t.Errorf("Test QueryOnObj Router with operators failed. Expected: %v. Recieved: %v", expectedRouters2, routerList2)
	}

	// Bad operator
	operators["ASN"] = "??"
	_, err = conn.QueryOnObject(queryRouter2, r3, operators)
	if err.Error() != "syntax error, unexpected ? near '? @ASN RETURN i' at position 1:32" {
		t.Errorf("Test QueryOnObj with bad operator. Expected syntax error: Received: %q", err.Error())
	}

	//Test with a link that returns 2 links
	queryLink := &LinkEdge{FromIP: "1.1.1.1"}
	l2 := &LinkEdge{}
	expectedLinks := []*LinkEdge{ls[0], ls[1]}
	linkList, err := conn.QueryOnObject(queryLink, l2, map[string]string{})
	if err != nil {
		t.Errorf("Failed to Query on Object Link")
	}
	if !exactMatch(expectedLinks, linkList) {
		t.Errorf("Test QueryOnObj Link failed. Expected: %v. Recieved: %v", expectedLinks, linkList)
	}

	//Test with empty query parameters, should return all.
	queryLink2 := &LinkEdge{}
	l3 := &LinkEdge{}
	expectedLinks2 := []*LinkEdge{ls[0], ls[1], ls[2]}
	linkList2, err := conn.QueryOnObject(queryLink2, l3, map[string]string{})
	if err != nil {
		t.Errorf("Failed to Query on Object Link2")
	}
	if !exactMatch(expectedLinks2, linkList2) {
		t.Errorf("Test QueryOnObj Link2 failed. Expected: %v. Recieved: %v", expectedLinks2, linkList2)
	}

}

func exactMatch(slice interface{}, bList []interface{}) bool {
	aList := reflect.ValueOf(slice)
	aMatch := make([]bool, aList.Len())
	bMatch := make([]bool, len(bList))
	for aIndex := 0; aIndex < aList.Len(); aIndex++ {
		for bIndex, bItem := range bList {
			if reflect.DeepEqual(aList.Index(aIndex).Interface(), bItem) {
				aMatch[aIndex] = true
				bMatch[bIndex] = true
				break
			}
		}
	}

	for _, aBool := range aMatch {
		if !aBool {
			return false
		}
	}
	for _, bBool := range bMatch {
		if !bBool {
			return false
		}
	}
	return true
}

func cleanUp(ac ArangoConn) {
	err := ac.db.Remove(context.Background())
	if err != nil {
		fmt.Printf("Error destroying us all: %v\n", err)
	}
}
