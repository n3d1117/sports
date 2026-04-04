# Football

Football only.

## Fast route

| Need | Run first | Reuse |
| --- | --- | --- |
| Club id | `sports search <club> --limit 1 --id` | `team_id` |
| Competition id | `sports search <competition> --type tournament --limit 1 --id` | `tournament_id` |
| Match id | `sports search "<home away>" --type event --limit 1 --id` | `event_id` |
| Day schedule | `sports football --json` | `events[]` |
| Grouped day schedule | `sports football tournaments --json` | `unique_tournament_id` |
| Team pages | `sports team <team-id> info --json` | related ids and fields |
| Team standings or season stats | `sports team <team-id> standings --json` or `stats --tournament <id> --season <id> --json` | matching rows and stats |
| Player page | `sports player <player-id> attributes|seasons|career --json` | payload fields |
| Event snapshot | `sports event <event-id> --json` | `available_sections[]` |
| Live feeds | `sports event <event-id> watch --json` or `sports football watch --json` | `summary`, `changed_fields`, `patch` |
| Sport feeds | `sports football live-tournaments|categories|top-players --json` | payload fields |

## Rules

- For season comparisons, start with team standings and team season stats.
- `player <id> attributes` is scouting-style profile data. It is not goals, assists, clean sheets, or season totals.
- Search affiliation text can reflect the present. Do not use it as proof of historical squad context.
- Event sections depend on state. One event never shows the whole football surface.
- If direct event search for a matchup fails, use one club's `team next|last` feed to find a meeting, then switch to `event <id> h2h-events`.
- Use explicit `--season` when the year matters.

## Common flows

Next match:

```bash
TEAM_ID=$(sports search fiorentina --limit 1 --id)
sports team "$TEAM_ID" next --limit 1 --json
```

Last result:

```bash
TEAM_ID=$(sports search fiorentina --limit 1 --id)
EVENT_ID=$(sports team "$TEAM_ID" last --limit 1 --json | jq -r '.events[0].event_id')
sports event "$EVENT_ID" --json
```

Season comparison surface:

```bash
sports team <team-id> standings --season <season-id> --json
sports team <team-id> stats --tournament <tournament-id> --season <season-id> --json
```

Goals and cards only:

```bash
sports event <event-id> section incidents --json \
  | jq -cr '.sections.incidents.incidents[]? | select(.incidentType == "goal" or .incidentType == "card") | {incidentType, time, player: (.player.name // null), text: (.text // null), homeScore, awayScore}'
```

Common post-match stats:

```bash
sports event <event-id> section statistics --json \
  | jq -cr '.sections.statistics.statistics[]? | select(.period == "ALL") | .groups[].statisticsItems[]? | select(.key == "ballPossession" or .key == "shotsOnGoal" or .key == "cornerKicks" or .key == "goalkeeperSaves") | {key, home, away}'
```

Live watch:

```bash
sports event <event-id> watch --json
sports event <event-id> watch --section statistics --section incidents --json
sports football watch --json
```

## Tournament sections

| Section | Use | Payload path |
| --- | --- | --- |
| `featured-events` | promoted events | `sections["featured-events"].featuredEvents[]` |
| `media` | tournament media metadata | `sections.media.media[]` |
| `info` | season facts and metadata | `sections.info.info` |
| `rounds` | current round and round list | `sections.rounds.currentRound`, `sections.rounds.rounds[]` |
| `standings/total` | main table | `sections["standings/total"].standings[].rows[]` |
| `cuptrees` | bracket or playoff tree | `sections.cuptrees.cupTrees[]` |
| `venues` | stadium list | `sections.venues.venues[]` |
| `player-statistics/types` | player leaderboard families | `sections["player-statistics/types"].types[]` |
| `team-statistics/types` | team leaderboard families | `sections["team-statistics/types"].types[]` |
| `player-of-the-season-race` | award or ranking race | `sections["player-of-the-season-race"].statisticsType`, `topPlayers[]` |
| `team-events/total` | competition-specific matchup history | `sections["team-events/total"].tournamentTeamEvents` |
| `team-of-the-week/periods` | team-of-the-week period ids | `sections["team-of-the-week/periods"].periods[]` |

## Event sections

| Section | Use | Payload path |
| --- | --- | --- |
| `statistics` | possession, shots, passes, fouls, duels | `sections.statistics.statistics[]` |
| `h2h` | rivalry summary and meetings | `sections.h2h.teamDuel`, `sections.h2h.managerDuel` |
| `incidents` | goals, cards, subs, score changes | `sections.incidents.incidents[]` |
| `lineups` | starters, bench, formations, missing players | `sections.lineups.confirmed`, `home`, `away` |
| `graph` | momentum or score-flow graph | `sections.graph.graphPoints[]` |
| `average-positions` | average player locations | `sections["average-positions"].home`, `away`, `substitutions` |
| `best-players` | best player card per side | `sections["best-players"].bestHomeTeamPlayer`, `bestAwayTeamPlayer` |
| `best-players/summary` | compact best-player summary | `sections["best-players/summary"].playerOfTheMatch`, `bestHomeTeamPlayers[]`, `bestAwayTeamPlayers[]` |
| `comments` | text feed | `sections.comments.comments[]` |
| `highlights` | highlight metadata and links | `sections.highlights.highlights[]` |
| `official-tweets` | social embeds | `sections["official-tweets"].tweets[]` |
| `team-streaks` | general and H2H streaks | `sections["team-streaks"].general`, `head2head` |
| `pregame-form` | pre-match form and ranking context | `sections["pregame-form"].homeTeam`, `awayTeam`, `label` |
| `managers` | manager cards | `sections.managers.homeManager`, `awayManager` |
| `votes` | crowd voting | `sections.votes.vote`, `sections.votes.firstTeamToScoreVote` |

## Match state

| State | Usually available |
| --- | --- |
| Upcoming | `h2h`, `lineups`, `pregame-form`, `team-streaks`, `managers`, `votes` |
| Live | `statistics`, `incidents`, `graph`, `average-positions`, `comments`, `best-players` |
| Finished | live sections plus `highlights`, `best-players/summary`, `official-tweets` |

## Gotchas

- `sports event <id> tv`, `tv-channel`, and `h2h-events` are named commands, not sections.
- Nested section names need the full path, for example `best-players/summary`.
- `team standings` can contain multiple tables. Keep the table name with the row when it matters.
- National-team qualifiers and mixed-format cups can blur group and knockout labels. `standings/total` is for tables; `cuptrees` is for knockout paths.
- If a football player season route returns `not found`, say that. Do not pretend `attributes` is a stats fallback.
