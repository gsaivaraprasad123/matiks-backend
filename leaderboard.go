package main

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	minRating = 100
	maxRating = 5000
)

type Leaderboard struct {
	mu sync.RWMutex

	users         map[int64]*User
	usernameIndex map[string][]int64

	ratingCount [maxRating + 1]int
	higherCount [maxRating + 2]int
}

func NewLeaderboard() *Leaderboard {
	lb := &Leaderboard{
		users:         make(map[int64]*User),
		usernameIndex: make(map[string][]int64),
	}
	return lb
}

//rank logic

func (lb *Leaderboard) recomputePrefix() {
	total := 0
	for r := maxRating; r >= minRating; r-- {
		lb.higherCount[r] = total
		total += lb.ratingCount[r]
	}
}

func (lb *Leaderboard) getRank(rating int) int {
	return lb.higherCount[rating] + 1
}

//user ops

func (lb *Leaderboard) AddUser(u *User) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.users[u.ID] = u
	lb.ratingCount[u.Rating]++
	lb.usernameIndex[u.Username] = append(lb.usernameIndex[u.Username], u.ID)
	lb.recomputePrefix()
}

func (lb *Leaderboard) UpdateRating(userID int64, delta int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	user := lb.users[userID]
	if user == nil {
		return
	}

	old := user.Rating
	newRating := old + delta

	if newRating < minRating {
		newRating = minRating
	}
	if newRating > maxRating {
		newRating = maxRating
	}

	if old == newRating {
		return
	}

	lb.ratingCount[old]--
	lb.ratingCount[newRating]++
	user.Rating = newRating
	lb.recomputePrefix()
}

// queuries

func (lb *Leaderboard) GetTop(limit int) []LeaderboardEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	res := []LeaderboardEntry{}

	for rating := maxRating; rating >= minRating && len(res) < limit; rating-- {
		for _, u := range lb.users {
			if u.Rating == rating {
				res = append(res, LeaderboardEntry{
					Rank:     lb.getRank(rating),
					Username: u.Username,
					Rating:   u.Rating,
				})
				if len(res) >= limit {
					break
				}
			}
		}
	}

	return res
}

func (lb *Leaderboard) Search(query string) []LeaderboardEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	query = strings.ToLower(query)
	res := []LeaderboardEntry{}

	for _, u := range lb.users {
		if strings.Contains(strings.ToLower(u.Username), query) {
			res = append(res, LeaderboardEntry{
				Rank:     lb.getRank(u.Rating),
				Username: u.Username,
				Rating:   u.Rating,
			})
		}
	}
	return res
}

// sim

func (lb *Leaderboard) StartRatingSimulation() {
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		for {
			id := int64(r.Intn(len(lb.users)) + 1)
			delta := r.Intn(100) - 50
			lb.UpdateRating(id, delta)
			time.Sleep(100 * time.Millisecond)
		}
	}()
}
