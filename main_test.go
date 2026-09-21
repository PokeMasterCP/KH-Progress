package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
			newHandler(tt.data).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
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
	handler := newHandler(nil)
	for _, path := range []string{"/api/achievements", "/missing", "/index.html"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", path, response.Code)
		}
	}
}
