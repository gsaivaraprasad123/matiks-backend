package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func leaderboardHandler(lb *Leaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil {
				limit = v
			}
		}
		json.NewEncoder(w).Encode(lb.GetTop(limit))
	}
}

func searchHandler(lb *Leaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		json.NewEncoder(w).Encode(lb.Search(query))
	}
}
