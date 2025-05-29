# insider-case

### Description

This project is a RESTful API to  simulate a football league of 4 teams and estimate the championship probabilities.

### Requirements

This project has been developed using go 1.24.3 and postgresql 16+

### How to run?

clone the repo using git clone https://github.com/ahmtakin/insider-case.git

start a postgresql server using docker etc.

create yourself a .env file under project root directory containing 
```bash
#if you are running locally
DB_HOST=localhost
#default postgresql port
DB_PORT=5432
DB_USER=your postgre user name
DB_PASSWORD=your password
DB_NAME=your db name
DB_SSLMODE=disable
```

under project root directory run 'go run main.go' 

access http://localhost:8081 for user interface

### API

#### Base URL

http://localhost:8081/api

#### Endpoints
The API contains different endpoints for different use cases
such as;

##### Create A League - POST /leagues
Users are expected to create a league before simulating it. To Create a league a user must give the league a name, a team count and the teams in with respect to the team size.
Each team must have an arbitrary strength value between 1000 and 3000 as well. the request 
```bash
curl -X POST http://localhost:8081/api/leagues \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Super Kupa",
    "team_count": 4,
    "teams": [
      { "name": "Fenerbahçe", "strength": 2000 },
      { "name": "Beşiktaş", "strength": 2100 },
      { "name": "Galatasaray", "strength": 1800 },
      { "name": "Trabzonspor", "strength": 1700 }
    ]
  }'
```

The response contains the initialized league entities as Teams, Fixtures and initial team stats which all are zero:

the response

```json
    "id": 70,
    "name": "Super Kupa",
    "team_count": 4,
    "max_weeks": 6,
    "curr_week": 1,
    "teams": [{
            "id": 257,
            "league_id": 70,
            "name": "Fenerbahçe",
            "strength": 2000,
            "stats": {
                "team_id": 257,
                "points": 0,
                "played": 0,
                "won": 0,
                "lost": 0,
                "draw": 0,
                "goals_for": 0,
                "goals_against": 0,
                "goal_diff": 0,
                "estimation": 0
            }
        }...
],
    "matches": [
        {
            "id": 709,
            "league_id": 70,
            "week": 1,
            "played": false,
            "home_team": 257,
            "away_team": 258,
            "home_score": 0,
            "away_score": 0
        }...
]

```
#### Simulate A Week - POST /leagues/simulate-week

If a user wants to the current week to be simulated, they can send a POST request to this endpoint with respective league id.

the request:
```bash
curl -X POST http://localhost:8081/api/leagues/simulate-week \
  -H "Content-Type: application/json" \
  -d '{
    "leagueID": 13
  }'
```
The response contains the updated league week, matches simulated and their results, updated team stats based on last played matches and a champion model that is expected to return the champion teams name in the last week.
the response:
```json
{
    "league_id": 13,
    "week": 3,
    "matches": [
        {
            "id": 147,
            "league_id": 13,
            "week": 2,
            "played": true,
            "home_team": 50,
            "away_team": 49,
            "home_score": 0,
            "away_score": 2
        }...
    ],
    "team_stats": [
        {
            "team_id": 50,
            "points": 1,
            "played": 2,
            "won": 0,
            "lost": 1,
            "draw": 1,
            "goals_for": 0,
            "goals_against": 2,
            "goal_diff": -2,
            "estimation": 0
        }...
    ],
     "champion": {"id": 0,
        "league_id": 0,
        "name": "",
        "strength": 0,
        "stats": {...}
    }
```

#### User Manually Play A Week - POST /leagues/user-play-week

This endpoint serves as an option for users to enter a weeks results manually. 

the request:
```bash
curl -X POST http://localhost:8081/api/leagues/user-play-week \
  -H "Content-Type: application/json" \
  -d '[
    {
        "match_id": 149,
        "league_id": 13,
        "week": 3,
        "home_team_id": 51,
        "away_team_id": 49,
        "home_score": 2,
        "away_score": 0
    },
    {
        "match_id": 150,
        "league_id": 13,
        "week": 3,
        "home_team_id": 52,
        "away_team_id": 50,
        "home_score": 0,
        "away_score": 3
    }
]'
```
the response contains the same items with simulate-week endpoint
the response:
```json
{
    "league_id": 13,
    "week": 4,
    "matches": [
        {
            "id": 149,
            "league_id": 13,
            "week": 3,
            "played": true,
            "home_team": 51,
            "away_team": 49,
            "home_score": 2,
            "away_score": 0,
            "result": 51
        },...]
    "team_stats": [
        {
            "team_id": 51,
            "points": 7,
            "played": 3,
            "won": 2,
            "lost": 0,
            "draw": 1,
            "goals_for": 3,
            "goals_against": 0,
            "goal_diff": 3,
            "estimation": 0
        },...]
    "champion": {
        "id": 0,
        "league_id": 0,
        "name": "",
        "strength": 0,
        "stats": {...}
  }}

```

#### Simulate All Remaining Matches - POST /leagues/play-remaining-matches

This endpoint allows users to simulate the remaining part of a league from the current week. 
the request:

```bash
curl -X POST http://localhost:8081/api/leagues/play-remaining-matches \
  -H "Content-Type: application/json" \
  -d '{
    "leagueID": 13
  }'
```
The response contains same items with previous 2 requests. but it will have a champion that is not null as well!!

```json
[...
,{
        "league_id": 13,
        "week": 6,
        "matches": [
            {
                "id": 155,
                "league_id": 13,
                "week": 6,
                "played": true,
                "home_team": 49,
                "away_team": 51,
                "home_score": 1,
                "away_score": 2
            }...
        ],
        "team_stats": [
            {
                "team_id": 49,
                "points": 9,
                "played": 6,
                "won": 3,
                "lost": 3,
                "draw": 0,
                "goals_for": 8,
                "goals_against": 7,
                "goal_diff": 1,
                "estimation": 0
            }...
        ],
        "champion": {
            "id": 51,
            "league_id": 13,
            "name": "Galatasaray",
            "strength": 1800,
            "stats": {
                "team_id": 0,
                "points": 0,
                "played": 0,
                "won": 0,
                "lost": 0,
                "draw": 0,
                "goals_for": 0,
                "goals_against": 0,
                "goal_diff": 0,
                "estimation": 0
            }
        }
    }
]
```

### Championship estimations
Another feature provided by the API is that for each week after week 4 the program simulates the remaining part of the league 10000 times and returns a championship estimation for each team in their team_stats.
```json
"team_stats": [
            {
                "team_id": 52,
                "points": 0,
                "played": 4,
                "won": 0,
                "lost": 4,
                "draw": 0,
                "goals_for": 0,
                "goals_against": 8,
                "goal_diff": -8,
                "estimation": 0
            },
            {
                "team_id": 49,
                "points": 9,
                "played": 4,
                "won": 3,
                "lost": 1,
                "draw": 0,
                "goals_for": 6,
                "goals_against": 2,
                "goal_diff": 4,
                "estimation": 0.6646
            },
            {
                "team_id": 50,
                "points": 4,
                "played": 4,
                "won": 1,
                "lost": 2,
                "draw": 1,
                "goals_for": 3,
                "goals_against": 4,
                "goal_diff": -1,
                "estimation": 0
            },
            {
                "team_id": 51,
                "points": 10,
                "played": 4,
                "won": 3,
                "lost": 0,
                "draw": 1,
                "goals_for": 5,
                "goals_against": 0,
                "goal_diff": 5,
                "estimation": 0.3354
            }
        ]
```

#### Simulation and Estimation Logic

The simulation logic follows a simple path to determine a match's result based on team strengths, home team advantage and form factor.

##### Team strength
an arbitrarily selected number between 1000-3000 by user. Greater strength means a better chance to win a game.

##### Home Team Advantage

teams tend to be more successful when playing in front of their fans, this gives a winning chance bonus by ratio of 1.1

##### Form factor

This compares two playing teams ratios of how much points they gain/how many games they played. This helps to add a dynamicity to the estimations.

#### Home Win Chance

Based on these numerical values, a home team winning probability is calculated. 
There is a default 20% chance for a draw and otherwise the away team wins. 

#### Match Results

Match results are calculated based on the winning or draw result. Using math/rand random scores are assigned to each team based on the game results.

### Championship Estimations

To estimate a champion after week 4 the program simulates the remaining part of the league 10000 times to return championship numbers of each team after 10000 iterations. The team with highest number of championships after 10000 iterations has the highest estimation to be the champion.
Since the main factor in game results is the team strength the results can be a bit odd if the teams has a high gap between their strengths but it mostly guesses fine.



#### Additional Endpoint That may be useful for different cases

##### Get Teams by League ID - GET /teams/{leagueID}
```bash
curl -X GET http://localhost:8081/api/teams/17
```
```json
[
    {
        "id": 65,
        "league_id": 17,
        "name": "Fenerbahçe",
        "strength": 2000,
        "stats": {
            "team_id": 0,
            "points": 0,
            "played": 0,
            "won": 0,
            "lost": 0,
            "draw": 0,
            "goals_for": 0,
            "goals_against": 0,
            "goal_diff": 0,
            "estimation": 0
        }
    }...
]
```
##### Get Matches by League ID and week - GET /matches/{leagueID}/{week}

```bash
curl -X GET http://localhost:8081/api/matches/17/2
```
```json
[
    {
        "id": 195,
        "league_id": 17,
        "week": 2,
        "played": true,
        "home_team": 67,
        "away_team": 65,
        "home_score": 2,
        "away_score": 3
    },...
]
```

##### Get All Matches By League ID - GET /matches/{leagueID}
```bash
curl -X GET http://localhost:8081/api/matches/17
```
```json
[
    {
        "id": 201,
        "league_id": 17,
        "week": 5,
        "played": false,
        "home_team": 65,
        "away_team": 67,
        "home_score": 0,
        "away_score": 0
    }...
]

```

##### Get Championship Estimations by LeagueID - GET /leagues/championship-estimations?{leagueID}
```bash
curl -X GET http://localhost:8081/api/leagues/championship-estimations?leagueID=1
```
```json
[
    {
        "league_id": 17,
        "week": 5,
        "team_id": 65,
        "estimation": 0.0518
    },
    {
        "league_id": 17,
        "week": 5,
        "team_id": 66,
        "estimation": 0.8756
    }...
]

```





