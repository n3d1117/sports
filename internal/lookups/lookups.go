package lookups

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sports/internal/provider/sofascore"
)

var now = time.Now

type Service interface {
	Search(context.Context, string, int) ([]sofascoreapi.SearchResult, error)
	TeamEvents(context.Context, int, string, int) ([]sofascoreapi.EventSummary, error)
	Event(context.Context, int) (sofascoreapi.EventDetail, error)
	EventTVChannels(context.Context, int) (json.RawMessage, error)
	EventTVChannelVotes(context.Context, int, int) (json.RawMessage, error)
	EventH2HEvents(context.Context, int) (sofascoreapi.EventH2HEventsResult, error)
	SportEvents(context.Context, string, string, int) ([]sofascoreapi.EventSummary, error)
	SportScheduledTournaments(context.Context, string, string, int) ([]sofascoreapi.ScheduledTournamentSummary, bool, error)
	DetectCountryAlpha2(context.Context) (string, error)
	TeamInfo(context.Context, int) (sofascoreapi.TeamInfoResult, error)
	TeamTournaments(context.Context, int) (json.RawMessage, error)
	TeamStandings(context.Context, int, int) (sofascoreapi.TeamStandingsResult, error)
	TeamTournamentStatistics(context.Context, int, int, int) (json.RawMessage, error)
	TeamFeaturedPlayers(context.Context, int) (json.RawMessage, error)
	TeamMediaVideos(context.Context, int) (json.RawMessage, error)
	TeamRankings(context.Context, int) (json.RawMessage, error)
	TeamTournamentRanks(context.Context, int, int, int) (json.RawMessage, error)
	TeamTournamentTopPlayers(context.Context, int, int, int) (json.RawMessage, error)
	PlayerAttributeOverviews(context.Context, int) (json.RawMessage, error)
	PlayerCharacteristics(context.Context, int) (json.RawMessage, error)
	PlayerNationalTeamStatistics(context.Context, int) (json.RawMessage, error)
	PlayerMedia(context.Context, int) (json.RawMessage, error)
	PlayerMediaVideos(context.Context, int) (json.RawMessage, error)
	PlayerUniqueTournaments(context.Context, int) (json.RawMessage, error)
	PlayerStatisticsSeasons(context.Context, int) (json.RawMessage, error)
	PlayerSeasonStatistics(context.Context, int, int, int, string) (json.RawMessage, error)
	PlayerSeasonRatings(context.Context, int, int, int, string) (json.RawMessage, error)
	PlayerSeasonHeatmap(context.Context, int, int, int, string) (json.RawMessage, error)
	PlayerPenaltyHistory(context.Context, int, int, int) (json.RawMessage, error)
	PlayerCareerStatistics(context.Context, int) (json.RawMessage, error)
	PlayerCareerStatisticsMatchType(context.Context, int, string) (json.RawMessage, error)
	PlayerShotActions(context.Context, int, int, int, string) (json.RawMessage, error)
	PlayerShotActionAreas(context.Context, int, int, string) (json.RawMessage, error)
	PlayerLastEvents(context.Context, int, int) ([]sofascoreapi.EventSummary, error)
	PlayerFeaturedEvent(context.Context, int) (json.RawMessage, error)
	PlayerStatisticsSeasonsTennis(context.Context, int) (json.RawMessage, error)
	PlayerYearStatistics(context.Context, int, int) (json.RawMessage, error)
	SportLiveTournaments(context.Context, string) (json.RawMessage, error)
	SportCategories(context.Context, string) (json.RawMessage, error)
	SportTrendingTopPlayers(context.Context, string) (json.RawMessage, error)
	TournamentScheduledEvents(context.Context, int, string, int) ([]sofascoreapi.EventSummary, error)
	TrendingEvents(context.Context, string, int) ([]sofascoreapi.TrendingEventSummary, error)
	WatchEvents(context.Context, []int, []string, bool, func(sofascoreapi.WatchRecord) error) error
	WatchSports(context.Context, []string, func(sofascoreapi.WatchRecord) error) error
	EventSection(context.Context, int, string) (json.RawMessage, error)
	ProbeEventSections(context.Context, int) ([]string, error)
	Sports(context.Context) ([]sofascoreapi.SportCount, error)
	SportSections(context.Context, string) (sofascoreapi.SportSectionDiscovery, error)
	Tournament(context.Context, int) (sofascoreapi.TournamentDetail, error)
	TournamentSeasons(context.Context, int) ([]sofascoreapi.TournamentSeason, error)
	TournamentSection(context.Context, int, int, string) (json.RawMessage, error)
	ProbeTournamentSections(context.Context, int, int) ([]string, error)
	TournamentEvents(context.Context, int, int, string, int, string, int) ([]sofascoreapi.EventSummary, error)
}

type SearchParams struct {
	Query      string `json:"query" jsonschema:"the free-text search query"`
	ResultType string `json:"type,omitempty" jsonschema:"optional SofaScore result type filter, for example team, event, or tournament"`
	Sport      string `json:"sport,omitempty" jsonschema:"optional sport slug filter, for example football or tennis"`
	Page       int    `json:"page,omitempty" jsonschema:"optional search results page, starting at 0"`
	Limit      int    `json:"limit,omitempty" jsonschema:"optional result limit; 0 means no extra limit"`
	IDOnly     bool   `json:"id,omitempty" jsonschema:"when true, also return the matching ids in an ids field"`
}

type SearchResponse struct {
	OK      bool                        `json:"ok"`
	Query   string                      `json:"query"`
	Page    int                         `json:"page"`
	Results []sofascoreapi.SearchResult `json:"results"`
	IDs     []int                       `json:"ids,omitempty"`
}

type SportsParams struct {
	Sport    string `json:"sport,omitempty" jsonschema:"optional sport slug"`
	Sections bool   `json:"sections,omitempty" jsonschema:"when true, discover likely event sections for the given sport"`
}

type SportsResponse struct {
	OK            bool                      `json:"ok"`
	Sport         string                    `json:"sport,omitempty"`
	SampleEventID int                       `json:"sample_event_id,omitempty"`
	Sections      []string                  `json:"sections,omitempty"`
	Sports        []sofascoreapi.SportCount `json:"sports,omitempty"`
}

type SportEventsParams struct {
	Sport string `json:"sport" jsonschema:"the sport slug, for example football, tennis, or basketball"`
	Date  string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Limit int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type SportEventsResponse struct {
	OK     bool                        `json:"ok"`
	Sport  string                      `json:"sport"`
	Date   string                      `json:"date"`
	Events []sofascoreapi.EventSummary `json:"events"`
}

type SportScheduledTournamentsParams struct {
	Sport string `json:"sport" jsonschema:"the sport slug, for example football, tennis, or basketball"`
	Date  string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Page  int    `json:"page,omitempty" jsonschema:"optional 1-based page number; defaults to 1"`
}

type SportScheduledTournamentsResponse struct {
	OK          bool                                      `json:"ok"`
	Sport       string                                    `json:"sport"`
	Date        string                                    `json:"date"`
	Page        int                                       `json:"page"`
	HasNextPage bool                                      `json:"has_next_page"`
	Tournaments []sofascoreapi.ScheduledTournamentSummary `json:"tournaments"`
}

type TrendingParams struct {
	Country string `json:"country,omitempty" jsonschema:"optional ISO alpha-2 country code; defaults to /country/alpha2"`
	Date    string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Limit   int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type TrendingResponse struct {
	OK      bool                                `json:"ok"`
	Country string                              `json:"country"`
	Date    string                              `json:"date"`
	Events  []sofascoreapi.TrendingEventSummary `json:"events"`
}

type EventParams struct {
	EventID                   int      `json:"event_id" jsonschema:"the SofaScore event id"`
	SectionsOnly              bool     `json:"sections_only,omitempty" jsonschema:"when true, return only the available section names for the event"`
	Sections                  []string `json:"sections,omitempty" jsonschema:"optional event sections to fetch"`
	AllowPartialSectionErrors bool     `json:"-" jsonschema:"-"`
}

type EventResponse struct {
	OK                bool              `json:"ok"`
	Partial           bool              `json:"partial,omitempty"`
	EventID           int               `json:"event_id"`
	StartTime         string            `json:"start_time,omitempty"`
	StatusType        string            `json:"status_type,omitempty"`
	StatusDescription string            `json:"status_description,omitempty"`
	Home              string            `json:"home,omitempty"`
	Away              string            `json:"away,omitempty"`
	Tournament        string            `json:"tournament,omitempty"`
	Sport             string            `json:"sport"`
	Venue             string            `json:"venue,omitempty"`
	HomeScore         *int              `json:"home_score,omitempty"`
	AwayScore         *int              `json:"away_score,omitempty"`
	AvailableSections []string          `json:"available_sections"`
	Event             any               `json:"event,omitempty"`
	Sections          map[string]any    `json:"sections,omitempty"`
	SectionErrors     map[string]string `json:"section_errors,omitempty"`
}

type TournamentParams struct {
	TournamentID              int      `json:"tournament_id" jsonschema:"the SofaScore tournament id"`
	SeasonID                  int      `json:"season_id,omitempty" jsonschema:"optional season id"`
	SectionsOnly              bool     `json:"sections_only,omitempty" jsonschema:"when true, return only the available section names for the tournament season"`
	Sections                  []string `json:"sections,omitempty" jsonschema:"optional tournament sections to fetch"`
	AllowPartialSectionErrors bool     `json:"-" jsonschema:"-"`
}

type TournamentResponse struct {
	OK                bool                            `json:"ok"`
	Partial           bool                            `json:"partial,omitempty"`
	TournamentID      int                             `json:"tournament_id"`
	Name              string                          `json:"name,omitempty"`
	Sport             string                          `json:"sport"`
	Category          string                          `json:"category,omitempty"`
	Country           string                          `json:"country,omitempty"`
	SeasonID          int                             `json:"season_id"`
	SeasonName        string                          `json:"season_name,omitempty"`
	AvailableSections []string                        `json:"available_sections"`
	Seasons           []sofascoreapi.TournamentSeason `json:"seasons,omitempty"`
	Tournament        any                             `json:"tournament,omitempty"`
	Sections          map[string]any                  `json:"sections,omitempty"`
	SectionErrors     map[string]string               `json:"section_errors,omitempty"`
}

type TournamentSeasonsParams struct {
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore tournament id"`
}

type TournamentSeasonsResponse struct {
	OK           bool                            `json:"ok"`
	TournamentID int                             `json:"tournament_id"`
	Seasons      []sofascoreapi.TournamentSeason `json:"seasons"`
}

type TournamentEventsParams struct {
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore tournament id"`
	SeasonID     int    `json:"season_id,omitempty" jsonschema:"optional season id"`
	Next         bool   `json:"next,omitempty" jsonschema:"when true, return upcoming events"`
	Last         bool   `json:"last,omitempty" jsonschema:"when true, return recent events"`
	Round        int    `json:"round,omitempty" jsonschema:"when greater than 0, return events for that round"`
	Slug         string `json:"slug,omitempty" jsonschema:"optional round slug; requires round"`
	Limit        int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type TournamentEventsResponse struct {
	OK           bool                        `json:"ok"`
	TournamentID int                         `json:"tournament_id"`
	SeasonID     int                         `json:"season_id"`
	SeasonName   string                      `json:"season_name,omitempty"`
	Mode         string                      `json:"mode"`
	Round        int                         `json:"round,omitempty"`
	Slug         string                      `json:"slug,omitempty"`
	Events       []sofascoreapi.EventSummary `json:"events"`
}

type TournamentScheduledEventsParams struct {
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	Date         string `json:"date,omitempty" jsonschema:"optional UTC date in YYYY-MM-DD format; defaults to today"`
	Limit        int    `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type TournamentScheduledEventsResponse struct {
	OK           bool                        `json:"ok"`
	TournamentID int                         `json:"tournament_id"`
	Date         string                      `json:"date"`
	Events       []sofascoreapi.EventSummary `json:"events"`
}

type TeamEventsParams struct {
	TeamID int  `json:"team_id" jsonschema:"the SofaScore team id or tennis player id"`
	Next   bool `json:"next,omitempty" jsonschema:"when true, return upcoming events"`
	Last   bool `json:"last,omitempty" jsonschema:"when true, return recent events"`
	Limit  int  `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type TeamEventsResponse struct {
	OK        bool                        `json:"ok"`
	TeamID    int                         `json:"team_id"`
	Direction string                      `json:"direction"`
	Events    []sofascoreapi.EventSummary `json:"events"`
}

type EventTVParams struct {
	EventID int `json:"event_id" jsonschema:"the SofaScore event id"`
}

type EventTVResponse struct {
	OK       bool `json:"ok"`
	EventID  int  `json:"event_id"`
	Channels any  `json:"channels"`
}

type EventTVChannelParams struct {
	EventID   int `json:"event_id" jsonschema:"the SofaScore event id"`
	ChannelID int `json:"channel_id" jsonschema:"the SofaScore TV channel id"`
}

type EventTVChannelResponse struct {
	OK          bool   `json:"ok"`
	EventID     int    `json:"event_id"`
	ChannelID   int    `json:"channel_id"`
	ChannelName string `json:"channel_name,omitempty"`
	Channel     any    `json:"channel,omitempty"`
	Votes       any    `json:"votes"`
}

type EventH2HEventsParams struct {
	EventID int `json:"event_id" jsonschema:"the SofaScore event id"`
}

type EventH2HEventsResponse struct {
	OK      bool   `json:"ok"`
	EventID int    `json:"event_id"`
	Slug    string `json:"slug"`
	Events  any    `json:"events"`
}

type TeamInfoParams struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamInfoResponse struct {
	OK            bool `json:"ok"`
	TeamID        int  `json:"team_id"`
	Team          any  `json:"team"`
	FeaturedEvent any  `json:"featured_event"`
	Performance   any  `json:"performance"`
	Tournaments   any  `json:"tournaments"`
}

type TeamTournamentsParams struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamTournamentsResponse struct {
	OK          bool `json:"ok"`
	TeamID      int  `json:"team_id"`
	Tournaments any  `json:"tournaments"`
}

type TeamStandingsParams struct {
	TeamID   int `json:"team_id" jsonschema:"the SofaScore team id"`
	SeasonID int `json:"season_id,omitempty" jsonschema:"optional season id"`
}

type TeamStandingsResponse struct {
	OK        bool `json:"ok"`
	TeamID    int  `json:"team_id"`
	SeasonID  int  `json:"season_id"`
	Standings any  `json:"standings"`
}

type TeamStatsParams struct {
	TeamID       int `json:"team_id" jsonschema:"the SofaScore team id"`
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int `json:"season_id" jsonschema:"the SofaScore season id"`
}

type TeamStatsResponse struct {
	OK           bool `json:"ok"`
	TeamID       int  `json:"team_id"`
	TournamentID int  `json:"tournament_id"`
	SeasonID     int  `json:"season_id"`
	Stats        any  `json:"stats"`
}

type TeamPlayersParams struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamPlayersResponse struct {
	OK      bool `json:"ok"`
	TeamID  int  `json:"team_id"`
	Players any  `json:"players"`
}

type TeamMediaParams struct {
	TeamID int `json:"team_id" jsonschema:"the SofaScore team id"`
}

type TeamMediaResponse struct {
	OK     bool `json:"ok"`
	TeamID int  `json:"team_id"`
	Videos any  `json:"videos"`
}

type TeamRankingsParams struct {
	TeamID       int `json:"team_id" jsonschema:"the SofaScore team id"`
	TournamentID int `json:"tournament_id,omitempty" jsonschema:"optional unique tournament id"`
	SeasonID     int `json:"season_id,omitempty" jsonschema:"optional season id"`
}

type TeamRankingsResponse struct {
	OK           bool   `json:"ok"`
	TeamID       int    `json:"team_id"`
	Sport        string `json:"sport,omitempty"`
	TournamentID int    `json:"tournament_id,omitempty"`
	SeasonID     int    `json:"season_id,omitempty"`
	Rankings     any    `json:"rankings"`
}

type TeamTopPlayersParams struct {
	TeamID       int `json:"team_id" jsonschema:"the SofaScore team id"`
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int `json:"season_id" jsonschema:"the SofaScore season id"`
}

type TeamTopPlayersResponse struct {
	OK           bool `json:"ok"`
	TeamID       int  `json:"team_id"`
	TournamentID int  `json:"tournament_id"`
	SeasonID     int  `json:"season_id"`
	Players      any  `json:"players"`
}

type PlayerAttributesParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerAttributesResponse struct {
	OK         bool `json:"ok"`
	PlayerID   int  `json:"player_id"`
	Attributes any  `json:"attributes"`
}

type PlayerMediaParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerMediaResponse struct {
	OK       bool `json:"ok"`
	PlayerID int  `json:"player_id"`
	Media    any  `json:"media"`
}

type PlayerMediaVideosParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerMediaVideosResponse struct {
	OK       bool `json:"ok"`
	PlayerID int  `json:"player_id"`
	Videos   any  `json:"videos"`
}

type PlayerLastEventsParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
	Limit    int `json:"limit,omitempty" jsonschema:"optional event limit; 0 means no extra limit"`
}

type PlayerLastEventsResponse struct {
	OK       bool                        `json:"ok"`
	PlayerID int                         `json:"player_id"`
	Events   []sofascoreapi.EventSummary `json:"events"`
}

type PlayerSeasonsParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerSeasonsResponse struct {
	OK       bool `json:"ok"`
	PlayerID int  `json:"player_id"`
	Seasons  any  `json:"seasons"`
}

type PlayerCareerParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerCareerResponse struct {
	OK               bool `json:"ok"`
	PlayerID         int  `json:"player_id"`
	Career           any  `json:"career"`
	MatchTypeOverall any  `json:"match_type_overall,omitempty"`
}

type PlayerSeasonStatsParams struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase,omitempty" jsonschema:"optional phase name; defaults to the sport-specific route default"`
}

type PlayerSeasonStatsResponse struct {
	OK           bool   `json:"ok"`
	PlayerID     int    `json:"player_id"`
	TournamentID int    `json:"tournament_id"`
	SeasonID     int    `json:"season_id"`
	Phase        string `json:"phase"`
	Stats        any    `json:"stats"`
}

type PlayerSeasonRatingsParams struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase,omitempty" jsonschema:"optional phase name; defaults to the sport-specific route default"`
}

type PlayerSeasonRatingsResponse struct {
	OK           bool   `json:"ok"`
	PlayerID     int    `json:"player_id"`
	TournamentID int    `json:"tournament_id"`
	SeasonID     int    `json:"season_id"`
	Phase        string `json:"phase"`
	Ratings      any    `json:"ratings"`
}

type PlayerCharacteristicsParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerCharacteristicsResponse struct {
	OK              bool `json:"ok"`
	PlayerID        int  `json:"player_id"`
	Characteristics any  `json:"characteristics"`
}

type PlayerNationalTeamStatsParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerNationalTeamStatsResponse struct {
	OK       bool `json:"ok"`
	PlayerID int  `json:"player_id"`
	Stats    any  `json:"stats"`
}

type PlayerTournamentsParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerTournamentsResponse struct {
	OK          bool `json:"ok"`
	PlayerID    int  `json:"player_id"`
	Tournaments any  `json:"tournaments"`
}

type PlayerSeasonHeatmapParams struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase" jsonschema:"the phase name, for example overall"`
}

type PlayerSeasonHeatmapResponse struct {
	OK           bool   `json:"ok"`
	PlayerID     int    `json:"player_id"`
	TournamentID int    `json:"tournament_id"`
	SeasonID     int    `json:"season_id"`
	Phase        string `json:"phase"`
	Heatmap      any    `json:"heatmap"`
}

type PlayerPenaltyHistoryParams struct {
	PlayerID     int `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int `json:"season_id" jsonschema:"the SofaScore season id"`
}

type PlayerPenaltyHistoryResponse struct {
	OK             bool `json:"ok"`
	PlayerID       int  `json:"player_id"`
	TournamentID   int  `json:"tournament_id"`
	SeasonID       int  `json:"season_id"`
	PenaltyHistory any  `json:"penalty_history"`
}

type PlayerShotActionsParams struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase" jsonschema:"the phase name, for example regularSeason"`
}

type PlayerShotActionsResponse struct {
	OK           bool   `json:"ok"`
	PlayerID     int    `json:"player_id"`
	TournamentID int    `json:"tournament_id"`
	SeasonID     int    `json:"season_id"`
	Phase        string `json:"phase"`
	ShotActions  any    `json:"shot_actions"`
}

type PlayerShotActionAreasParams struct {
	PlayerID     int    `json:"player_id" jsonschema:"the SofaScore player id"`
	TournamentID int    `json:"tournament_id" jsonschema:"the SofaScore unique tournament id"`
	SeasonID     int    `json:"season_id" jsonschema:"the SofaScore season id"`
	Phase        string `json:"phase" jsonschema:"the phase name, for example regularSeason"`
}

type PlayerShotActionAreasResponse struct {
	OK           bool   `json:"ok"`
	PlayerID     int    `json:"player_id"`
	TournamentID int    `json:"tournament_id"`
	SeasonID     int    `json:"season_id"`
	Phase        string `json:"phase"`
	Areas        any    `json:"areas"`
}

type PlayerYearStatsParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
	Year     int `json:"year" jsonschema:"the calendar year, for example 2026"`
}

type PlayerYearStatsResponse struct {
	OK       bool `json:"ok"`
	PlayerID int  `json:"player_id"`
	Year     int  `json:"year"`
	Stats    any  `json:"stats"`
}

type PlayerFeaturedEventParams struct {
	PlayerID int `json:"player_id" jsonschema:"the SofaScore player id"`
}

type PlayerFeaturedEventResponse struct {
	OK       bool `json:"ok"`
	PlayerID int  `json:"player_id"`
	Event    any  `json:"event"`
}

type SportLiveTournamentsParams struct {
	Sport string `json:"sport" jsonschema:"the sport slug"`
}

type SportLiveTournamentsResponse struct {
	OK          bool   `json:"ok"`
	Sport       string `json:"sport"`
	Tournaments any    `json:"tournaments"`
}

type SportCategoriesParams struct {
	Sport string `json:"sport" jsonschema:"the sport slug"`
}

type SportCategoriesResponse struct {
	OK         bool   `json:"ok"`
	Sport      string `json:"sport"`
	Categories any    `json:"categories"`
}

type SportTopPlayersParams struct {
	Sport string `json:"sport" jsonschema:"the sport slug"`
}

type SportTopPlayersResponse struct {
	OK      bool   `json:"ok"`
	Sport   string `json:"sport"`
	Players any    `json:"players"`
}

func Search(ctx context.Context, svc Service, params SearchParams) (SearchResponse, error) {
	query := strings.TrimSpace(params.Query)
	if query == "" {
		return SearchResponse{}, invalidf("search query is required")
	}
	if params.Page < 0 {
		return SearchResponse{}, invalidf("page must be zero or greater")
	}

	results, err := svc.Search(ctx, query, params.Page)
	if err != nil {
		return SearchResponse{}, translateServiceError(err)
	}

	filtered := filterSearchResults(results, params.ResultType, params.Sport)
	filtered = limitSearchResults(filtered, params.Limit)

	response := SearchResponse{
		OK:      true,
		Query:   query,
		Page:    params.Page,
		Results: filtered,
	}
	if params.IDOnly {
		response.IDs = make([]int, 0, len(filtered))
		for _, result := range filtered {
			response.IDs = append(response.IDs, result.ID)
		}
	}

	return response, nil
}

func Sports(ctx context.Context, svc Service, params SportsParams) (SportsResponse, error) {
	sport := strings.TrimSpace(params.Sport)
	if params.Sections {
		if sport == "" {
			return SportsResponse{}, invalidf("sports --sections requires a sport slug")
		}

		discovery, err := svc.SportSections(ctx, sport)
		if err != nil {
			return SportsResponse{}, translateServiceError(err)
		}

		return SportsResponse{
			OK:            true,
			Sport:         discovery.Sport,
			SampleEventID: discovery.SampleEventID,
			Sections:      discovery.Sections,
		}, nil
	}

	sports, err := svc.Sports(ctx)
	if err != nil {
		return SportsResponse{}, translateServiceError(err)
	}
	if sport != "" {
		sports = filterSports(sports, sport)
	}

	return SportsResponse{
		OK:     true,
		Sports: sports,
	}, nil
}

func SportEvents(ctx context.Context, svc Service, params SportEventsParams) (SportEventsResponse, error) {
	sport := strings.TrimSpace(params.Sport)
	if sport == "" {
		return SportEventsResponse{}, invalidf("sports events requires a sport slug")
	}

	date, err := normalizeDate(params.Date)
	if err != nil {
		return SportEventsResponse{}, err
	}

	events, err := svc.SportEvents(ctx, sport, date, params.Limit)
	if err != nil {
		return SportEventsResponse{}, translateServiceError(err)
	}
	events = limitEventResults(events, params.Limit)

	return SportEventsResponse{
		OK:     true,
		Sport:  sport,
		Date:   date,
		Events: events,
	}, nil
}

func SportScheduledTournaments(ctx context.Context, svc Service, params SportScheduledTournamentsParams) (SportScheduledTournamentsResponse, error) {
	sport := strings.TrimSpace(params.Sport)
	if sport == "" {
		return SportScheduledTournamentsResponse{}, invalidf("sports tournaments requires a sport slug")
	}
	date, err := normalizeDate(params.Date)
	if err != nil {
		return SportScheduledTournamentsResponse{}, err
	}
	page := params.Page
	if page == 0 {
		page = 1
	}
	if page < 1 {
		return SportScheduledTournamentsResponse{}, invalidf("page must be 1 or greater")
	}

	tournaments, hasNextPage, err := svc.SportScheduledTournaments(ctx, sport, date, page)
	if err != nil {
		return SportScheduledTournamentsResponse{}, translateServiceError(err)
	}

	return SportScheduledTournamentsResponse{
		OK:          true,
		Sport:       sport,
		Date:        date,
		Page:        page,
		HasNextPage: hasNextPage,
		Tournaments: tournaments,
	}, nil
}

func Trending(ctx context.Context, svc Service, params TrendingParams) (TrendingResponse, error) {
	date, err := normalizeDate(params.Date)
	if err != nil {
		return TrendingResponse{}, err
	}

	country := strings.ToUpper(strings.TrimSpace(params.Country))
	if country == "" {
		country, err = svc.DetectCountryAlpha2(ctx)
		if err != nil {
			return TrendingResponse{}, translateServiceError(err)
		}
		country = strings.ToUpper(strings.TrimSpace(country))
	}
	if err := validateCountryAlpha2(country); err != nil {
		return TrendingResponse{}, err
	}

	events, err := svc.TrendingEvents(ctx, country, 0)
	if err != nil {
		return TrendingResponse{}, translateServiceError(err)
	}

	filtered := make([]sofascoreapi.TrendingEventSummary, 0, len(events))
	for _, event := range events {
		if event.StartTime.UTC().Format("2006-01-02") != date {
			continue
		}
		filtered = append(filtered, event)
	}
	filtered = limitTrendingEventResults(filtered, params.Limit)

	return TrendingResponse{
		OK:      true,
		Country: country,
		Date:    date,
		Events:  filtered,
	}, nil
}

func Event(ctx context.Context, svc Service, params EventParams) (EventResponse, error) {
	if params.EventID <= 0 {
		return EventResponse{}, invalidf("event id must be a positive integer")
	}
	if params.SectionsOnly && len(params.Sections) > 0 {
		return EventResponse{}, invalidf("event --sections cannot be combined with --section")
	}

	event, err := svc.Event(ctx, params.EventID)
	if err != nil {
		return EventResponse{}, translateServiceError(err)
	}

	availableSections, err := svc.ProbeEventSections(ctx, params.EventID)
	if err != nil {
		return EventResponse{}, translateServiceError(err)
	}

	response := EventResponse{
		OK:                true,
		EventID:           params.EventID,
		StartTime:         event.StartTime.Format(time.RFC3339),
		StatusType:        event.StatusType,
		StatusDescription: event.StatusDescription,
		Home:              event.Home,
		Away:              event.Away,
		Tournament:        event.Tournament,
		Sport:             event.Sport,
		Venue:             event.Venue,
		HomeScore:         event.HomeScore,
		AwayScore:         event.AwayScore,
		AvailableSections: availableSections,
	}
	if params.SectionsOnly {
		return response, nil
	}

	eventValue, err := decodeRaw(event.Raw)
	if err != nil {
		return EventResponse{}, translateServiceError(err)
	}
	response.Event = eventValue
	response.Sections = map[string]any{}

	for _, section := range params.Sections {
		payload, err := svc.EventSection(ctx, params.EventID, section)
		if err != nil {
			if !params.AllowPartialSectionErrors {
				return EventResponse{}, eventSectionError(params.EventID, section, err)
			}
			if response.SectionErrors == nil {
				response.SectionErrors = map[string]string{}
			}
			response.SectionErrors[section] = eventSectionError(params.EventID, section, err).Error()
			continue
		}

		value, err := decodeRaw(payload)
		if err != nil {
			return EventResponse{}, translateServiceError(err)
		}
		response.Sections[section] = value
	}

	response.Partial = len(response.SectionErrors) > 0
	return response, nil
}

func Tournament(ctx context.Context, svc Service, params TournamentParams) (TournamentResponse, error) {
	if params.TournamentID <= 0 {
		return TournamentResponse{}, invalidf("tournament id must be a positive integer")
	}
	if params.SectionsOnly && len(params.Sections) > 0 {
		return TournamentResponse{}, invalidf("tournament --sections cannot be combined with --section")
	}

	tournament, err := svc.Tournament(ctx, params.TournamentID)
	if err != nil {
		return TournamentResponse{}, translateServiceError(err)
	}

	season, seasons, err := resolveTournamentSeason(ctx, svc, params.TournamentID, params.SeasonID)
	if err != nil {
		return TournamentResponse{}, err
	}

	availableSections, err := svc.ProbeTournamentSections(ctx, params.TournamentID, season.ID)
	if err != nil {
		return TournamentResponse{}, translateServiceError(err)
	}

	response := TournamentResponse{
		OK:                true,
		TournamentID:      params.TournamentID,
		Name:              tournament.Name,
		Sport:             tournament.Sport,
		Category:          tournament.Category,
		Country:           tournament.Country,
		SeasonID:          season.ID,
		SeasonName:        seasonDisplayName(season),
		AvailableSections: availableSections,
	}
	if params.SectionsOnly {
		return response, nil
	}

	tournamentValue, err := decodeRaw(tournament.Raw)
	if err != nil {
		return TournamentResponse{}, translateServiceError(err)
	}
	response.Seasons = seasons
	response.Tournament = tournamentValue
	response.Sections = map[string]any{}

	for _, section := range params.Sections {
		payload, err := svc.TournamentSection(ctx, params.TournamentID, season.ID, section)
		if err != nil {
			if !params.AllowPartialSectionErrors {
				return TournamentResponse{}, tournamentSectionError(params.TournamentID, season.ID, section, err)
			}
			if response.SectionErrors == nil {
				response.SectionErrors = map[string]string{}
			}
			response.SectionErrors[section] = tournamentSectionError(params.TournamentID, season.ID, section, err).Error()
			continue
		}

		value, err := decodeRaw(payload)
		if err != nil {
			return TournamentResponse{}, translateServiceError(err)
		}
		response.Sections[section] = value
	}

	response.Partial = len(response.SectionErrors) > 0
	return response, nil
}

func TournamentSeasons(ctx context.Context, svc Service, params TournamentSeasonsParams) (TournamentSeasonsResponse, error) {
	if params.TournamentID <= 0 {
		return TournamentSeasonsResponse{}, invalidf("tournament id must be a positive integer")
	}

	seasons, err := svc.TournamentSeasons(ctx, params.TournamentID)
	if err != nil {
		return TournamentSeasonsResponse{}, translateServiceError(err)
	}

	return TournamentSeasonsResponse{
		OK:           true,
		TournamentID: params.TournamentID,
		Seasons:      seasons,
	}, nil
}

func TournamentEvents(ctx context.Context, svc Service, params TournamentEventsParams) (TournamentEventsResponse, error) {
	if params.TournamentID <= 0 {
		return TournamentEventsResponse{}, invalidf("tournament id must be a positive integer")
	}

	slug := strings.TrimSpace(params.Slug)
	if slug != "" && params.Round <= 0 {
		return TournamentEventsResponse{}, invalidf("tournament events --slug requires --round")
	}

	modeCount := 0
	if params.Next {
		modeCount++
	}
	if params.Last {
		modeCount++
	}
	if params.Round > 0 {
		modeCount++
	}
	if modeCount != 1 {
		return TournamentEventsResponse{}, invalidf("pass exactly one of --next, --last, or --round")
	}

	season, _, err := resolveTournamentSeason(ctx, svc, params.TournamentID, params.SeasonID)
	if err != nil {
		return TournamentEventsResponse{}, err
	}

	mode := "round"
	if params.Next {
		mode = "next"
	}
	if params.Last {
		mode = "last"
	}

	events, err := svc.TournamentEvents(ctx, params.TournamentID, season.ID, mode, params.Round, slug, params.Limit)
	if err != nil {
		return TournamentEventsResponse{}, translateServiceError(err)
	}
	events = limitEventResults(events, params.Limit)

	return TournamentEventsResponse{
		OK:           true,
		TournamentID: params.TournamentID,
		SeasonID:     season.ID,
		SeasonName:   seasonDisplayName(season),
		Mode:         mode,
		Round:        params.Round,
		Slug:         slug,
		Events:       events,
	}, nil
}

func TournamentScheduledEvents(ctx context.Context, svc Service, params TournamentScheduledEventsParams) (TournamentScheduledEventsResponse, error) {
	if params.TournamentID <= 0 {
		return TournamentScheduledEventsResponse{}, invalidf("tournament id must be a positive integer")
	}
	date, err := normalizeDate(params.Date)
	if err != nil {
		return TournamentScheduledEventsResponse{}, err
	}

	events, err := svc.TournamentScheduledEvents(ctx, params.TournamentID, date, params.Limit)
	if err != nil {
		return TournamentScheduledEventsResponse{}, translateServiceError(err)
	}
	events = limitEventResults(events, params.Limit)

	return TournamentScheduledEventsResponse{
		OK:           true,
		TournamentID: params.TournamentID,
		Date:         date,
		Events:       events,
	}, nil
}

func TeamEvents(ctx context.Context, svc Service, params TeamEventsParams) (TeamEventsResponse, error) {
	if params.TeamID <= 0 {
		return TeamEventsResponse{}, invalidf("team id must be a positive integer")
	}
	if params.Next == params.Last {
		return TeamEventsResponse{}, invalidf("pass exactly one of --next or --last")
	}

	direction := "last"
	if params.Next {
		direction = "next"
	}

	events, err := svc.TeamEvents(ctx, params.TeamID, direction, params.Limit)
	if err != nil {
		return TeamEventsResponse{}, translateServiceError(err)
	}
	events = limitEventResults(events, params.Limit)

	return TeamEventsResponse{
		OK:        true,
		TeamID:    params.TeamID,
		Direction: direction,
		Events:    events,
	}, nil
}

func EventTV(ctx context.Context, svc Service, params EventTVParams) (EventTVResponse, error) {
	if params.EventID <= 0 {
		return EventTVResponse{}, invalidf("event id must be a positive integer")
	}

	raw, err := svc.EventTVChannels(ctx, params.EventID)
	if err != nil {
		return EventTVResponse{}, translateServiceError(err)
	}
	channels, err := decodeRaw(raw)
	if err != nil {
		return EventTVResponse{}, translateServiceError(err)
	}

	return EventTVResponse{
		OK:       true,
		EventID:  params.EventID,
		Channels: channels,
	}, nil
}

func EventTVChannel(ctx context.Context, svc Service, params EventTVChannelParams) (EventTVChannelResponse, error) {
	if params.EventID <= 0 {
		return EventTVChannelResponse{}, invalidf("event id must be a positive integer")
	}
	if params.ChannelID <= 0 {
		return EventTVChannelResponse{}, invalidf("channel id must be a positive integer")
	}

	raw, err := svc.EventTVChannelVotes(ctx, params.ChannelID, params.EventID)
	if err != nil {
		return EventTVChannelResponse{}, translateServiceError(err)
	}
	votes, err := decodeRaw(raw)
	if err != nil {
		return EventTVChannelResponse{}, translateServiceError(err)
	}

	response := EventTVChannelResponse{
		OK:        true,
		EventID:   params.EventID,
		ChannelID: params.ChannelID,
		Votes:     votes,
	}

	root, ok := votes.(map[string]any)
	if !ok {
		return response, nil
	}
	tvChannelVotes, ok := root["tvChannelVotes"].(map[string]any)
	if !ok {
		return response, nil
	}
	channel, ok := tvChannelVotes["tvChannel"].(map[string]any)
	if !ok {
		return response, nil
	}
	response.Channel = channel
	if name, ok := channel["name"].(string); ok {
		response.ChannelName = strings.TrimSpace(name)
	}

	return response, nil
}

func EventH2HEvents(ctx context.Context, svc Service, params EventH2HEventsParams) (EventH2HEventsResponse, error) {
	if params.EventID <= 0 {
		return EventH2HEventsResponse{}, invalidf("event id must be a positive integer")
	}

	result, err := svc.EventH2HEvents(ctx, params.EventID)
	if err != nil {
		return EventH2HEventsResponse{}, translateServiceError(err)
	}
	events, err := decodeRaw(result.Raw)
	if err != nil {
		return EventH2HEventsResponse{}, translateServiceError(err)
	}

	return EventH2HEventsResponse{
		OK:      true,
		EventID: params.EventID,
		Slug:    result.Slug,
		Events:  events,
	}, nil
}

func TeamInfo(ctx context.Context, svc Service, params TeamInfoParams) (TeamInfoResponse, error) {
	if params.TeamID <= 0 {
		return TeamInfoResponse{}, invalidf("team id must be a positive integer")
	}

	result, err := svc.TeamInfo(ctx, params.TeamID)
	if err != nil {
		return TeamInfoResponse{}, translateServiceError(err)
	}
	team, err := decodeRaw(result.Team)
	if err != nil {
		return TeamInfoResponse{}, translateServiceError(err)
	}
	featuredEvent, err := decodeRaw(result.FeaturedEvent)
	if err != nil {
		return TeamInfoResponse{}, translateServiceError(err)
	}
	performance, err := decodeRaw(result.Performance)
	if err != nil {
		return TeamInfoResponse{}, translateServiceError(err)
	}
	tournaments, err := decodeRaw(result.Tournaments)
	if err != nil {
		return TeamInfoResponse{}, translateServiceError(err)
	}

	return TeamInfoResponse{
		OK:            true,
		TeamID:        params.TeamID,
		Team:          team,
		FeaturedEvent: featuredEvent,
		Performance:   performance,
		Tournaments:   tournaments,
	}, nil
}

func TeamTournaments(ctx context.Context, svc Service, params TeamTournamentsParams) (TeamTournamentsResponse, error) {
	if params.TeamID <= 0 {
		return TeamTournamentsResponse{}, invalidf("team id must be a positive integer")
	}

	raw, err := svc.TeamTournaments(ctx, params.TeamID)
	if err != nil {
		return TeamTournamentsResponse{}, translateServiceError(err)
	}
	tournaments, err := decodeRaw(raw)
	if err != nil {
		return TeamTournamentsResponse{}, translateServiceError(err)
	}

	return TeamTournamentsResponse{
		OK:          true,
		TeamID:      params.TeamID,
		Tournaments: tournaments,
	}, nil
}

func TeamStandings(ctx context.Context, svc Service, params TeamStandingsParams) (TeamStandingsResponse, error) {
	if params.TeamID <= 0 {
		return TeamStandingsResponse{}, invalidf("team id must be a positive integer")
	}
	if params.SeasonID < 0 {
		return TeamStandingsResponse{}, invalidf("season id must be zero or greater")
	}

	result, err := svc.TeamStandings(ctx, params.TeamID, params.SeasonID)
	if err != nil {
		return TeamStandingsResponse{}, translateServiceError(err)
	}
	standings, err := decodeRaw(result.Raw)
	if err != nil {
		return TeamStandingsResponse{}, translateServiceError(err)
	}

	return TeamStandingsResponse{
		OK:        true,
		TeamID:    params.TeamID,
		SeasonID:  result.Season.SeasonID,
		Standings: standings,
	}, nil
}

func TeamStats(ctx context.Context, svc Service, params TeamStatsParams) (TeamStatsResponse, error) {
	if params.TeamID <= 0 || params.TournamentID <= 0 || params.SeasonID <= 0 {
		return TeamStatsResponse{}, invalidf("team, tournament, and season ids must be positive integers")
	}

	raw, err := svc.TeamTournamentStatistics(ctx, params.TeamID, params.TournamentID, params.SeasonID)
	if err != nil {
		return TeamStatsResponse{}, translateServiceError(err)
	}
	stats, err := decodeRaw(raw)
	if err != nil {
		return TeamStatsResponse{}, translateServiceError(err)
	}

	return TeamStatsResponse{
		OK:           true,
		TeamID:       params.TeamID,
		TournamentID: params.TournamentID,
		SeasonID:     params.SeasonID,
		Stats:        stats,
	}, nil
}

func TeamPlayers(ctx context.Context, svc Service, params TeamPlayersParams) (TeamPlayersResponse, error) {
	if params.TeamID <= 0 {
		return TeamPlayersResponse{}, invalidf("team id must be a positive integer")
	}

	raw, err := svc.TeamFeaturedPlayers(ctx, params.TeamID)
	if err != nil {
		return TeamPlayersResponse{}, translateServiceError(err)
	}
	players, err := decodeRaw(raw)
	if err != nil {
		return TeamPlayersResponse{}, translateServiceError(err)
	}

	return TeamPlayersResponse{
		OK:      true,
		TeamID:  params.TeamID,
		Players: players,
	}, nil
}

func TeamMedia(ctx context.Context, svc Service, params TeamMediaParams) (TeamMediaResponse, error) {
	if params.TeamID <= 0 {
		return TeamMediaResponse{}, invalidf("team id must be a positive integer")
	}

	raw, err := svc.TeamMediaVideos(ctx, params.TeamID)
	if err != nil {
		return TeamMediaResponse{}, translateServiceError(err)
	}
	videos, err := decodeRaw(raw)
	if err != nil {
		return TeamMediaResponse{}, translateServiceError(err)
	}

	return TeamMediaResponse{
		OK:     true,
		TeamID: params.TeamID,
		Videos: videos,
	}, nil
}

func TeamRankings(ctx context.Context, svc Service, params TeamRankingsParams) (TeamRankingsResponse, error) {
	if params.TeamID <= 0 {
		return TeamRankingsResponse{}, invalidf("team id must be a positive integer")
	}

	hasTournament := params.TournamentID > 0
	hasSeason := params.SeasonID > 0
	if hasTournament != hasSeason {
		return TeamRankingsResponse{}, invalidf("team rankings requires both tournament id and season id together")
	}

	var (
		raw any
		err error
	)
	if hasTournament {
		rawPayload, serviceErr := svc.TeamTournamentRanks(ctx, params.TeamID, params.TournamentID, params.SeasonID)
		if serviceErr != nil {
			return TeamRankingsResponse{}, translateServiceError(serviceErr)
		}
		raw, err = decodeRaw(rawPayload)
		if err != nil {
			return TeamRankingsResponse{}, translateServiceError(err)
		}
		return TeamRankingsResponse{
			OK:           true,
			TeamID:       params.TeamID,
			TournamentID: params.TournamentID,
			SeasonID:     params.SeasonID,
			Rankings:     raw,
		}, nil
	}

	rawPayload, serviceErr := svc.TeamRankings(ctx, params.TeamID)
	if serviceErr != nil {
		return TeamRankingsResponse{}, translateServiceError(serviceErr)
	}
	raw, err = decodeRaw(rawPayload)
	if err != nil {
		return TeamRankingsResponse{}, translateServiceError(err)
	}

	return TeamRankingsResponse{
		OK:       true,
		TeamID:   params.TeamID,
		Sport:    "tennis",
		Rankings: raw,
	}, nil
}

func TeamTopPlayers(ctx context.Context, svc Service, params TeamTopPlayersParams) (TeamTopPlayersResponse, error) {
	if params.TeamID <= 0 || params.TournamentID <= 0 || params.SeasonID <= 0 {
		return TeamTopPlayersResponse{}, invalidf("team, tournament, and season ids must be positive integers")
	}

	raw, err := svc.TeamTournamentTopPlayers(ctx, params.TeamID, params.TournamentID, params.SeasonID)
	if err != nil {
		return TeamTopPlayersResponse{}, translateServiceError(err)
	}
	players, err := decodeRaw(raw)
	if err != nil {
		return TeamTopPlayersResponse{}, translateServiceError(err)
	}

	return TeamTopPlayersResponse{
		OK:           true,
		TeamID:       params.TeamID,
		TournamentID: params.TournamentID,
		SeasonID:     params.SeasonID,
		Players:      players,
	}, nil
}

func PlayerAttributes(ctx context.Context, svc Service, params PlayerAttributesParams) (PlayerAttributesResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerAttributesResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerAttributeOverviews(ctx, params.PlayerID)
	if err != nil {
		return PlayerAttributesResponse{}, translateServiceError(err)
	}
	attributes, err := decodeRaw(raw)
	if err != nil {
		return PlayerAttributesResponse{}, translateServiceError(err)
	}

	return PlayerAttributesResponse{
		OK:         true,
		PlayerID:   params.PlayerID,
		Attributes: attributes,
	}, nil
}

func PlayerMedia(ctx context.Context, svc Service, params PlayerMediaParams) (PlayerMediaResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerMediaResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerMedia(ctx, params.PlayerID)
	if err != nil {
		return PlayerMediaResponse{}, translateServiceError(err)
	}
	media, err := decodeRaw(raw)
	if err != nil {
		return PlayerMediaResponse{}, translateServiceError(err)
	}

	return PlayerMediaResponse{
		OK:       true,
		PlayerID: params.PlayerID,
		Media:    media,
	}, nil
}

func PlayerMediaVideos(ctx context.Context, svc Service, params PlayerMediaVideosParams) (PlayerMediaVideosResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerMediaVideosResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerMediaVideos(ctx, params.PlayerID)
	if isHTTPNotFound(err) {
		raw, err = svc.TeamMediaVideos(ctx, params.PlayerID)
	}
	if err != nil {
		return PlayerMediaVideosResponse{}, translateServiceError(err)
	}
	videos, err := decodeRaw(raw)
	if err != nil {
		return PlayerMediaVideosResponse{}, translateServiceError(err)
	}

	return PlayerMediaVideosResponse{
		OK:       true,
		PlayerID: params.PlayerID,
		Videos:   videos,
	}, nil
}

func PlayerLastEvents(ctx context.Context, svc Service, params PlayerLastEventsParams) (PlayerLastEventsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerLastEventsResponse{}, invalidf("player id must be a positive integer")
	}

	events, err := svc.PlayerLastEvents(ctx, params.PlayerID, params.Limit)
	if isHTTPNotFound(err) {
		events, err = svc.TeamEvents(ctx, params.PlayerID, "last", params.Limit)
	}
	if err != nil {
		return PlayerLastEventsResponse{}, translateServiceError(err)
	}

	return PlayerLastEventsResponse{
		OK:       true,
		PlayerID: params.PlayerID,
		Events:   events,
	}, nil
}

func PlayerSeasons(ctx context.Context, svc Service, params PlayerSeasonsParams) (PlayerSeasonsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerSeasonsResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerStatisticsSeasons(ctx, params.PlayerID)
	if isHTTPNotFound(err) {
		raw, err = svc.PlayerStatisticsSeasonsTennis(ctx, params.PlayerID)
	}
	if err != nil {
		return PlayerSeasonsResponse{}, translateServiceError(err)
	}
	seasons, err := decodeRaw(raw)
	if err != nil {
		return PlayerSeasonsResponse{}, translateServiceError(err)
	}

	return PlayerSeasonsResponse{
		OK:       true,
		PlayerID: params.PlayerID,
		Seasons:  seasons,
	}, nil
}

func PlayerCareer(ctx context.Context, svc Service, params PlayerCareerParams) (PlayerCareerResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerCareerResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerCareerStatistics(ctx, params.PlayerID)
	if err != nil {
		return PlayerCareerResponse{}, translateServiceError(err)
	}
	career, err := decodeRaw(raw)
	if err != nil {
		return PlayerCareerResponse{}, translateServiceError(err)
	}

	var matchTypeOverall any
	raw, err = svc.PlayerCareerStatisticsMatchType(ctx, params.PlayerID, "overall")
	switch {
	case err == nil:
		matchTypeOverall, err = decodeRaw(raw)
		if err != nil {
			return PlayerCareerResponse{}, translateServiceError(err)
		}
	case isHTTPNotFound(err):
		matchTypeOverall = nil
	default:
		return PlayerCareerResponse{}, translateServiceError(err)
	}

	return PlayerCareerResponse{
		OK:               true,
		PlayerID:         params.PlayerID,
		Career:           career,
		MatchTypeOverall: matchTypeOverall,
	}, nil
}

func PlayerSeasonStats(ctx context.Context, svc Service, params PlayerSeasonStatsParams) (PlayerSeasonStatsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerSeasonStatsResponse{}, invalidf("player id must be a positive integer")
	}
	if params.TournamentID <= 0 {
		return PlayerSeasonStatsResponse{}, invalidf("tournament id must be a positive integer")
	}
	if params.SeasonID <= 0 {
		return PlayerSeasonStatsResponse{}, invalidf("season id must be a positive integer")
	}

	phase, raw, err := fetchPlayerSeasonPhase(ctx, svc.PlayerSeasonStatistics, params.PlayerID, params.TournamentID, params.SeasonID, params.Phase)
	if err != nil {
		return PlayerSeasonStatsResponse{}, err
	}
	stats, err := decodeRaw(raw)
	if err != nil {
		return PlayerSeasonStatsResponse{}, translateServiceError(err)
	}

	return PlayerSeasonStatsResponse{
		OK:           true,
		PlayerID:     params.PlayerID,
		TournamentID: params.TournamentID,
		SeasonID:     params.SeasonID,
		Phase:        phase,
		Stats:        stats,
	}, nil
}

func PlayerSeasonRatings(ctx context.Context, svc Service, params PlayerSeasonRatingsParams) (PlayerSeasonRatingsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerSeasonRatingsResponse{}, invalidf("player id must be a positive integer")
	}
	if params.TournamentID <= 0 {
		return PlayerSeasonRatingsResponse{}, invalidf("tournament id must be a positive integer")
	}
	if params.SeasonID <= 0 {
		return PlayerSeasonRatingsResponse{}, invalidf("season id must be a positive integer")
	}

	phase, raw, err := fetchPlayerSeasonPhase(ctx, svc.PlayerSeasonRatings, params.PlayerID, params.TournamentID, params.SeasonID, params.Phase)
	if err != nil {
		return PlayerSeasonRatingsResponse{}, err
	}
	ratings, err := decodeRaw(raw)
	if err != nil {
		return PlayerSeasonRatingsResponse{}, translateServiceError(err)
	}

	return PlayerSeasonRatingsResponse{
		OK:           true,
		PlayerID:     params.PlayerID,
		TournamentID: params.TournamentID,
		SeasonID:     params.SeasonID,
		Phase:        phase,
		Ratings:      ratings,
	}, nil
}

func PlayerCharacteristics(ctx context.Context, svc Service, params PlayerCharacteristicsParams) (PlayerCharacteristicsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerCharacteristicsResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerCharacteristics(ctx, params.PlayerID)
	if err != nil {
		return PlayerCharacteristicsResponse{}, translateServiceError(err)
	}
	characteristics, err := decodeRaw(raw)
	if err != nil {
		return PlayerCharacteristicsResponse{}, translateServiceError(err)
	}

	return PlayerCharacteristicsResponse{
		OK:              true,
		PlayerID:        params.PlayerID,
		Characteristics: characteristics,
	}, nil
}

func PlayerNationalTeamStats(ctx context.Context, svc Service, params PlayerNationalTeamStatsParams) (PlayerNationalTeamStatsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerNationalTeamStatsResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerNationalTeamStatistics(ctx, params.PlayerID)
	if err != nil {
		return PlayerNationalTeamStatsResponse{}, translateServiceError(err)
	}
	stats, err := decodeRaw(raw)
	if err != nil {
		return PlayerNationalTeamStatsResponse{}, translateServiceError(err)
	}

	return PlayerNationalTeamStatsResponse{
		OK:       true,
		PlayerID: params.PlayerID,
		Stats:    stats,
	}, nil
}

func PlayerTournaments(ctx context.Context, svc Service, params PlayerTournamentsParams) (PlayerTournamentsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerTournamentsResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerUniqueTournaments(ctx, params.PlayerID)
	if err != nil {
		return PlayerTournamentsResponse{}, translateServiceError(err)
	}
	tournaments, err := decodeRaw(raw)
	if err != nil {
		return PlayerTournamentsResponse{}, translateServiceError(err)
	}

	return PlayerTournamentsResponse{
		OK:          true,
		PlayerID:    params.PlayerID,
		Tournaments: tournaments,
	}, nil
}

func PlayerSeasonHeatmap(ctx context.Context, svc Service, params PlayerSeasonHeatmapParams) (PlayerSeasonHeatmapResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerSeasonHeatmapResponse{}, invalidf("player id must be a positive integer")
	}
	if params.TournamentID <= 0 {
		return PlayerSeasonHeatmapResponse{}, invalidf("tournament id must be a positive integer")
	}
	if params.SeasonID <= 0 {
		return PlayerSeasonHeatmapResponse{}, invalidf("season id must be a positive integer")
	}
	phase := strings.TrimSpace(params.Phase)
	if phase == "" {
		return PlayerSeasonHeatmapResponse{}, invalidf("phase is required")
	}

	raw, err := svc.PlayerSeasonHeatmap(ctx, params.PlayerID, params.TournamentID, params.SeasonID, phase)
	if err != nil {
		return PlayerSeasonHeatmapResponse{}, translateServiceError(err)
	}
	heatmap, err := decodeRaw(raw)
	if err != nil {
		return PlayerSeasonHeatmapResponse{}, translateServiceError(err)
	}

	return PlayerSeasonHeatmapResponse{
		OK:           true,
		PlayerID:     params.PlayerID,
		TournamentID: params.TournamentID,
		SeasonID:     params.SeasonID,
		Phase:        phase,
		Heatmap:      heatmap,
	}, nil
}

func PlayerPenaltyHistory(ctx context.Context, svc Service, params PlayerPenaltyHistoryParams) (PlayerPenaltyHistoryResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerPenaltyHistoryResponse{}, invalidf("player id must be a positive integer")
	}
	if params.TournamentID <= 0 {
		return PlayerPenaltyHistoryResponse{}, invalidf("tournament id must be a positive integer")
	}
	if params.SeasonID <= 0 {
		return PlayerPenaltyHistoryResponse{}, invalidf("season id must be a positive integer")
	}

	raw, err := svc.PlayerPenaltyHistory(ctx, params.PlayerID, params.TournamentID, params.SeasonID)
	if err != nil {
		return PlayerPenaltyHistoryResponse{}, translateServiceError(err)
	}
	penaltyHistory, err := decodeRaw(raw)
	if err != nil {
		return PlayerPenaltyHistoryResponse{}, translateServiceError(err)
	}

	return PlayerPenaltyHistoryResponse{
		OK:             true,
		PlayerID:       params.PlayerID,
		TournamentID:   params.TournamentID,
		SeasonID:       params.SeasonID,
		PenaltyHistory: penaltyHistory,
	}, nil
}

func PlayerShotActions(ctx context.Context, svc Service, params PlayerShotActionsParams) (PlayerShotActionsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerShotActionsResponse{}, invalidf("player id must be a positive integer")
	}
	if params.TournamentID <= 0 {
		return PlayerShotActionsResponse{}, invalidf("tournament id must be a positive integer")
	}
	if params.SeasonID <= 0 {
		return PlayerShotActionsResponse{}, invalidf("season id must be a positive integer")
	}
	phase := strings.TrimSpace(params.Phase)
	if phase == "" {
		return PlayerShotActionsResponse{}, invalidf("phase is required")
	}

	raw, err := svc.PlayerShotActions(ctx, params.PlayerID, params.TournamentID, params.SeasonID, phase)
	if err != nil {
		return PlayerShotActionsResponse{}, translateServiceError(err)
	}
	shotActions, err := decodeRaw(raw)
	if err != nil {
		return PlayerShotActionsResponse{}, translateServiceError(err)
	}

	return PlayerShotActionsResponse{
		OK:           true,
		PlayerID:     params.PlayerID,
		TournamentID: params.TournamentID,
		SeasonID:     params.SeasonID,
		Phase:        phase,
		ShotActions:  shotActions,
	}, nil
}

func PlayerShotActionAreas(ctx context.Context, svc Service, params PlayerShotActionAreasParams) (PlayerShotActionAreasResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerShotActionAreasResponse{}, invalidf("player id must be a positive integer")
	}
	if params.TournamentID <= 0 {
		return PlayerShotActionAreasResponse{}, invalidf("tournament id must be a positive integer")
	}
	if params.SeasonID <= 0 {
		return PlayerShotActionAreasResponse{}, invalidf("season id must be a positive integer")
	}
	phase := strings.TrimSpace(params.Phase)
	if phase == "" {
		return PlayerShotActionAreasResponse{}, invalidf("phase is required")
	}

	raw, err := svc.PlayerShotActionAreas(ctx, params.TournamentID, params.SeasonID, phase)
	if err != nil {
		return PlayerShotActionAreasResponse{}, translateServiceError(err)
	}
	areas, err := decodeRaw(raw)
	if err != nil {
		return PlayerShotActionAreasResponse{}, translateServiceError(err)
	}

	return PlayerShotActionAreasResponse{
		OK:           true,
		PlayerID:     params.PlayerID,
		TournamentID: params.TournamentID,
		SeasonID:     params.SeasonID,
		Phase:        phase,
		Areas:        areas,
	}, nil
}

func PlayerYearStats(ctx context.Context, svc Service, params PlayerYearStatsParams) (PlayerYearStatsResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerYearStatsResponse{}, invalidf("player id must be a positive integer")
	}
	if params.Year <= 0 {
		return PlayerYearStatsResponse{}, invalidf("year must be a positive integer")
	}

	raw, err := svc.PlayerYearStatistics(ctx, params.PlayerID, params.Year)
	if err != nil {
		return PlayerYearStatsResponse{}, translateServiceError(err)
	}
	stats, err := decodeRaw(raw)
	if err != nil {
		return PlayerYearStatsResponse{}, translateServiceError(err)
	}

	return PlayerYearStatsResponse{
		OK:       true,
		PlayerID: params.PlayerID,
		Year:     params.Year,
		Stats:    stats,
	}, nil
}

func PlayerFeaturedEvent(ctx context.Context, svc Service, params PlayerFeaturedEventParams) (PlayerFeaturedEventResponse, error) {
	if params.PlayerID <= 0 {
		return PlayerFeaturedEventResponse{}, invalidf("player id must be a positive integer")
	}

	raw, err := svc.PlayerFeaturedEvent(ctx, params.PlayerID)
	if err != nil {
		return PlayerFeaturedEventResponse{}, translateServiceError(err)
	}
	event, err := decodeRaw(raw)
	if err != nil {
		return PlayerFeaturedEventResponse{}, translateServiceError(err)
	}

	return PlayerFeaturedEventResponse{
		OK:       true,
		PlayerID: params.PlayerID,
		Event:    event,
	}, nil
}

func SportLiveTournaments(ctx context.Context, svc Service, params SportLiveTournamentsParams) (SportLiveTournamentsResponse, error) {
	sport := strings.TrimSpace(params.Sport)
	if sport == "" {
		return SportLiveTournamentsResponse{}, invalidf("sport is required")
	}

	raw, err := svc.SportLiveTournaments(ctx, sport)
	if err != nil {
		return SportLiveTournamentsResponse{}, translateServiceError(err)
	}
	tournaments, err := decodeRaw(raw)
	if err != nil {
		return SportLiveTournamentsResponse{}, translateServiceError(err)
	}

	return SportLiveTournamentsResponse{
		OK:          true,
		Sport:       sport,
		Tournaments: tournaments,
	}, nil
}

func SportCategories(ctx context.Context, svc Service, params SportCategoriesParams) (SportCategoriesResponse, error) {
	sport := strings.TrimSpace(params.Sport)
	if sport == "" {
		return SportCategoriesResponse{}, invalidf("sport is required")
	}

	raw, err := svc.SportCategories(ctx, sport)
	if err != nil {
		return SportCategoriesResponse{}, translateServiceError(err)
	}
	categories, err := decodeRaw(raw)
	if err != nil {
		return SportCategoriesResponse{}, translateServiceError(err)
	}

	return SportCategoriesResponse{
		OK:         true,
		Sport:      sport,
		Categories: categories,
	}, nil
}

func SportTopPlayers(ctx context.Context, svc Service, params SportTopPlayersParams) (SportTopPlayersResponse, error) {
	sport := strings.TrimSpace(params.Sport)
	if sport == "" {
		return SportTopPlayersResponse{}, invalidf("sport is required")
	}

	raw, err := svc.SportTrendingTopPlayers(ctx, sport)
	if err != nil {
		return SportTopPlayersResponse{}, translateServiceError(err)
	}
	players, err := decodeRaw(raw)
	if err != nil {
		return SportTopPlayersResponse{}, translateServiceError(err)
	}

	return SportTopPlayersResponse{
		OK:      true,
		Sport:   sport,
		Players: players,
	}, nil
}

func filterSearchResults(results []sofascoreapi.SearchResult, resultType, sport string) []sofascoreapi.SearchResult {
	resultType = normalizeSearchResultType(resultType)
	sport = strings.TrimSpace(sport)
	filtered := make([]sofascoreapi.SearchResult, 0, len(results))
	for _, result := range results {
		if resultType != "" && result.Type != resultType {
			continue
		}
		if sport != "" && result.Sport != sport {
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered
}

func filterSports(sports []sofascoreapi.SportCount, slug string) []sofascoreapi.SportCount {
	slug = strings.TrimSpace(slug)
	filtered := make([]sofascoreapi.SportCount, 0, len(sports))
	for _, sport := range sports {
		if sport.Slug == slug {
			filtered = append(filtered, sport)
		}
	}
	return filtered
}

func limitSearchResults(results []sofascoreapi.SearchResult, limit int) []sofascoreapi.SearchResult {
	if limit <= 0 || limit >= len(results) {
		return results
	}
	return results[:limit]
}

func limitEventResults(events []sofascoreapi.EventSummary, limit int) []sofascoreapi.EventSummary {
	if limit <= 0 || limit >= len(events) {
		return events
	}
	return events[:limit]
}

func limitTrendingEventResults(events []sofascoreapi.TrendingEventSummary, limit int) []sofascoreapi.TrendingEventSummary {
	if limit <= 0 || limit >= len(events) {
		return events
	}
	return events[:limit]
}

func normalizeSearchResultType(resultType string) string {
	switch strings.TrimSpace(resultType) {
	case "tournament":
		return "uniqueTournament"
	default:
		return strings.TrimSpace(resultType)
	}
}

func resolveTournamentSeason(ctx context.Context, svc Service, tournamentID, seasonID int) (sofascoreapi.TournamentSeason, []sofascoreapi.TournamentSeason, error) {
	seasons, err := svc.TournamentSeasons(ctx, tournamentID)
	if err != nil {
		return sofascoreapi.TournamentSeason{}, nil, translateServiceError(err)
	}
	if len(seasons) == 0 {
		return sofascoreapi.TournamentSeason{}, nil, &Error{
			Kind:    ErrorKindNotFound,
			Message: fmt.Sprintf("no seasons found for tournament %d", tournamentID),
		}
	}

	if seasonID <= 0 {
		return seasons[0], seasons, nil
	}

	for _, season := range seasons {
		if season.ID == seasonID {
			return season, seasons, nil
		}
	}

	return sofascoreapi.TournamentSeason{}, seasons, &Error{
		Kind:    ErrorKindNotFound,
		Message: fmt.Sprintf("season %d not found for tournament %d", seasonID, tournamentID),
	}
}

func normalizeDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return now().UTC().Format("2006-01-02"), nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return "", invalidf("date must use YYYY-MM-DD")
	}
	return value, nil
}

func validateCountryAlpha2(value string) error {
	if len(value) != 2 {
		return invalidf("country must be a 2-letter ISO alpha-2 code")
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return invalidf("country must be a 2-letter ISO alpha-2 code")
		}
	}
	return nil
}

func seasonDisplayName(season sofascoreapi.TournamentSeason) string {
	if strings.TrimSpace(season.Name) != "" {
		return season.Name
	}
	return season.Year
}

func isHTTPNotFound(err error) bool {
	var statusErr *sofascoreapi.HTTPStatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode == 404
}

func fetchPlayerSeasonPhase(
	ctx context.Context,
	fetch func(context.Context, int, int, int, string) (json.RawMessage, error),
	playerID, tournamentID, seasonID int,
	phase string,
) (string, json.RawMessage, error) {
	phase = strings.TrimSpace(phase)
	if phase != "" {
		raw, err := fetch(ctx, playerID, tournamentID, seasonID, phase)
		if err != nil {
			return "", nil, translateServiceError(err)
		}
		return phase, raw, nil
	}

	for _, candidate := range []string{"overall", "regularSeason"} {
		raw, err := fetch(ctx, playerID, tournamentID, seasonID, candidate)
		if err == nil {
			return candidate, raw, nil
		}
		if !isHTTPNotFound(err) {
			return "", nil, translateServiceError(err)
		}
	}

	return "", nil, translateServiceError(&sofascoreapi.HTTPStatusError{StatusCode: 404})
}

func decodeRaw(raw json.RawMessage) (any, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}
