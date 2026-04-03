# Basketball

Basketball only.

## Fast route

| Need | Run first | Reuse |
| --- | --- | --- |
| Team id | `sports search <team> --sport basketball --limit 1 --id` | `team_id` |
| Competition id | `sports search <competition> --sport basketball --type tournament --limit 1 --id` | `tournament_id` |
| Game id | `sports search "<home away>" --sport basketball --type event --limit 1 --id` | `event_id` |
| Day schedule | `sports basketball --json` | `events[]` |
| Grouped day schedule | `sports basketball tournaments --json` | `unique_tournament_id` |
| Team pages | `sports team <team-id> info --json` | related ids and fields |
| Rankings or top players | `sports team <team-id> rankings|top-players --tournament <id> --season <id> --json` | returned rows |
| Player page | `sports player <player-id> last|seasons|career --json` | payload fields |
| Event snapshot | `sports event <event-id> --json` | `available_sections[]` |
| Live feeds | `sports event <event-id> watch --json` or `sports basketball watch --json` | `summary`, `changed_fields`, `patch` |
| Sport feeds | `sports basketball live-tournaments|categories|top-players --json` | payload fields |

## Rules

- `rankings` and `top-players` usually need both `--tournament` and `--season`.
- `season-stats` and `season-ratings` can omit `--phase`; `shot-actions` and `shot-action-areas` should pass `--phase regularSeason`.
- `attributes` is not a safe basketball route here. Do not assume it exists because football has it.
- `not found` means the route is not exposed for that player-season pair. Do not swap in team or event data without saying so.
- Standings payloads can include conference, division, and league-wide tables. Keep the table name.
- Basketball lineups are roster-style, not formation-style.

## Common flows

Next game:

```bash
TEAM_ID=$(sports search "virtus bologna" --sport basketball --limit 1 --id)
sports team "$TEAM_ID" next --limit 1 --json
```

Last result:

```bash
TEAM_ID=$(sports search "virtus bologna" --sport basketball --limit 1 --id)
EVENT_ID=$(sports team "$TEAM_ID" last --limit 1 --json | jq -r '.events[0].event_id')
sports event "$EVENT_ID" --json
```

Team pages:

```bash
sports team <team-id> info --json
sports team <team-id> standings --json
sports team <team-id> stats --tournament <id> --season <id> --json
sports team <team-id> rankings --tournament <id> --season <id> --json
sports team <team-id> top-players --tournament <id> --season <id> --json
```

Postgame best player plus team stats:

```bash
sports event <event-id> section best-players --json \
  | jq -cr '.sections["best-players"] | [(.bestHomeTeamPlayer // null), (.bestAwayTeamPlayer // null)] | map(select(. != null)) | max_by(.additionalStatistics.points // 0) | {player: .player.name, points: .additionalStatistics.points, assists: .additionalStatistics.assists, rebounds: .additionalStatistics.rebounds, side: .label}'
```

```bash
sports event <event-id> section statistics --json \
  | jq -cr '.sections.statistics.statistics[]? | select(.period == "ALL") | .groups[].statisticsItems[]? | select(.key == "rebounds" or .key == "assists" or .key == "turnovers" or .key == "blocks" or .key == "fieldGoals" or .key == "threePointFieldGoals") | {key, home, away}'
```

Live watch:

```bash
sports event <event-id> watch --json
sports basketball watch --json
```

## Tournament sections

| Section | Use | Payload path |
| --- | --- | --- |
| `featured-events` | promoted events | `sections["featured-events"].featuredEvents[]` |
| `media` | tournament media metadata | `sections.media.media[]` |
| `player-news` | tournament player news | `sections["player-news"].playerNews[]` |
| `info` | season facts and metadata | `sections.info.info` |
| `standings/total` | standings table | `sections["standings/total"].standings[].rows[]` |
| `venues` | arena list | `sections.venues.venues[]` |
| `player-statistics/types` | player leaderboard families | `sections["player-statistics/types"].types[]` |
| `team-statistics/types` | team leaderboard families | `sections["team-statistics/types"].types[]` |
| `team-events/total` | competition-specific matchup history | `sections["team-events/total"].tournamentTeamEvents` |
| `draft` | draft metadata | `sections.draft` |
| `team-of-the-week/periods` | team-of-the-week periods | `sections["team-of-the-week/periods"].periods[]` |

## Event sections

| Section | Use | Payload path |
| --- | --- | --- |
| `statistics` | quarter splits, shooting, rebounds, assists, turnovers | `sections.statistics.statistics[]` |
| `h2h` | matchup summary and meetings | `sections.h2h.teamDuel`, `sections.h2h.managerDuel` |
| `incidents` | scoring timeline and period markers | `sections.incidents.incidents[]` |
| `lineups` | roster lists and stat lines | `sections.lineups.confirmed`, `home`, `away` |
| `graph` | score progression graph | `sections.graph.graphPoints[]` |
| `best-players` | top player cards | `sections["best-players"].bestHomeTeamPlayer`, `bestAwayTeamPlayer` |
| `comments` | text feed | `sections.comments.comments[]` |
| `highlights` | highlight metadata and links | `sections.highlights.highlights[]` |
| `team-streaks` | general and H2H streaks | `sections["team-streaks"].general`, `head2head` |
| `pregame-form` | form and seeding context | `sections["pregame-form"].homeTeam`, `awayTeam`, `label` |
| `managers` | coach cards | `sections.managers.homeManager`, `awayManager` |
| `votes` | crowd voting | `sections.votes.vote`, `sections.votes.firstTeamToScoreVote` |

## Match state

| State | Usually available |
| --- | --- |
| Upcoming | `h2h`, `lineups`, `pregame-form`, `team-streaks`, `managers`, `votes` |
| Live | `statistics`, `incidents`, `lineups`, `graph` |
| Finished | live sections plus `best-players`, `comments`, `highlights` |

## Gotchas

- `sports event <id> tv`, `tv-channel`, and `h2h-events` are named commands, not sections.
- `team last` can include games not fully finished. Check `status.type`.
- `standings/total` is often conference-based or group-based, not one flat league table.
- `shot-actions` is the raw shot map. `shot-action-areas` is zone buckets and percentages. They are not interchangeable.
- When you answer with shot zones in natural language, translate the numeric area ids into court zones. Raw ids are useless to humans.
