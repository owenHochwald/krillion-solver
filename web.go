package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"krillion-solver/internal/game"
)

//go:embed web/*
var webFiles embed.FS

func webHandler(load func(context.Context) (game.Result, error)) http.Handler {
	assets, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(assets))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /data.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()
		result, err := load(ctx)
		if err != nil {
			http.Error(w, "Could not load today's answers: "+game.SafeText(err.Error()), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.Handle("GET /", files)
	return mux
}

func serveWeb(ctx context.Context, addr string, demo bool) error {
	load := func(ctx context.Context) (game.Result, error) {
		if demo {
			return game.Demo(), nil
		}
		return game.NewClient().Solve(ctx)
	}
	server := &http.Server{Addr: addr, Handler: webHandler(load), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
