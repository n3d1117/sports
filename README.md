# Sports

Reliable sports data for humans and agents.

- Current scope: football, basketball, tennis.
- Current provider: SofaScore.

Use it for schedules, standings, match stats, player pages, tournament pages, live updates and so on.

## Demo
Feed the skill to your agent of choice and you get instant access to all kinds of data:

| Examples |
| --- |
| ![demo](https://github.com/user-attachments/assets/217ffd12-bbd9-4599-bff4-07951ebaf334) |

[^demo]: Example output from an agent using the `sports` skill.

## Install

Tagged release:

```bash
brew tap n3d1117/sports https://github.com/n3d1117/sports
brew install n3d1117/sports/sports
```

Install the skill:

```bash
npx skills add n3d1117/sports --skill sports
```

## CLI

Use CLI help for flags:

```bash
sports --help
sports <command> --help
```

Core routes:

- `search <query>`: resolve ids
- `list`: list supported sports
- `<sport>`: flat day schedule
- `<sport> tournaments`: grouped day schedule
- `<sport> sections`: event sections for that sport
- `event <id>`: event snapshot
- `event <id> section <name>`: one event section
- `event <id> watch` follows live event updates
- `event <id> tv|tv-channel|h2h-events`: named event routes
- `team <id> ...`: schedules, standings, stats, rankings, roster, media
- `player <id> ...`: profile, seasons, career, season stats, sport-specific extras
- `tournament <id> ...`: metadata, sections, seasons, event lists
- `trending`: trending event feed

Examples:

```bash
sports search "Arsenal" --sport football --limit 1 --id
sports football --date 2026-04-04 --json
sports event 15636234 --json
sports event 15636234 section statistics --json
sports team 37 standings --json
sports tournament 17 seasons --json
```

Output notes:

- TTY stdout is text.
- Piped stdout is JSON.
- `--json` forces JSON or NDJSON.
- Exit codes: `2` invalid input, `3` upstream error, `4` not found.

## MCP Server

Run over stdio:

```bash
go run ./cmd/sports-mcp
```

Run over Streamable HTTP:

```bash
go run ./cmd/sports-mcp --http :8080
```

Help:

```bash
sports-mcp --help
```

Main MCP tool groups:

- routing: `search`, `sports`, `sports_events`, `sports_tournaments`, `trending`
- event: `event`, `event_tv`, `event_tv_channel`, `event_h2h_events`
- tournament: `tournament`, `tournament_seasons`, `tournament_events`, `tournament_scheduled_events`
- team: `team_events`, `team_info`, `team_tournaments`, `team_standings`, `team_stats`, `team_players`, `team_media`, `team_rankings`, `team_top_players`
- player: `player_attributes`, `player_media`, `player_media_videos`, `player_events_last`, `player_seasons`, `player_career`, `player_season_stats`, `player_season_ratings`, `player_characteristics`, `player_national_team_stats`, `player_tournaments`, `player_season_heatmap`, `player_penalty_history`, `player_shot_actions`, `player_shot_action_areas`, `player_year_stats`, `player_featured_event`
- sport feeds: `sport_live_tournaments`, `sport_categories`, `sport_top_players`

Live resources:

- `sports://live/event/{event_id}{?section,all_sections}`
- `sports://live/events/{event_ids}{?section,all_sections}`
- `sports://live/sport/{sport}`
- `sports://live/sports/{sports}`

Live resources are subscribe-first: subscribe, then read the same URI.

## Docs

- skill entry point: [`skills/sports/SKILL.md`](skills/sports/SKILL.md)
- shared guide: [`skills/sports/references/generic.md`](skills/sports/references/generic.md)
- football: [`skills/sports/references/sports/football.md`](skills/sports/references/sports/football.md)
- basketball: [`skills/sports/references/sports/basketball.md`](skills/sports/references/sports/basketball.md)
- tennis: [`skills/sports/references/sports/tennis.md`](skills/sports/references/sports/tennis.md)

## Development

```bash
go test ./...
go run ./cmd/sports --help
go run ./cmd/sports-mcp --help
```

## Disclaimer

This project is unofficial and independent. It uses SofaScore as the current data provider but is not affiliated with, endorsed by, or sponsored by SofaScore. SofaScore names, marks, data, and other rights belong to their respective owners.
