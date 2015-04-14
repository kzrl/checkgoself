package main

import (
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {

	//Print the request details to STDOUT
	log.Printf("+%v", r)
}

func main() {
	log.Println("Listening on port", 4242)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":4242", nil))
}
