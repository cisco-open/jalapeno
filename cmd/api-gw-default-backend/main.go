package main

import (
	//"encoding/json"
	"fmt"
	"net/http"

	"goji.io"
	"goji.io/pat"
)

/*
type Hello struct {
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := pat.Param(r, "name")
	fmt.Fprintf(w, "Hello, %s!", name)
}

func apibase(w http.ResponseWriter, r *http.Request) {
	h := &Hello{Name: "Foo", Msg: "Bar"}
	encoder := json.NewEncoder(w)
	encoder.Encode(h)
}
*/

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "w00t")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	//name := pat.Param(r, "name")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "<html><head><title>NOPE</title></head><body><h2>NOPE</h2></body></html>")
}

func main() {
	mux := goji.NewMux()
	//mux.HandleFunc(pat.Get("/hello/:name"), hello)
	mux.HandleFunc(pat.New("/*"), notFound)
	mux.HandleFunc(pat.Get("/healthz"), healthzHandler)

	fmt.Println("Running")
	http.ListenAndServe(":8000", mux)
}
