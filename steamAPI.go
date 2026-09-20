package main

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
)

const BASE_URL = "https://api.steampowered.com/ISteamUserStats"

type achievement struct {
	Name     string `json:"apiname"`
	Achieved int    `json:"achieved"`
	icon     string
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

func GetAchievements(apiKey, steamId string) ([]achievement, error) {
	schemaURL := buildURL("GetSchemaForGame", apiKey, steamId)
	statusURL := buildURL("GetPlayerAchievements", apiKey, steamId)

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

	achievements = updateAchievementsFromSchema(achievements, schema)
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

func updateAchievementsFromSchema(achievements []achievement, s map[string]schema) []achievement {
	for i := range achievements {
		info, ok := s[achievements[i].Name]
		if !ok {
			continue
		}

		achievements[i].Name = info.DisplayName
		achievements[i].icon = info.Icon
	}
	return achievements
}
