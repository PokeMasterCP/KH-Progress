package main

import (
	"bytes"
	"cmp"
	"embed"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"time"
)

const (
	port = ":8080"
)

//go:embed index.html
var indexHTML embed.FS

const recentLimit = 5

type pageData struct {
	Achievements []achievement
	Games        []gameProgress
	Recent       []achievement
	Total        int
	Completed    int
	Remaining    int
	Percentage   int
	FirstUnlock  time.Time
	LatestUnlock time.Time
	SyncedAt     time.Time
	Failed       bool
}

type gameProgress struct {
	Key          string
	Name         string
	Total        int
	Completed    int
	Percentage   int
	LatestUnlock time.Time
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
	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"iso":   func(t time.Time) string { return t.Format(time.RFC3339) },
		"date":  func(t time.Time) string { return t.Format("Jan 2, 2006") },
		"month": func(t time.Time) string { return t.Format("Jan 2006") },
	}).ParseFS(indexHTML, "index.html"))

	render := func(w http.ResponseWriter, status int, data pageData) {
		var page bytes.Buffer
		if err := tmpl.Execute(&page, data); err != nil {
			log.Printf("error rendering achievements: %v", err)
			http.Error(w, "Unable to render achievement journal", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		_, _ = w.Write(page.Bytes())
	}

	s := http.NewServeMux()
	s.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		achievements, err := api.GetAchievements()
		if err != nil {
			log.Printf("error fetching achievements: %v", err)
			render(w, http.StatusInternalServerError, pageData{Failed: true, SyncedAt: time.Now().UTC()})
			return
		}
		render(w, http.StatusOK, buildPage(achievements, time.Now()))
	})
	return s
}

// buildPage summarises the achievements overall, per game, and by most recent unlock.
func buildPage(achievements []achievement, now time.Time) pageData {
	data := pageData{Achievements: achievements, Total: len(achievements), SyncedAt: now.UTC()}
	for _, g := range games {
		data.Games = append(data.Games, gameProgress{Key: g.Key, Name: g.Name})
	}
	other := gameProgress{Key: "other", Name: "Other achievements"}

	var dated []achievement
	for _, a := range achievements {
		progress := &other
		for i := range data.Games {
			if data.Games[i].Key == a.GameKey {
				progress = &data.Games[i]
				break
			}
		}
		progress.Total++
		if !a.Unlocked() {
			continue
		}
		data.Completed++
		progress.Completed++
		if t := a.UnlockDate(); !t.IsZero() {
			dated = append(dated, a)
			if t.After(progress.LatestUnlock) {
				progress.LatestUnlock = t
			}
		}
	}
	data.Games = append(data.Games, other)
	data.Games = slices.DeleteFunc(data.Games, func(g gameProgress) bool { return g.Total == 0 })
	for i := range data.Games {
		data.Games[i].Percentage = percent(data.Games[i].Completed, data.Games[i].Total)
	}
	data.Remaining = data.Total - data.Completed
	data.Percentage = percent(data.Completed, data.Total)

	slices.SortStableFunc(dated, func(a, b achievement) int { return cmp.Compare(b.UnlockTime, a.UnlockTime) })
	if len(dated) > 0 {
		data.LatestUnlock = dated[0].UnlockDate()
		data.FirstUnlock = dated[len(dated)-1].UnlockDate()
	}
	data.Recent = dated[:min(len(dated), recentLimit)]
	return data
}

func percent(completed, total int) int {
	if total == 0 {
		return 0
	}
	return int(math.Round(float64(completed) / float64(total) * 100))
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
