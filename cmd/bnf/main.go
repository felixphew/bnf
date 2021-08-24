package main

import (
	"log"
	"net/http"

	"github.com/felixphew/bnf"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", bnf.Handler))
}
