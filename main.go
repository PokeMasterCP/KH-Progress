package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
)

const (
	port = ":8080"
)

//go:embed index.html
var indexHTML embed.FS

func main() {
	fmt.Printf("server listening on %s", port)

	s := http.NewServeMux()
	s.Handle("/", http.FileServer(http.FS(indexHTML)))

	if err := http.ListenAndServe(port, s); err != nil {
		log.Fatal(err)
	}
}
