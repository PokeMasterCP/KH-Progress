package main

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"time"
)

const BASE_URL = "https://api.steampowered.com/ISteamUserStats"

type achievement struct {
	APIName     string `json:"apiname"`
	Achieved    int    `json:"achieved"`
	UnlockTime  int64  `json:"unlocktime"`
	Name        string `json:"-"`
	Description string `json:"-"`
	Hidden      bool   `json:"-"`
	Game        string `json:"-"`
	GameKey     string `json:"-"`
	Icon        string `json:"-"`
	IconGray    string `json:"-"`
}

func (a achievement) Unlocked() bool {
	return a.Achieved == 1
}

// UnlockDate is the zero time when Steam has no unlock timestamp.
func (a achievement) UnlockDate() time.Time {
	if !a.Unlocked() || a.UnlockTime <= 0 {
		return time.Time{}
	}
	return time.Unix(a.UnlockTime, 0).UTC()
}

type achievementsResponse struct {
	PlayerStats struct {
		Achievements []achievement `json:"achievements"`
	} `json:"playerstats"`
}

type schemaResponse struct {
	Game struct {
		AvailableGameStats struct {
			Achievements []schema `json:"achievements"`
		} `json:"availableGameStats"`
	} `json:"game"`
}

type schema struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Hidden      int    `json:"hidden"`
	Icon        string `json:"icon"`
	IconGray    string `json:"icongray"`
}

// game groups the collection's achievements by their ACH_### number range.
type game struct {
	Key         string
	Name        string
	Short       string
	first, last int
}

var games = []game{
	{"kh1", "Kingdom Hearts Final Mix", "KH I", 1, 55},
	{"recom", "Re:Chain of Memories", "Re:CoM", 56, 102},
	{"kh2", "Kingdom Hearts II Final Mix", "KH II", 103, 152},
	{"bbs", "Birth by Sleep Final Mix", "BbS", 153, 197},
}

func (s *steamAPI) GetAchievements() ([]achievement, error) {
	schemaURL := buildURL("GetSchemaForGame", s.apiKey, s.steamId)
	statusURL := buildURL("GetPlayerAchievements", s.apiKey, s.steamId)

	achResponse, err := makeRequest[achievementsResponse](statusURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching achievements: %v", err)
	}
	achievements := achResponse.PlayerStats.Achievements

	schemaResponse, err := makeRequest[schemaResponse](schemaURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching schema: %v", err)
	}
	schema := makeSchema(schemaResponse.Game.AvailableGameStats.Achievements)

	achievements = updateAchievementsFromSchema(achievements, schema, makeGameMap())
	return achievements, nil
}

func buildURL(endpoint, apiKey, steamId string) string {
	var fullURL string
	params := "?key=" + apiKey + "&steamid=" + steamId + "&appid=2552430"

	switch endpoint {
	case "GetSchemaForGame":
		fullURL = BASE_URL + "/GetSchemaForGame/v2/" + params
	case "GetPlayerAchievements":
		fullURL = BASE_URL + "/GetPlayerAchievements/v1/" + params
	default:
		return ""
	}

	return fullURL
}

func makeRequest[T any](url string) (T, error) {
	var result T
	resp, err := http.Get(url)
	if err != nil {
		return result, fmt.Errorf("error making GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	if err := json.UnmarshalRead(resp.Body, &result); err != nil {
		return result, fmt.Errorf("Error parsing achievements JSON: %v", err)
	}

	return result, nil
}

func makeSchema(s []schema) map[string]schema {
	schemaMap := make(map[string]schema, len(s))
	for _, achievement := range s {
		schemaMap[achievement.Name] = achievement
	}
	return schemaMap
}

func updateAchievementsFromSchema(achievements []achievement, schema map[string]schema, games map[string]game) []achievement {
	for i := range achievements {
		s := schema[achievements[i].APIName]
		achievements[i].Name = s.DisplayName
		achievements[i].Description = s.Description
		achievements[i].Hidden = s.Hidden == 1
		achievements[i].Icon = s.Icon
		achievements[i].IconGray = s.IconGray
		g := games[achievements[i].APIName]
		achievements[i].Game = g.Name
		achievements[i].GameKey = g.Key
	}
	return achievements
}

func makeGameMap() map[string]game {
	result := make(map[string]game, 197)
	for _, game := range games {
		for n := game.first; n <= game.last; n++ {
			result[fmt.Sprintf("ACH_%03d", n)] = game
		}
	}
	return result
}
