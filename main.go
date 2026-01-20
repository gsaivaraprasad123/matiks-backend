package main

import (
	"log"
	"net/http"
)

func main() {
	lb := NewLeaderboard()

	seedUsers(lb, 10000)
	lb.StartRatingSimulation()

	http.HandleFunc("/leaderboard", leaderboardHandler(lb))
	http.HandleFunc("/search", searchHandler(lb))

	log.Println("Server running on PORT: 8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
