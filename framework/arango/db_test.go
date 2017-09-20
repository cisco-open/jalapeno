package arango

import (
	"context"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	cfg := ArangoConfig{
		URL:      "http://127.0.0.1:8529",
		User:     "root",
		Password: "vojltorb",
		Database: "test",
	}
	ac, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to NewArango: %v", err)
	}
	defer cleanUp(ac)

	ps := []Prefix{
		Prefix{Key: "a"},
		Prefix{Key: "b"},
		Prefix{Key: "c"},
		Prefix{Key: "d"},
	}
	for _, p := range ps {
		err := ac.Add(&p)
		if err != nil {
			t.Fatalf("Error adding prefix: %v", err)
		}
	}

	rs := []Router{
		Router{Key: "mikey", Name: "mike"},
		Router{Key: "stekey", Name: "steve"},
		Router{Key: "stekey2", Name: "stevto"},
		Router{Key: "matkey", Name: "matt"},
	}
	for _, r := range rs {
		err := ac.Add(&r)
		if err != nil {
			t.Fatalf("Error adding router: %v", err)
		}
	}

	as := []ASEdge{
		ASEdge{From: "Prefixes/a", To: "Prefixes/b", Key: "ab"},
	}
	for _, a := range as {
		err := ac.Add(&a)
		if err != nil {
			t.Fatalf("Error adding asEdge: %v", err)
		}
	}

	ls := []LinkEdge{
		LinkEdge{From: "Routers/mikey", To: "Routers/stekey", Key: "ms"},
		LinkEdge{From: "Routers/mikey", To: "Routers/matkey", Key: "mm"},
	}
	for _, l := range ls {
		err := ac.Add(&l)
		if err != nil {
			t.Fatalf("Error adding lsEdge: %v", err)
		}
	}

	// HALP
	r := Router{}

	s, err := ac.Query("FOR r in Routers RETURN r", r)
	if err != nil {
		t.Fatalf("Error fetching steve: %v", err)
	}
	fmt.Printf("\nQ: FOR r in Routers RETURN r\n")
	fmt.Printf("%+v\n", s)

	for _, a := range s {
		fmt.Printf("%T:%+v\n\n", a, a)
	}

	rr := new(string)
	s, err = ac.Query("FOR r in Routers RETURN r._name", rr)
	if err != nil {
		t.Fatalf("Error fetching steve: %v", err)
	}
	fmt.Printf("\nQ: FOR r in Routers RETURN r._name\n")
	fmt.Printf("%+v\n", s)

	for _, a := range s {
		fmt.Printf("%T:%+v\n\n", a, a)
	}

	lTest := LinkEdge{Key: "ms"}
	err = ac.Read(&lTest)
	if err != nil {
		t.Fatalf("Error reading lsEdge: %v", err)
	}
	fmt.Printf("LTEST: %+v\n\n", lTest)

	type BB struct {
		Router Router
		Edge   LinkEdge
	}
	mike := BB{}

	c, err := ac.Query("FOR router, edge in outbound 'Routers/mikey' LinkEdges return {router, edge}", mike)
	if err != nil {
		t.Fatalf("Error querying: %v", err)
	}
	fmt.Printf("\nQ: FOR router, edge in outbound 'Routers/mikey' LinkEdges return {router, edge}\n")
	fmt.Printf("%v\n", c)
	for _, a := range c {
		fmt.Printf("%T:%+v\n\n", a, a)
	}
}

func cleanUp(ac ArangoConn) {
	err := ac.db.Remove(context.Background())
	if err != nil {
		fmt.Printf("Error destroying us all: %v\n", err)
	}
}
