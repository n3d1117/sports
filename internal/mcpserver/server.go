package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sports/internal/buildinfo"
	"sports/internal/lookups"
)

type SearchInput struct {
	Query string `json:"query" jsonschema:"the free-text search query"`
	Type  string `json:"type,omitempty" jsonschema:"optional result type filter, for example team, event, or tournament"`
	Sport string `json:"sport,omitempty" jsonschema:"optional sport slug filter"`
	Page  int    `json:"page,omitempty" jsonschema:"optional search results page, starting at 0"`
	Limit int    `json:"limit,omitempty" jsonschema:"optional result limit; 0 means no extra limit"`
	ID    bool   `json:"id,omitempty" jsonschema:"when true, also return the matching ids in an ids field"`
}

type SportsInput struct {
	Sport    string `json:"sport,omitempty" jsonschema:"optional sport slug"`
	Sections bool   `json:"sections,omitempty" jsonschema:"when true, discover likely event sections for the given sport"`
}

type SportsEventsInput struct {
	Sport string `json:"sport" jsonschema:"the sport slug, for example football, tennis, or basketball"`
	Date  string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Limit int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type SportsTournamentsInput struct {
	Sport string `json:"sport" jsonschema:"the sport slug, for example football, tennis, or basketball"`
	Date  string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Page  int    `json:"page,omitempty" jsonschema:"optional 1-based page number; defaults to 1"`
}

type TrendingInput struct {
	Country string `json:"country,omitempty" jsonschema:"optional ISO alpha-2 country code; defaults to /country/alpha2"`
	Date    string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Limit   int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type EventInput struct {
	EventID      int      `json:"event_id" jsonschema:"the SofaScore event id"`
	SectionsOnly bool     `json:"sections_only,omitempty" jsonschema:"when true, return only the available section names for the event"`
	Sections     []string `json:"sections,omitempty" jsonschema:"optional event sections to fetch"`
}

type TournamentInput struct {
	TournamentID int      `json:"tournament_id" jsonschema:"the SofaScore tournament id"`
	SeasonID     int      `json:"season_id,omitempty" jsonschema:"optional season id"`
	SectionsOnly bool     `json:"sections_only,omitempty" jsonschema:"when true, return only the available section names for the tournament season"`
	Sections     []string `json:"sections,omitempty" jsonschema:"optional tournament sections to fetch"`
}

type TournamentSeasonsInput struct {
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore tournament id"`
}

type TournamentEventsInput struct {
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore tournament id"`
	SeasonID     int    `json:"season_id,omitempty" jsonschema:"optional season id"`
	Next         bool   `json:"next,omitempty" jsonschema:"when true, return upcoming events"`
	Last         bool   `json:"last,omitempty" jsonschema:"when true, return recent events"`
	Round        int    `json:"round,omitempty" jsonschema:"when greater than 0, return events for that round"`
	Slug         string `json:"slug,omitempty" jsonschema:"optional round slug; requires round"`
	Limit        int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type TournamentScheduledEventsInput struct {
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	Date         string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Limit        int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type TeamEventsInput struct {
	TeamID int  `json:"team_id" jsonschema:"the SofaScore team id or tennis player id"`
	Next   bool `json:"next,omitempty" jsonschema:"when true, return upcoming events"`
	Last   bool `json:"last,omitempty" jsonschema:"when true, return recent events"`
	Limit  int  `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type EventTVInput struct {
	EventID int `json:"event_id" jsonschema:"the SofaScore event id"`
}

type EventTVChannelInput struct {
	EventID   int `json:"event_id" jsonschema:"the SofaScore event id"`
	ChannelID int `json:"channel_id" jsonschema:"the SofaScore TV channel id"`
}

type EventH2HEventsInput struct {
	EventID int `json:"event_id" jsonschema:"the SofaScore event id"`
}

type TeamInfoInput struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamTournamentsInput struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamStandingsInput struct {
	TeamID   int `json:"team_id" jsonschema:"the SofaScore team id"`
	SeasonID int `json:"season_id,omitempty" jsonschema:"optional season id"`
}

type TeamStatsInput struct {
	TeamID       int `json:"team_id" jsonschema:"the SofaScore team id"`
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int `json:"season_id" jsonschema:"the SofaScore season id"`
}

type TeamPlayersInput struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamMediaInput struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamRankingsInput struct {
	TeamID       int `json:"team_id" jsonschema:"the SofaScore team id"`
	TournamentID int `json:"tournament_id,omitempty" jsonschema:"optional unique tournament id"`
	SeasonID     int `json:"season_id,omitempty" jsonschema:"optional season id"`
}

type TeamTopPlayersInput struct {
	TeamID       int `json:"team_id" jsonschema:"the SofaScore team id"`
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int `json:"season_id" jsonschema:"the SofaScore season id"`
}

type PlayerAttributesInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerMediaInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerMediaVideosInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerLastEventsInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
	Limit    int `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type PlayerSeasonsInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerCareerInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerSeasonStatsInput struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase,omitempty" jsonschema:"optional phase name; defaults to the sport-specific route default"`
}

type PlayerSeasonRatingsInput struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase,omitempty" jsonschema:"optional phase name; defaults to the sport-specific route default"`
}

type PlayerCharacteristicsInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerNationalTeamStatsInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerTournamentsInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerSeasonHeatmapInput struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase" jsonschema:"the phase name, for example overall"`
}

type PlayerPenaltyHistoryInput struct {
	PlayerID     int `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int `json:"season_id" jsonschema:"the SofaScore season id"`
}

type PlayerShotActionsInput struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase" jsonschema:"the phase name, for example regularSeason"`
}

type PlayerShotActionAreasInput struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase" jsonschema:"the phase name, for example regularSeason"`
}

type PlayerYearStatsInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
	Year     int `json:"year" jsonschema:"the calendar year, for example 2026"`
}

type PlayerFeaturedEventInput struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type SportLiveTournamentsInput struct {
	Sport string `json:"sport" jsonschema:"the sport slug, for example football or basketball"`
}

type SportCategoriesInput struct {
	Sport string `json:"sport" jsonschema:"the sport slug, for example football or basketball"`
}

type SportTopPlayersInput struct {
	Sport string `json:"sport" jsonschema:"the sport slug, for example football or basketball"`
}

func New(client lookups.Service) *mcp.Server {
	liveHub := newLiveHub(client)
	opts := &mcp.ServerOptions{
		SubscribeHandler:   liveHub.subscribe,
		UnsubscribeHandler: liveHub.unsubscribe,
	}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "sports",
		Version: buildinfo.Current(),
	}, opts)
	liveHub.server = server

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Search SofaScore teams, events, and tournaments.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, lookups.SearchResponse, error) {
		response, err := lookups.Search(ctx, client, lookups.SearchParams{
			Query:      input.Query,
			ResultType: input.Type,
			Sport:      input.Sport,
			Page:       input.Page,
			Limit:      input.Limit,
			IDOnly:     input.ID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sports",
		Description: "List sport slugs or discover likely event sections for one sport.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SportsInput) (*mcp.CallToolResult, lookups.SportsResponse, error) {
		response, err := lookups.Sports(ctx, client, lookups.SportsParams{
			Sport:    input.Sport,
			Sections: input.Sections,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sports_events",
		Description: "Fetch scheduled events for one sport on one UTC date.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SportsEventsInput) (*mcp.CallToolResult, lookups.SportEventsResponse, error) {
		response, err := lookups.SportEvents(ctx, client, lookups.SportEventsParams{
			Sport: input.Sport,
			Date:  input.Date,
			Limit: input.Limit,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sports_tournaments",
		Description: "Fetch scheduled tournament groups for one sport, date, and page.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SportsTournamentsInput) (*mcp.CallToolResult, lookups.SportScheduledTournamentsResponse, error) {
		response, err := lookups.SportScheduledTournaments(ctx, client, lookups.SportScheduledTournamentsParams{
			Sport: input.Sport,
			Date:  input.Date,
			Page:  input.Page,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "trending",
		Description: "Fetch trending events for a country and UTC date.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TrendingInput) (*mcp.CallToolResult, lookups.TrendingResponse, error) {
		response, err := lookups.Trending(ctx, client, lookups.TrendingParams{
			Country: input.Country,
			Date:    input.Date,
			Limit:   input.Limit,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "event",
		Description: "Fetch one event and optional event sections.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input EventInput) (*mcp.CallToolResult, lookups.EventResponse, error) {
		response, err := lookups.Event(ctx, client, lookups.EventParams{
			EventID:                   input.EventID,
			SectionsOnly:              input.SectionsOnly,
			Sections:                  input.Sections,
			AllowPartialSectionErrors: true,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "event_tv",
		Description: "Fetch TV channel ids by country for one event.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input EventTVInput) (*mcp.CallToolResult, lookups.EventTVResponse, error) {
		response, err := lookups.EventTV(ctx, client, lookups.EventTVParams{
			EventID: input.EventID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "event_tv_channel",
		Description: "Resolve one event-specific TV channel id to channel details and votes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input EventTVChannelInput) (*mcp.CallToolResult, lookups.EventTVChannelResponse, error) {
		response, err := lookups.EventTVChannel(ctx, client, lookups.EventTVChannelParams{
			EventID:   input.EventID,
			ChannelID: input.ChannelID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "event_h2h_events",
		Description: "Fetch the H2H event list feed for one event via its customId.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input EventH2HEventsInput) (*mcp.CallToolResult, lookups.EventH2HEventsResponse, error) {
		response, err := lookups.EventH2HEvents(ctx, client, lookups.EventH2HEventsParams{
			EventID: input.EventID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "tournament",
		Description: "Fetch tournament metadata and optional tournament sections.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TournamentInput) (*mcp.CallToolResult, lookups.TournamentResponse, error) {
		response, err := lookups.Tournament(ctx, client, lookups.TournamentParams{
			TournamentID:              input.TournamentID,
			SeasonID:                  input.SeasonID,
			SectionsOnly:              input.SectionsOnly,
			Sections:                  input.Sections,
			AllowPartialSectionErrors: true,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "tournament_seasons",
		Description: "List seasons for one tournament.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TournamentSeasonsInput) (*mcp.CallToolResult, lookups.TournamentSeasonsResponse, error) {
		response, err := lookups.TournamentSeasons(ctx, client, lookups.TournamentSeasonsParams{
			TournamentID: input.TournamentID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "tournament_events",
		Description: "Fetch next, last, or round events for a tournament season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TournamentEventsInput) (*mcp.CallToolResult, lookups.TournamentEventsResponse, error) {
		response, err := lookups.TournamentEvents(ctx, client, lookups.TournamentEventsParams{
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
			Next:         input.Next,
			Last:         input.Last,
			Round:        input.Round,
			Slug:         input.Slug,
			Limit:        input.Limit,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "tournament_scheduled_events",
		Description: "Fetch scheduled events for one unique tournament on one UTC date.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TournamentScheduledEventsInput) (*mcp.CallToolResult, lookups.TournamentScheduledEventsResponse, error) {
		response, err := lookups.TournamentScheduledEvents(ctx, client, lookups.TournamentScheduledEventsParams{
			TournamentID: input.TournamentID,
			Date:         input.Date,
			Limit:        input.Limit,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_events",
		Description: "Fetch next or last events for a team or tennis player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamEventsInput) (*mcp.CallToolResult, lookups.TeamEventsResponse, error) {
		response, err := lookups.TeamEvents(ctx, client, lookups.TeamEventsParams{
			TeamID: input.TeamID,
			Next:   input.Next,
			Last:   input.Last,
			Limit:  input.Limit,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_info",
		Description: "Fetch team metadata, featured event, performance, and tournaments.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamInfoInput) (*mcp.CallToolResult, lookups.TeamInfoResponse, error) {
		response, err := lookups.TeamInfo(ctx, client, lookups.TeamInfoParams{
			TeamID: input.TeamID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_tournaments",
		Description: "Fetch the unique tournaments for one team.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamTournamentsInput) (*mcp.CallToolResult, lookups.TeamTournamentsResponse, error) {
		response, err := lookups.TeamTournaments(ctx, client, lookups.TeamTournamentsParams{
			TeamID: input.TeamID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_standings",
		Description: "Fetch standings for one team, optionally for a specific season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamStandingsInput) (*mcp.CallToolResult, lookups.TeamStandingsResponse, error) {
		response, err := lookups.TeamStandings(ctx, client, lookups.TeamStandingsParams{
			TeamID:   input.TeamID,
			SeasonID: input.SeasonID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_stats",
		Description: "Fetch overall statistics for one team in one tournament season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamStatsInput) (*mcp.CallToolResult, lookups.TeamStatsResponse, error) {
		response, err := lookups.TeamStats(ctx, client, lookups.TeamStatsParams{
			TeamID:       input.TeamID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_players",
		Description: "Fetch featured players for one team.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamPlayersInput) (*mcp.CallToolResult, lookups.TeamPlayersResponse, error) {
		response, err := lookups.TeamPlayers(ctx, client, lookups.TeamPlayersParams{
			TeamID: input.TeamID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_media",
		Description: "Fetch media videos for one team.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamMediaInput) (*mcp.CallToolResult, lookups.TeamMediaResponse, error) {
		response, err := lookups.TeamMedia(ctx, client, lookups.TeamMediaParams{
			TeamID: input.TeamID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_rankings",
		Description: "Fetch direct team rankings or tournament-season ranks for one team.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamRankingsInput) (*mcp.CallToolResult, lookups.TeamRankingsResponse, error) {
		response, err := lookups.TeamRankings(ctx, client, lookups.TeamRankingsParams{
			TeamID:       input.TeamID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "team_top_players",
		Description: "Fetch tournament-season top players for one team.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TeamTopPlayersInput) (*mcp.CallToolResult, lookups.TeamTopPlayersResponse, error) {
		response, err := lookups.TeamTopPlayers(ctx, client, lookups.TeamTopPlayersParams{
			TeamID:       input.TeamID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_attributes",
		Description: "Fetch attribute overviews for one player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerAttributesInput) (*mcp.CallToolResult, lookups.PlayerAttributesResponse, error) {
		response, err := lookups.PlayerAttributes(ctx, client, lookups.PlayerAttributesParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_media",
		Description: "Fetch media entries for one player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerMediaInput) (*mcp.CallToolResult, lookups.PlayerMediaResponse, error) {
		response, err := lookups.PlayerMedia(ctx, client, lookups.PlayerMediaParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_media_videos",
		Description: "Fetch media videos for one player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerMediaVideosInput) (*mcp.CallToolResult, lookups.PlayerMediaVideosResponse, error) {
		response, err := lookups.PlayerMediaVideos(ctx, client, lookups.PlayerMediaVideosParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_events_last",
		Description: "Fetch recent events for one player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerLastEventsInput) (*mcp.CallToolResult, lookups.PlayerLastEventsResponse, error) {
		response, err := lookups.PlayerLastEvents(ctx, client, lookups.PlayerLastEventsParams{
			PlayerID: input.PlayerID,
			Limit:    input.Limit,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_seasons",
		Description: "Fetch available statistics seasons for one player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerSeasonsInput) (*mcp.CallToolResult, lookups.PlayerSeasonsResponse, error) {
		response, err := lookups.PlayerSeasons(ctx, client, lookups.PlayerSeasonsParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_career",
		Description: "Fetch career statistics for one player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerCareerInput) (*mcp.CallToolResult, lookups.PlayerCareerResponse, error) {
		response, err := lookups.PlayerCareer(ctx, client, lookups.PlayerCareerParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_season_stats",
		Description: "Fetch season statistics for one player in one tournament season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerSeasonStatsInput) (*mcp.CallToolResult, lookups.PlayerSeasonStatsResponse, error) {
		response, err := lookups.PlayerSeasonStats(ctx, client, lookups.PlayerSeasonStatsParams{
			PlayerID:     input.PlayerID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
			Phase:        input.Phase,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_season_ratings",
		Description: "Fetch season ratings for one player in one tournament season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerSeasonRatingsInput) (*mcp.CallToolResult, lookups.PlayerSeasonRatingsResponse, error) {
		response, err := lookups.PlayerSeasonRatings(ctx, client, lookups.PlayerSeasonRatingsParams{
			PlayerID:     input.PlayerID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
			Phase:        input.Phase,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_characteristics",
		Description: "Fetch characteristics for one football player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerCharacteristicsInput) (*mcp.CallToolResult, lookups.PlayerCharacteristicsResponse, error) {
		response, err := lookups.PlayerCharacteristics(ctx, client, lookups.PlayerCharacteristicsParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_national_team_stats",
		Description: "Fetch national-team statistics for one football player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerNationalTeamStatsInput) (*mcp.CallToolResult, lookups.PlayerNationalTeamStatsResponse, error) {
		response, err := lookups.PlayerNationalTeamStats(ctx, client, lookups.PlayerNationalTeamStatsParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_tournaments",
		Description: "Fetch unique tournaments for one football player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerTournamentsInput) (*mcp.CallToolResult, lookups.PlayerTournamentsResponse, error) {
		response, err := lookups.PlayerTournaments(ctx, client, lookups.PlayerTournamentsParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_season_heatmap",
		Description: "Fetch season heatmap data for one football player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerSeasonHeatmapInput) (*mcp.CallToolResult, lookups.PlayerSeasonHeatmapResponse, error) {
		response, err := lookups.PlayerSeasonHeatmap(ctx, client, lookups.PlayerSeasonHeatmapParams{
			PlayerID:     input.PlayerID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
			Phase:        input.Phase,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_penalty_history",
		Description: "Fetch penalty history for one football player in one tournament season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerPenaltyHistoryInput) (*mcp.CallToolResult, lookups.PlayerPenaltyHistoryResponse, error) {
		response, err := lookups.PlayerPenaltyHistory(ctx, client, lookups.PlayerPenaltyHistoryParams{
			PlayerID:     input.PlayerID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_shot_actions",
		Description: "Fetch shot actions for one basketball player in one tournament season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerShotActionsInput) (*mcp.CallToolResult, lookups.PlayerShotActionsResponse, error) {
		response, err := lookups.PlayerShotActions(ctx, client, lookups.PlayerShotActionsParams{
			PlayerID:     input.PlayerID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
			Phase:        input.Phase,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_shot_action_areas",
		Description: "Fetch shot-action area data for one basketball player in one tournament season.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerShotActionAreasInput) (*mcp.CallToolResult, lookups.PlayerShotActionAreasResponse, error) {
		response, err := lookups.PlayerShotActionAreas(ctx, client, lookups.PlayerShotActionAreasParams{
			PlayerID:     input.PlayerID,
			TournamentID: input.TournamentID,
			SeasonID:     input.SeasonID,
			Phase:        input.Phase,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_year_stats",
		Description: "Fetch year statistics for one tennis player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerYearStatsInput) (*mcp.CallToolResult, lookups.PlayerYearStatsResponse, error) {
		response, err := lookups.PlayerYearStats(ctx, client, lookups.PlayerYearStatsParams{
			PlayerID: input.PlayerID,
			Year:     input.Year,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "player_featured_event",
		Description: "Fetch the featured event for one tennis player.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input PlayerFeaturedEventInput) (*mcp.CallToolResult, lookups.PlayerFeaturedEventResponse, error) {
		response, err := lookups.PlayerFeaturedEvent(ctx, client, lookups.PlayerFeaturedEventParams{
			PlayerID: input.PlayerID,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sport_live_tournaments",
		Description: "Fetch live tournaments for one sport.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SportLiveTournamentsInput) (*mcp.CallToolResult, lookups.SportLiveTournamentsResponse, error) {
		response, err := lookups.SportLiveTournaments(ctx, client, lookups.SportLiveTournamentsParams{
			Sport: input.Sport,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sport_categories",
		Description: "Fetch categories for one sport.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SportCategoriesInput) (*mcp.CallToolResult, lookups.SportCategoriesResponse, error) {
		response, err := lookups.SportCategories(ctx, client, lookups.SportCategoriesParams{
			Sport: input.Sport,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sport_top_players",
		Description: "Fetch trending top players for one sport.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SportTopPlayersInput) (*mcp.CallToolResult, lookups.SportTopPlayersResponse, error) {
		response, err := lookups.SportTopPlayers(ctx, client, lookups.SportTopPlayersParams{
			Sport: input.Sport,
		})
		return nil, response, err
	})

	registerLiveResources(server, liveHub)

	return server
}
