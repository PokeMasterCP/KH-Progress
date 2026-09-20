package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	port = ":8080"
)

//go:embed index.html
var indexHTML embed.FS

func main() {
	apiKey, steamId, err := validateEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	achievements, err := GetAchievements(apiKey, steamId)
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range achievements {
		fmt.Printf("Name: %s, Achieved: %d, Icon: %s\n", a.Name, a.Achieved, a.icon)
	}

	fmt.Printf("server listening on %s", port)

	s := http.NewServeMux()
	s.Handle("/", http.FileServer(http.FS(indexHTML)))

	if err := http.ListenAndServe(port, s); err != nil {
		log.Fatal(err)
	}
}

func validateEnvVars() (string, string, error) {
	API_KEY := os.Getenv("API_KEY")
	if API_KEY == "" {
		return "", "", fmt.Errorf("API_KEY not set")
	}

	STEAM_ID := os.Getenv("STEAM_ID")
	if STEAM_ID == "" {
		return "", "", fmt.Errorf("STEAM_ID not set")
	}

	return API_KEY, STEAM_ID, nil
}
