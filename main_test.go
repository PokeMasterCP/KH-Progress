package main

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type achievementSourceFunc func() ([]achievement, error)

func (f achievementSourceFunc) GetAchievements() ([]achievement, error) {
	return f()
}

func get(t *testing.T, handler http.Handler) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	return response
}

func TestAchievementPage(t *testing.T) {
	tests := []struct {
		name string
		data []achievement
		want []string
	}{
		{
			name: "rendered achievements and rounded progress",
			data: []achievement{
				{APIName: "ACH_001", Name: "A journey begins", Description: "Start the adventure.", Achieved: 1, UnlockTime: 1709424000, Game: "Kingdom Hearts Final Mix", GameKey: "kh1", Icon: "https://example.com/icon.jpg"},
				{APIName: "ACH_002", Achieved: 1, GameKey: "kh1", Game: "Kingdom Hearts Final Mix"},
				{APIName: "ACH_003", Name: "Still locked", Hidden: true, Achieved: 0, GameKey: "kh1", Game: "Kingdom Hearts Final Mix", Icon: "https://example.com/color.jpg", IconGray: "https://example.com/gray.jpg"},
			},
			want: []string{
				"2 of 3 achievements unlocked", "3 achievements · 67% complete",
				"67<small>%</small>", `<h3 class="ach-name">A journey begins</h3>`, `<h3 class="ach-name">ACH_002</h3>`,
				`<p class="ach-desc">Start the adventure.</p>`, "Hidden achievement. Steam keeps its description secret.",
				`class="ach" data-game="kh1" data-achieved="1" data-time="1709424000"`,
				`src="https://example.com/icon.jpg"`, `src="https://example.com/gray.jpg"`, `class="ach locked"`,
				`datetime="2024-03-03T00:00:00Z" data-format="date">Mar 3, 2024</time>`,
				`<span class="game-count"><b>2<span>/3</span></b><em>67%</em></span>`, `style="--c: 2; --n: 3"`,
				`data-game="kh1" data-name="Kingdom Hearts Final Mix"`,
				`role="status" hidden>`,
			},
		},
		{
			name: "empty collection",
			want: []string{"No achievements came back from Steam yet.", "0<small>%</small>", `role="status">`, "Steam returned no achievements."},
		},
		{
			name: "escape Steam data",
			data: []achievement{{Name: "<script>alert(1)</script>", Description: "<b>bold</b>", Game: `" onclick="alert(1)`, Icon: "javascript:alert(1)"}},
			want: []string{"&lt;script&gt;alert(1)&lt;/script&gt;", "&lt;b&gt;bold&lt;/b&gt;", `&#34; onclick=&#34;alert(1)`, `src="#ZgotmplZ"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := get(t, newHandler(achievementSourceFunc(func() ([]achievement, error) { return tt.data, nil })))
			if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Header().Get("Content-Type"))
			}
			if got := response.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			body := response.Body.String()
			if strings.Contains(body, "ZgotmplZ") && tt.name != "escape Steam data" {
				t.Errorf("template rejected a value as unsafe:\n%s", body)
			}
			for _, want := range tt.want {
				if !strings.Contains(body, want) {
					t.Errorf("rendered HTML missing %q", want)
				}
			}
		})
	}
}

func TestBuildPage(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	data := buildPage([]achievement{
		{APIName: "ACH_001", GameKey: "kh1", Achieved: 1, UnlockTime: 300},
		{APIName: "ACH_002", GameKey: "kh1", Achieved: 1, UnlockTime: 100},
		{APIName: "ACH_056", GameKey: "recom", Achieved: 0},
		{APIName: "ACH_103", GameKey: "kh2", Achieved: 1, UnlockTime: 200},
		{APIName: "ACH_104", GameKey: "kh2", Achieved: 1},
		{APIName: "MYSTERY", Achieved: 1, UnlockTime: 400},
	}, now)

	if data.Total != 6 || data.Completed != 5 || data.Remaining != 1 || data.Percentage != 83 {
		t.Errorf("totals = %d/%d remaining %d at %d%%, want 5/6 remaining 1 at 83%%", data.Completed, data.Total, data.Remaining, data.Percentage)
	}
	var keys []string
	for _, g := range data.Games {
		keys = append(keys, g.Key)
	}
	if got := strings.Join(keys, ","); got != "kh1,recom,kh2,other" {
		t.Errorf("games = %s, want kh1,recom,kh2,other (games without achievements omitted)", got)
	}
	if kh2 := data.Games[2]; kh2.Completed != 2 || kh2.Total != 2 || kh2.Percentage != 100 || kh2.LatestUnlock.Unix() != 200 {
		t.Errorf("kh2 progress = %+v", kh2)
	}
	if recom := data.Games[1]; recom.Percentage != 0 || !recom.LatestUnlock.IsZero() {
		t.Errorf("recom progress = %+v", recom)
	}
	var recent []string
	for _, a := range data.Recent {
		recent = append(recent, a.APIName)
	}
	if got := strings.Join(recent, ","); got != "MYSTERY,ACH_001,ACH_103,ACH_002" {
		t.Errorf("recent = %s, want newest dated unlocks first", got)
	}
	if data.FirstUnlock.Unix() != 100 || data.LatestUnlock.Unix() != 400 || !data.SyncedAt.Equal(now) {
		t.Errorf("first %v latest %v synced %v", data.FirstUnlock, data.LatestUnlock, data.SyncedAt)
	}
}

func TestSteamResponsesDecode(t *testing.T) {
	var player achievementsResponse
	if err := json.Unmarshal([]byte(`{"playerstats":{"steamID":"1","gameName":"KH","achievements":[
		{"apiname":"ACH_001","achieved":1,"unlocktime":1709424000},
		{"apiname":"ACH_150","achieved":0,"unlocktime":0}],"success":true}}`), &player); err != nil {
		t.Fatal(err)
	}
	var game schemaResponse
	if err := json.Unmarshal([]byte(`{"game":{"gameName":"KH","availableGameStats":{"achievements":[
		{"name":"ACH_001","defaultvalue":0,"displayName":"Dive into the Heart","hidden":0,"description":"Begin the journey.","icon":"https://cdn/1.jpg","icongray":"https://cdn/1g.jpg"},
		{"name":"ACH_150","defaultvalue":0,"displayName":"Secret","hidden":1,"icon":"https://cdn/2.jpg","icongray":"https://cdn/2g.jpg"}]}}}`), &game); err != nil {
		t.Fatal(err)
	}

	got := updateAchievementsFromSchema(player.PlayerStats.Achievements, makeSchema(game.Game.AvailableGameStats.Achievements), makeGameMap())
	want := []achievement{
		{APIName: "ACH_001", Achieved: 1, UnlockTime: 1709424000, Name: "Dive into the Heart", Description: "Begin the journey.", Game: "Kingdom Hearts Final Mix", GameKey: "kh1", Icon: "https://cdn/1.jpg", IconGray: "https://cdn/1g.jpg"},
		{APIName: "ACH_150", Name: "Secret", Hidden: true, Game: "Kingdom Hearts II Final Mix", GameKey: "kh2", Icon: "https://cdn/2.jpg", IconGray: "https://cdn/2g.jpg"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d achievements, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("achievement %d:\n got %+v\nwant %+v", i, got[i], want[i])
		}
	}
	if !got[1].UnlockDate().IsZero() || got[0].UnlockDate().Year() != 2024 {
		t.Errorf("unlock dates: %v %v", got[0].UnlockDate(), got[1].UnlockDate())
	}
}

func TestUnknownRoutes(t *testing.T) {
	handler := newHandler(achievementSourceFunc(func() ([]achievement, error) {
		t.Fatal("unknown routes must not fetch achievements")
		return nil, nil
	}))
	for _, path := range []string{"/api/achievements", "/missing", "/index.html"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", path, response.Code)
		}
	}
}

func TestPageRefreshFetchesLatestAchievements(t *testing.T) {
	calls := 0
	handler := newHandler(achievementSourceFunc(func() ([]achievement, error) {
		calls++
		return []achievement{{Name: "A journey begins", Achieved: calls - 1}}, nil
	}))
	if calls != 0 {
		t.Fatal("achievements were fetched before a page request")
	}
	for _, want := range []string{"0 of 1 achievements unlocked", "1 of 1 achievements unlocked"} {
		response := get(t, handler)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), want) {
			t.Fatalf("expected fresh stats %q, got status %d: %s", want, response.Code, response.Body.String())
		}
	}
	if calls != 2 {
		t.Fatalf("got %d fetches for two page loads, want 2", calls)
	}
}

func TestPageRecoversAfterSteamError(t *testing.T) {
	calls := 0
	handler := newHandler(achievementSourceFunc(func() ([]achievement, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("upstream unavailable")
		}
		return []achievement{{Name: "A journey begins", Achieved: 1}}, nil
	}))
	response := get(t, handler)
	body := response.Body.String()
	if response.Code != http.StatusInternalServerError || !strings.Contains(body, "Steam didn't send your achievements") || strings.Contains(body, "upstream unavailable") {
		t.Fatalf("unexpected error response: %d %s", response.Code, body)
	}
	if response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("error page Content-Type = %q", response.Header().Get("Content-Type"))
	}
	response = get(t, handler)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "1 of 1 achievements unlocked") {
		t.Fatalf("next page load did not recover: %d %s", response.Code, response.Body.String())
	}
}
