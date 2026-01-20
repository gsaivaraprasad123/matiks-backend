## Scalable Leaderboard System with Search & Tie-Aware Ranking (Backend)

This backend powers a **highly scalable, in-memory leaderboard** with:

- **Tie-aware ranking** (same rating ⇒ same rank group, rank is based on how many users have strictly higher rating)
- **Fast top-N queries** (e.g. top 50 players)
- **Real-time rating simulation** to mimic a live, evolving system
- **Case-insensitive search** by username
- **CORS-enabled HTTP API** for easy integration with web/mobile frontends

The service is written in **Go** and currently runs as a single in-memory instance, but the design is structured so it can be extended toward more distributed/sharded setups.

The backend is deployed on **Render** and publicly available at:

- **Base URL**: `https://matiks-backend-mamw.onrender.com`

---

## Features

- **In-memory leaderboard**
  - Stores users with `id`, `username`, and `rating`.
  - Uses arrays to maintain **rating frequency** and a **prefix sum of higher ratings** to compute ranks efficiently.

- **Tie-aware ranking**
  - Rank is computed based on **how many users have a strictly higher rating**, so all users with the same rating share the same rank group.
  - Example: if 10 users have a higher rating than rating `R`, then every user with rating `R` gets rank `11`.

- **Top-N leaderboard**
  - Fetch the top `N` players ordered by rating (descending).
  - Uses the tie-aware rank computation for each entry.

- **Username search**
  - Case-insensitive substring search over usernames.
  - Returns each matching user with their current rank and rating.

- **Continuous rating simulation**
  - A background goroutine randomly adjusts player ratings every 100ms.
  - Keeps the leaderboard changing over time, simulating real-world dynamics.

- **CORS-enabled API**
  - `Access-Control-Allow-Origin: *` set on all endpoints to simplify integration from any frontend origin.

---

## High-Level Architecture

- **Language**: Go
- **Entry point**: `main.go`
- **Core components**:
  - `Leaderboard` struct in `leaderboard.go`:
    - Holds:
      - `users`: `map[int64]*User`
      - `usernameIndex`: `map[string][]int64` (for search)
      - `ratingCount`: frequency of each rating
      - `higherCount`: prefix sums used to compute ranks
    - Methods:
      - `AddUser(u *User)`
      - `UpdateRating(userID int64, delta int)`
      - `GetTop(limit int) []LeaderboardEntry`
      - `Search(query string) []LeaderboardEntry`
      - `StartRatingSimulation()`
  - Models in `models.go`:
    - `User`: `{ id, username, rating }`
    - `LeaderboardEntry`: `{ rank, username, rating }`
  - Utility functions in `utils.go`:
    - `seedUsers(lb *Leaderboard, n int)` – seeds `n` users with random usernames and ratings.
    - `enableCORS(next http.Handler)` – wraps handlers with CORS headers.
  - HTTP handlers in `handlers.go`:
    - `leaderboardHandler(lb *Leaderboard)`
    - `searchHandler(lb *Leaderboard)`

`main.go` wires everything together:

- Creates a new `Leaderboard` with `NewLeaderboard()`.
- Seeds it with 10,000 users (`seedUsers(lb, 10000)`).
- Starts the continuous rating simulation (`lb.StartRatingSimulation()`).
- Exposes HTTP endpoints `/leaderboard` and `/search`.

---

## Data Model

- **User**
  - `id` (`int64`)
  - `username` (`string`)
  - `rating` (`int`)
  - Ratings are clamped between `minRating = 100` and `maxRating = 5000`.

- **LeaderboardEntry**
  - `rank` (`int`) – computed via prefix sums over ratings
  - `username` (`string`)
  - `rating` (`int`)

---

## Ranking Algorithm (Tie-Aware)

The leaderboard maintains:

- `ratingCount[r]`: number of users with rating `r`
- `higherCount[r]`: number of users with rating strictly greater than `r`

On any rating change or new user:

1. Update `ratingCount[oldRating]--` and `ratingCount[newRating]++`.
2. Recompute `higherCount` from `maxRating` down to `minRating`:
   - `higherCount[r] = totalHigher`
   - `totalHigher += ratingCount[r]`
3. A user’s rank is:

   \[
   \text{rank} = \text{higherCount}[\text{rating}] + 1
   \]

This means:

- All users with the same rating share the same rank number.
- The rank is stable under concurrent reads because operations are protected with a read/write mutex.

---

## HTTP API

### Base URL

- **Production (Render)**: `https://matiks-backend-mamw.onrender.com`

---

### `GET /leaderboard`

**Description**: Fetch the top N users ordered by rating (desc), with tie-aware ranks.

- **Query parameters**:
  - `limit` (optional, `int`, default `50`): maximum number of entries to return.

- **Response** – `200 OK`, JSON array of `LeaderboardEntry`:

```json
[
  {
    "rank": 1,
    "username": "virat_9999",
    "rating": 4980
  },
  {
    "rank": 1,
    "username": "virat_1234",
    "rating": 4980
  },
  {
    "rank": 3,
    "username": "rohit_42",
    "rating": 4950
  }
]
```

**Example request**:

```bash
curl "https://matiks-backend-mamw.onrender.com/leaderboard?limit=50"
```

---

### `GET /search`

**Description**: Search players by username (case-insensitive substring), returning their current ranks and ratings.

- **Query parameters**:
  - `query` (**required**, `string`): substring to search for in usernames.

- **Response** – `200 OK`, JSON array of `LeaderboardEntry`:

```json
[
  {
    "rank": 120,
    "username": "rahul_102",
    "rating": 3105
  },
  {
    "rank": 467,
    "username": "rahul_876",
    "rating": 2650
  }
]
```

**Example request**:

```bash
curl "https://matiks-backend-mamw.onrender.com/search?query=rahul"
```

---

## Local Development

### Prerequisites

- Go (version from `go.mod`, e.g. Go 1.24+)

### Install dependencies

From the `backend` directory:

```bash
cd /Users/saivaraprasadgandhe/Developer/matiks/backend
go mod tidy
```

### Run locally

```bash
go run .
```

By default the server listens on:

- `http://localhost:8080`

You can override the port via the `PORT` environment variable:

```bash
PORT=9090 go run .
```

---

## Render Deployment

The backend is deployed to **Render** as a web service:

- **URL**: `https://matiks-backend-mamw.onrender.com`
- **Environment**:
  - Render automatically injects the `PORT` environment variable.
  - The app binds to `:${PORT}` in `main.go`.
- **Build & Run**:
  - Build command: Render can auto-detect (`go build`).
  - Start command: `go run .` or `./leaderboard` (depending on your Render configuration).

Since the leaderboard is **in-memory**, deployment restarts will reset the data. The service reseeds itself with 10,000 random users and restarts the rating simulation on each boot.