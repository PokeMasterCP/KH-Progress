package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type achievementSourceFunc func() ([]achievement, error)

func (f achievementSourceFunc) GetAchievements() ([]achievement, error) {
	return f()
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
				{APIName: "ACH_001", Name: "A journey begins", Achieved: 1, Game: "Kingdom Hearts Final Mix", Icon: "https://example.com/icon.jpg"},
				{APIName: "ACH_002", Achieved: 1},
				{APIName: "ACH_003", Name: "Still locked", Achieved: 0},
			},
			want: []string{
				"2 of 3 achievements unlocked", "3 achievements · 67% complete",
				`value="67"`, "<h3>A journey begins</h3>", "<h3>ACH_002</h3>",
				`data-game="Kingdom Hearts Final Mix" data-achieved="1"`,
				`src="https://example.com/icon.jpg"`, `class="achievement locked"`,
				`aria-label="Unlocked"`, `aria-label="Locked"`,
				`role="status" hidden>No achievements`,
			},
		},
		{
			name: "empty collection",
			want: []string{"0 of 0 achievements unlocked", `value="0"`, `role="status">No achievements found for this collection.`},
		},
		{
			name: "escape Steam data",
			data: []achievement{{Name: "<script>alert(1)</script>", Game: `" onclick="alert(1)`, Icon: "javascript:alert(1)"}},
			want: []string{"&lt;script&gt;alert(1)&lt;/script&gt;", `data-game="&#34; onclick=&#34;alert(1)"`, `src="#ZgotmplZ"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			api := achievementSourceFunc(func() ([]achievement, error) { return tt.data, nil })
			newHandler(api).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Header().Get("Content-Type"))
			}
			for _, want := range tt.want {
				if !strings.Contains(response.Body.String(), want) {
					t.Errorf("rendered HTML missing %q", want)
				}
			}
		})
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
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
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
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusInternalServerError || strings.TrimSpace(response.Body.String()) != "Unable to fetch achievements" {
		t.Fatalf("unexpected error response: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "1 of 1 achievements unlocked") {
		t.Fatalf("next page load did not recover: %d %s", response.Code, response.Body.String())
	}
}
