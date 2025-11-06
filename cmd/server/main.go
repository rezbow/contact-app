package main

import (
	"log"
	"net/http"

	contactapp "github.com/rezbow/contact-app"
)

func main() {
	store := contactapp.NewinMemoryStore()
	server := contactapp.NewContactServer(store)
	log.Println(http.ListenAndServe(":8080", server))
}
