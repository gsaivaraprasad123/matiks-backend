package main

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}
