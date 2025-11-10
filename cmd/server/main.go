package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	contactapp "github.com/rezbow/contact-app"
)

func main() {
	store := contactapp.NewinMemoryStore()
	server := contactapp.NewContactServer(store)
	http.DefaultServeMux.Handle("/", server)
	log.Println(http.ListenAndServe(":8080", http.DefaultServeMux))
}
