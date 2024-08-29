package main

import (
	"fmt"
	"homium/crafty"
	"homium/goaccess"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/crafty", crafty.CraftyHandler)
	http.HandleFunc("/goaccess", goaccess.GoaccessHandler)
	port := 4577
	fmt.Printf("Serving on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
