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
		Prefix{IP: "192.68.0.0/24"},
		Prefix{IP: "192.68.0.1/24"},
		Prefix{IP: "192.68.0.2/24"},
		Prefix{IP: "192.68.0.3/24"},
	}
	for _, p := range ps {
		err := ac.Add(p)
		if err != nil {
			t.Fatalf("Error adding prefix: %v", err)
		}
	}

	rs := []Router{
		Router{Name: "mike", Key: "mikey"},
		Router{Name: "steve", Key: "stekey"},
		Router{Name: "matt", Key: "matt"},
	}
	for _, r := range rs {
		err := ac.Add(r)
		if err != nil {
			t.Fatalf("Error adding router: %v", err)
		}
	}

	s, err := ac.Query("FOR r in Router FILTER r._name == \"steve\" RETURN r._id")
	if err != nil {
		t.Fatalf("Error fetching steve: %v", err)
	}
	fmt.Printf("\n STEVE: %+v\n", s)

	as := []ASEdge{
		ASEdge{From: "Prefix/287", To: "Prefix/283", Name: "287->283"},
	}
	for _, a := range as {
		err := ac.Add(a)
		if err != nil {
			t.Fatalf("Error adding asEdge: %v", err)
		}
	}

	ls := []LinkEdge{
		LinkEdge{From: "Router/mikey", To: "Router/stekey", Name: "mikey->stekey"},
	}
	for _, l := range ls {
		err := ac.Add(l)
		if err != nil {
			t.Fatalf("Error adding lsEdge: %v", err)
		}
	}

	c, err := ac.Query("FOR router, edge in outbound 'Router/mikey' LinkEdge return {router, edge}")
	if err != nil {
		t.Fatalf("Error querying: %v", err)
	}
	fmt.Printf("%v\n", c)

	c, err = ac.Query("FOR router in Router return router")
	if err != nil {
		t.Fatalf("Error querying: %v", err)
	}
	for i, r := range c {
		fmt.Printf("Router: %d. Val: %+v\n", i, r)
	}

}

func cleanUp(ac ArangoConn) {
	err := ac.db.Remove(context.Background())
	if err != nil {
		fmt.Printf("Error destroying us all: %v\n", err)
	}
}
