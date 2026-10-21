package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"krillion-solver/internal/game"
)

func TestWebHandler(t *testing.T) {
	handler := webHandler(func(context.Context) (game.Result, error) { return game.Demo(), nil })
	for _, path := range []string{"/", "/styles.css", "/app.js", "/data.json"} {
		r := httptest.NewRecorder()
		handler.ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
		if r.Code != http.StatusOK {
			t.Fatalf("%s returned %d", path, r.Code)
		}
		if path == "/data.json" && (!strings.Contains(r.Body.String(), `"demo":true`) || r.Header().Get("Cache-Control") != "no-store") {
			t.Fatal("daily data was not served as a fresh solver result")
		}
	}
}

func TestWebLoadFailure(t *testing.T) {
	handler := webHandler(func(context.Context) (game.Result, error) { return game.Result{}, errors.New("unavailable") })
	r := httptest.NewRecorder()
	handler.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/data.json", nil))
	if r.Code != http.StatusBadGateway || !strings.Contains(r.Body.String(), "unavailable") {
		t.Fatalf("unexpected response: %d %s", r.Code, r.Body.String())
	}
}
