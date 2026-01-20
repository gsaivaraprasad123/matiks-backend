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

	users map[int64]*User

	ratingBuckets map[int]map[int64]*User

	usernamePrefix map[string][]int64

	ratingCount [maxRating + 1]int
	higherCount [maxRating + 2]int
}

func NewLeaderboard() *Leaderboard {
	return &Leaderboard{
		users:          make(map[int64]*User),
		ratingBuckets:  make(map[int]map[int64]*User),
		usernamePrefix: make(map[string][]int64),
	}
}

func (lb *Leaderboard) getRank(rating int) int {
	return lb.higherCount[rating] + 1
}

//ops

func (lb *Leaderboard) indexUsername(username string, id int64) {
	username = strings.ToLower(username)
	for i := 1; i <= len(username); i++ {
		prefix := username[:i]
		lb.usernamePrefix[prefix] = append(lb.usernamePrefix[prefix], id)
	}
}

func (lb *Leaderboard) AddUser(u *User) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.users[u.ID] = u

	if lb.ratingBuckets[u.Rating] == nil {
		lb.ratingBuckets[u.Rating] = make(map[int64]*User)
	}
	lb.ratingBuckets[u.Rating][u.ID] = u

	lb.ratingCount[u.Rating]++

	for r := u.Rating - 1; r >= minRating; r-- {
		lb.higherCount[r]++
	}

	lb.indexUsername(u.Username, u.ID)
}

func (lb *Leaderboard) UpdateRating(userID int64, delta int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	u := lb.users[userID]
	if u == nil {
		return
	}

	old := u.Rating
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

	delete(lb.ratingBuckets[old], userID)
	lb.ratingCount[old]--

	for r := old - 1; r >= minRating; r-- {
		lb.higherCount[r]--
	}

	if lb.ratingBuckets[newRating] == nil {
		lb.ratingBuckets[newRating] = make(map[int64]*User)
	}
	lb.ratingBuckets[newRating][userID] = u
	lb.ratingCount[newRating]++

	for r := newRating - 1; r >= minRating; r-- {
		lb.higherCount[r]++
	}

	u.Rating = newRating
}

//queries

func (lb *Leaderboard) GetTop(limit int) []LeaderboardEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	res := []LeaderboardEntry{}

	for rating := maxRating; rating >= minRating && len(res) < limit; rating-- {
		for _, u := range lb.ratingBuckets[rating] {
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
	return res
}

func (lb *Leaderboard) Search(query string) []LeaderboardEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	query = strings.ToLower(query)
	ids := lb.usernamePrefix[query]

	res := []LeaderboardEntry{}
	for _, id := range ids {
		u := lb.users[id]
		res = append(res, LeaderboardEntry{
			Rank:     lb.getRank(u.Rating),
			Username: u.Username,
			Rating:   u.Rating,
		})
	}
	return res
}

//sim

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
