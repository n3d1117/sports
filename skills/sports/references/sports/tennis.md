# Tennis

Tennis only.

## Fast route

| Need | Run first | Reuse |
| --- | --- | --- |
| Player id | `sports search "<player>" --sport tennis --limit 1 --id` | `player_id` |
| Tournament id | `sports search "<tournament>" --sport tennis --type tournament --limit 1 --id` | `tournament_id` |
| Match id | `sports search "<player1 player2>" --sport tennis --type event --limit 1 --id` | `event_id` |
| Day schedule | `sports tennis --json` | `events[]` |
| Grouped day schedule | `sports tennis tournaments --json` | `unique_tournament_id` |
| Player pages | `sports player <player-id> seasons|year-stats|featured-event --json` | payload fields |
| Recent or upcoming matches | `sports team <player-id> last|next --json` | `events[]` |
| Rankings or tournament stats | `sports team <player-id> rankings --json` or `stats --tournament <id> --season <id> --json` | returned rows |
| Event snapshot | `sports event <event-id> --json` | `available_sections[]` |
| Live feeds | `sports event <event-id> watch --json` or `sports tennis watch --json` | `summary`, `changed_fields`, `patch` |

## Rules

- Tennis uses a mixed surface. Use `team <player-id> last|next` for match lists.
- Do not trust `player <player-id> last` for tennis. Some ids collide with non-tennis player routes and can return the wrong sport without a 404.
- Rankings and tournament stats still live under the `team` command family.
- Search can return tennis players as `type=team`. That is normal here.
- If direct event search for a matchup fails, stop broadening the query and use one player's recent matches to find a meeting.
- Recent feeds can mix singles and doubles. Filter to singles when the user asks about singles form.
- `tournament scheduled` does not take `--season`.
- Future draws can show placeholder entrants. Do not oversell them as fixed opponents.

## Common flows

Next match:

```bash
PLAYER_ID=$(sports search "Jannik Sinner" --sport tennis --limit 1 --id)
sports team "$PLAYER_ID" next --limit 1 --json
```

Recent matches:

```bash
PLAYER_ID=$(sports search "Jannik Sinner" --sport tennis --limit 1 --id)
sports team "$PLAYER_ID" last --limit 5 --json
```

Between tournaments fallback:

```bash
TOURNAMENT_ID=$(sports search "Monte Carlo" --sport tennis --type tournament --limit 1 --id)
sports tournament "$TOURNAMENT_ID" scheduled --date 2026-04-06 --limit 5 --json
```

Last result:

```bash
PLAYER_ID=$(sports search "Jannik Sinner" --sport tennis --limit 1 --id)
EVENT_ID=$(sports team "$PLAYER_ID" last --limit 1 --json | jq -r '.events[0].event_id')
sports event "$EVENT_ID" --json
```

H2H fallback when direct event search is empty:

```bash
PLAYER_ID=$(sports search "<player-a>" --sport tennis --limit 1 --id)
EVENT_ID=$(sports team "$PLAYER_ID" last --limit 25 --json | jq -r '.events[] | select((.home.name == "<player-a>" and .away.name == "<player-b>") or (.home.name == "<player-b>" and .away.name == "<player-a>")) | .event_id' | head -n 1)
sports event "$EVENT_ID" h2h-events --json
```

Rankings, tournament stats, TV, and H2H:

```bash
sports team <player-id> rankings --json
sports team <player-id> stats --tournament <id> --season <id> --json
sports event <event-id> tv --json
sports event <event-id> tv-channel <channel-id> --json
sports event <event-id> h2h-events --json
```

Live watch:

```bash
sports event <event-id> watch --json
sports event <event-id> watch --section point-by-point --json
sports tennis watch --json
```

## Tournament sections

| Section | Use | Payload path |
| --- | --- | --- |
| `featured-events` | promoted matches | `sections["featured-events"].featuredEvents[]` |
| `media` | tournament media metadata | `sections.media.media[]` |
| `info` | host city, prize money, competitor counts | `sections.info.info` |
| `rounds` | current round and round list | `sections.rounds.currentRound`, `sections.rounds.rounds[]` |
| `cuptrees` | bracket tree | `sections.cuptrees.cupTrees[]` |
| `venues` | venue list | `sections.venues.venues[]` |
| `team-statistics/types` | tournament stat families | `sections["team-statistics/types"].types[]` |

## Event sections

| Section | Use | Payload path |
| --- | --- | --- |
| `statistics` | serve, return, break-point, winners summary | `sections.statistics.statistics[]` |
| `h2h` | rivalry summary and previous meetings | `sections.h2h.teamDuel`, `sections.h2h.managerDuel` |
| `highlights` | highlight metadata and links | `sections.highlights.highlights[]` |
| `team-streaks` | general and H2H streaks | `sections["team-streaks"].general`, `head2head` |
| `tennis-power` | tennis-specific comparison view | `sections["tennis-power"].tennisPowerRankings[]` |
| `point-by-point` | set, game, and point flow | `sections["point-by-point"].pointByPoint[]` |
| `votes` | crowd voting | `sections.votes.vote`, `sections.votes.firstTeamToScoreVote` |

## Match state

| State | Usually available |
| --- | --- |
| Upcoming | `h2h`, `team-streaks`, `tennis-power`, `point-by-point`, `votes` |
| Live | `statistics`, `point-by-point`, `team-streaks`, `tennis-power` |
| Finished | live sections plus `highlights` |

## Gotchas

- `sports event <id> tv`, `tv-channel`, and `h2h-events` are named commands, not sections.
- `sports tournament <id> next` can return `resource not found`. Use `last`, `round`, or `scheduled` instead.
- `team <player-id> next` can return an empty `events` array between tournaments.
- Do not expect football-style sections like `lineups`, `incidents`, `pregame-form`, or `official-tweets`.
- If the user asks for recent form, filter out doubles unless they explicitly want mixed recent results.
