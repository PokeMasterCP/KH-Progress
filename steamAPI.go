package main

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
)

const BASE_URL = "https://api.steampowered.com/ISteamUserStats"

type achievement struct {
	APIName  string `json:"apiname"`
	Name     string
	Achieved int `json:"achieved"`
	Game     string
	Icon     string
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
	Icon        string `json:"icon"`
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

func updateAchievementsFromSchema(achievements []achievement, schema map[string]schema, games map[string]string) []achievement {
	for i := range achievements {
		achievements[i].Name = schema[achievements[i].APIName].DisplayName
		achievements[i].Icon = schema[achievements[i].APIName].Icon
		achievements[i].Game = games[achievements[i].APIName]
	}
	return achievements
}

func makeGameMap() map[string]string {
	games := []struct {
		first, last int
		name        string
	}{
		{1, 55, "Kingdom Hearts Final Mix"},
		{56, 102, "Re:Chain of Memories"},
		{103, 152, "Kingdom Hearts II Final Mix"},
		{153, 197, "Birth by Sleep Final Mix"},
	}

	result := make(map[string]string, 197)
	for _, game := range games {
		for n := game.first; n <= game.last; n++ {
			result[fmt.Sprintf("ACH_%03d", n)] = game.name
		}
	}
	return result
}
