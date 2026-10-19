package main

import (
	"log"
	"net/http"
	"tesis/adaptativo/backend/api"
)

func main() {
	log.Println("API v0.1 disponible en http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", api.New().Handler()))
}
