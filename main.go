package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	lb := NewLeaderboard()

	seedUsers(lb, 10000)
	lb.StartRatingSimulation()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/leaderboard", enableCORS(leaderboardHandler(lb)))
	http.Handle("/search", enableCORS(searchHandler(lb)))

	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
