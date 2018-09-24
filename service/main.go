package main

import (
	"log"
	"net/http"

	"github.com/dgf/goqldom"
	"github.com/graphql-go/handler"
)

var (
	version = "version"
	commit  = "commit"
	date    = "date"
)

func main() {
	schema, err := goqldom.Schema(version + " " + date + " " + commit)
	if err != nil {
		log.Fatalf("failed to create schema, error: %v", err)
	}
	http.Handle("/graphql", handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	}))

	http.Handle("/", http.FileServer(goqldom.Assets))
	http.ListenAndServe(":8080", nil)
}
