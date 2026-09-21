package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
)

const (
	port = ":8080"
)

//go:embed index.html
var indexHTML embed.FS

type pageData struct {
	Achievements []achievement
	Completed    int
	Percentage   int
}

type steamAPI struct {
	apiKey  string
	steamId string
}

func main() {
	apiKey, steamId, err := validateEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	api := &steamAPI{apiKey: apiKey, steamId: steamId}

	log.Printf("server listening on %s", port)

	s := newHandler(api)

	if err := http.ListenAndServe(port, s); err != nil {
		log.Fatal(err)
	}
}

// achievementSource lets handler tests supply fake data without calling Steam.
// The running application uses steamAPI to implement it.
type achievementSource interface {
	GetAchievements() ([]achievement, error)
}

func newHandler(api achievementSource) http.Handler {
	tmpl := template.Must(template.ParseFS(indexHTML, "index.html"))

	s := http.NewServeMux()
	s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		achievements, err := api.GetAchievements()
		if err != nil {
			log.Printf("error fetching achievements: %v", err)
			http.Error(w, "Unable to fetch achievements", http.StatusInternalServerError)
			return
		}

		data := pageData{Achievements: achievements}
		for _, a := range achievements {
			if a.Achieved == 1 {
				data.Completed++
			}
		}

		if len(achievements) > 0 {
			data.Percentage = int(math.Round(float64(data.Completed) / float64(len(achievements)) * 100))
		}

		var page bytes.Buffer
		if err := tmpl.Execute(&page, data); err != nil {
			log.Printf("error rendering achievements: %v", err)
			http.Error(w, "Unable to render achievement journal", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page.Bytes())
	})
	return s
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
