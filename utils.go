package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func seedUsers(lb *Leaderboard, n int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	names := []string{
		"rahul", "arjun", "ayush", "rohit", "virat",
		"ananya", "priya", "neha", "riya", "isha",
		"vikram", "aman", "nathan", "vara", "ishant",
	}

	for i := 1; i <= n; i++ {
		base := names[r.Intn(len(names))]

		user := &User{
			ID:       int64(i),
			Username: fmt.Sprintf("%s_%d", base, i),
			Rating:   r.Intn(maxRating-minRating+1) + minRating,
		}
		lb.AddUser(user)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
