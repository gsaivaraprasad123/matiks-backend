# Scalable Leaderboard System with Search & Tie-Aware Ranking (Backend)

This repository contains the backend for a high-performance, tie-aware leaderboard system designed to handle large user bases with fast ranking, search, and live updates.

The project was developed in two stages:

- **Baseline implementation** – focused on correctness, clarity, and correctness of tie-aware ranking.
- **Optimized million-scale implementation** – refactored to remove global scans and scale efficiently to 1M+ users using minimal structural changes.

The backend is written in Go and currently runs as a single in-memory instance, but the internal data structures are designed to closely mirror how production systems (e.g. Redis sorted sets) work, making future scaling straightforward.

The backend is deployed on Render and publicly available at:

**Base URL:**
 https://matiks-backend-mamw.onrender.com

## Key Features

- **Tie-aware ranking**
  - Users with the same rating share the same rank.
  - Rank is based on how many users have a strictly higher rating.

- **Fast top-N leaderboard queries**
  - Efficiently fetch top players (e.g. top 50).

- **Real-time rating simulation**
  - Background goroutine simulates continuous rating changes.

- **Case-insensitive username search**
  - Returns live rank and rating for matching users.

- **CORS-enabled HTTP API**
  - Ready for web and mobile frontends.

## Technology Stack

- **Language**: Go
- **Concurrency**: sync.RWMutex
- **Deployment**: Render
- **Storage**: In-memory 

---

## Version 1 — Baseline Implementation (Correctness-First)

### Goal

The first version focuses on correct ranking logic, thread safety, and clarity.

This version works well for 10k–50k users and is ideal for validating logic and behavior.

### Architecture (Baseline)

#### Core Data Structures

```go
users         map[int64]*User
usernameIndex map[string][]int64

ratingCount [5001]int
higherCount [5002]int
```

- **users**: stores all users by ID.
- **usernameIndex**: supports username search.
- **ratingCount**: number of users per rating.
- **higherCount**: prefix sums used to compute ranks.

### Ranking Algorithm (Tie-Aware)

For each rating r:

- `ratingCount[r]` = number of users with rating r
- `higherCount[r]` = number of users with rating strictly greater than r

**Rank formula:**

```
rank = higherCount[rating] + 1
```

This guarantees:

- Same rating ⇒ same rank
- Rank reflects number of higher-rated users

### Baseline Limitations

While correct, this version has scaling bottlenecks:

| Operation | Complexity | Issue at Scale |
|-----------|-----------|----------------|
| Leaderboard fetch | O(N × R) | Scans all users |
| Search | O(N) | Full scan |
| Rank recompute | O(R) per update | Unnecessary repetition |

Where:

- **N** = number of users
- **R** = rating range (100–5000)

These operations become too slow at hundreds of thousands or millions of users.

---

## Version 2 — Optimized Million-Scale Implementation (Current)

### Goal

Refactor the baseline implementation with minimal changes to support 1M+ users, while preserving correctness and simplicity.

No databases, no Redis — only smarter in-memory structures.

### Key Optimizations

#### 1. Rating Buckets (Critical Change)

Instead of scanning all users, users are bucketed by rating:

```go
ratingBuckets map[int]map[int64]*User
```

This allows:

- Efficient iteration by rating
- No global scans for leaderboard queries

#### 2. Prefix-Based Username Index

Instead of substring scanning:

```go
usernamePrefix map[string][]int64
```

**Example:**

```
"r"     → [1, 4, 9]
"ra"    → [1, 9]
"rah"   → [1]
```

Search becomes:

- O(results) instead of O(N)
- Case-insensitive
- Instant, even at large scale

#### 3. Incremental Rank Maintenance

Instead of recomputing rank prefixes on every update:

- `higherCount` is updated incrementally
- Rating updates adjust only affected ranges

This removes unnecessary recomputation loops.

### Architecture (Optimized)

#### Core Data Structures

```go
users           map[int64]*User
ratingBuckets   map[int]map[int64]*User
usernamePrefix  map[string][]int64

ratingCount [5001]int
higherCount [5002]int
```

### Complexity After Optimization

| Operation | Complexity |
|-----------|-----------|
| Rank lookup | O(1) |
| Leaderboard fetch | O(R + K) |
| Search | O(results) |
| Rating update | O(R) (bounded) |
| Users supported | 1M+ |

Where:

- **R** = 5000 (fixed rating range)
- **K** = top-N size (e.g. 50)

This makes performance independent of total user count for reads.

---

## HTTP API

### Base URL

**Production:**
https://matiks-backend-mamw.onrender.com

### `GET /leaderboard`

Fetch the top N users by rating (descending), with tie-aware ranks.

**Query Params**

- `limit` (optional, default 50)

**Example**

```bash
curl "https://matiks-backend-mamw.onrender.com/leaderboard?limit=50"
```

**Response**

```json
[
  { "rank": 1, "username": "virat_9999", "rating": 4980 },
  { "rank": 1, "username": "virat_1234", "rating": 4980 },
  { "rank": 3, "username": "rohit_42", "rating": 4950 }
]
```

### `GET /search`

Search users by username (case-insensitive).

**Query Params**

- `query` (required)

**Example**

```bash
curl "https://matiks-backend-mamw.onrender.com/search?query=rahul"
```

**Response**

```json
[
  { "rank": 120, "username": "rahul_102", "rating": 3105 },
  { "rank": 467, "username": "rahul_876", "rating": 2650 }
]
```

---

## Local Development

### Prerequisites

- Go (as specified in go.mod)

### Run Locally

```bash
go mod tidy
go run .
```

Server runs on:

- `http://localhost:8080`

Override port if needed:

```bash
PORT=9090 go run .
```

---

## Deployment Notes

- **Deployed on Render**
- Uses dynamic `PORT` from environment
- In-memory storage (data resets on restart)
- Reseeds with 10,000 users on boot
- Rating simulation restarts automatically

---

## Design Rationale (Interview-Ready)

This project demonstrates:

-  Correct handling of tie-aware ranking
-  Awareness of algorithmic bottlenecks
-  Ability to refactor for scale with minimal changes
-  Clean concurrency handling in Go
-  A clear migration path to Redis or sharded systems

The optimized in-memory design closely mirrors how Redis Sorted Sets compute ranks, making future production scaling straightforward.
