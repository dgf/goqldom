package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/dgf/goqldom"
	"github.com/graphql-go/handler"
	"github.com/pkg/browser"
)

var (
	addr    string
	version = "version"
	commit  = "commit"
	date    = "date"
)

func init() {
	flag.StringVar(&addr, "addr", ":0", "TCP address to listen on")
}

func main() {
	flag.Parse()

	log.Println("Starting goqldom service...")
	log.Println("create GraphQL schema instance")
	schema, err := goqldom.Schema(version + " " + date + " " + commit)
	if err != nil {
		log.Fatalf("failed to create schema, error: %v", err)
	}

	log.Println("register GraphQL handle /graphql")
	http.Handle("/graphql", handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	}))

	log.Println("register static assets")
	http.Handle("/", http.FileServer(goqldom.Assets))

	log.Println("listen TCP on: " + addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	url := "http://" + listener.Addr().(*net.TCPAddr).String()
	log.Printf("Running on: " + url)
	browser.OpenURL(fmt.Sprintf(url));
	if err := http.Serve(listener, nil); err != nil {
		log.Fatalf("something went wrong: %s", err.Error())
	}
}
