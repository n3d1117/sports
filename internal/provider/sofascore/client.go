package sofascoreapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://www.sofascore.com/api/v1"

var candidateEventSections = []string{
	"statistics",
	"h2h",
	"incidents",
	"lineups",
	"graph",
	"average-positions",
	"best-players",
	"comments",
	"highlights",
	"innings",
	"official-tweets",
	"team-streaks",
	"tennis-power",
	"point-by-point",
	"best-players/summary",
	"pregame-form",
	"managers",
	"umpires",
	"votes",
	"weather",
	"esports-games",
}

type probeMethod string

const (
	probeMethodHead probeMethod = "HEAD"
	probeMethodGet  probeMethod = "GET"
)

type tournamentSectionCandidate struct {
	Name   string
	Scope  string
	Method probeMethod
}

var candidateTournamentSections = []tournamentSectionCandidate{
	{Name: "featured-events", Scope: "tournament", Method: probeMethodGet},
	{Name: "media", Scope: "tournament", Method: probeMethodHead},
	{Name: "player-news", Scope: "tournament", Method: probeMethodHead},
	{Name: "info", Scope: "season", Method: probeMethodGet},
	{Name: "rounds", Scope: "season", Method: probeMethodHead},
	{Name: "standings/total", Scope: "season", Method: probeMethodHead},
	{Name: "groups", Scope: "season", Method: probeMethodHead},
	{Name: "cuptrees", Scope: "season", Method: probeMethodHead},
	{Name: "venues", Scope: "season", Method: probeMethodHead},
	{Name: "player-statistics/types", Scope: "season", Method: probeMethodGet},
	{Name: "team-statistics/types", Scope: "season", Method: probeMethodGet},
	{Name: "player-of-the-season-race", Scope: "season", Method: probeMethodGet},
	{Name: "team-events/total", Scope: "season", Method: probeMethodGet},
	{Name: "draft", Scope: "season", Method: probeMethodHead},
	{Name: "team-of-the-week/periods", Scope: "season", Method: probeMethodGet},
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	now        func() time.Time
}

type SearchResult struct {
	Type     string  `json:"type"`
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Sport    string  `json:"sport,omitempty"`
	Category string  `json:"category,omitempty"`
	Country  string  `json:"country,omitempty"`
	Team     string  `json:"team,omitempty"`
	Score    float64 `json:"score,omitempty"`
	Slug     string  `json:"slug,omitempty"`
}

type EventSummary struct {
	EventID           int       `json:"event_id"`
	StartTime         time.Time `json:"start_time"`
	StatusType        string    `json:"status_type,omitempty"`
	StatusDescription string    `json:"status_description,omitempty"`
	Home              string    `json:"home"`
	Away              string    `json:"away"`
	Tournament        string    `json:"tournament,omitempty"`
	Sport             string    `json:"sport,omitempty"`
}

type TrendingEventSummary struct {
	Rank              int       `json:"rank"`
	EventID           int       `json:"event_id"`
	StartTime         time.Time `json:"start_time"`
	StatusType        string    `json:"status_type,omitempty"`
	StatusDescription string    `json:"status_description,omitempty"`
	Home              string    `json:"home"`
	Away              string    `json:"away"`
	Tournament        string    `json:"tournament,omitempty"`
	Sport             string    `json:"sport,omitempty"`
}

type EventDetail struct {
	EventID           int             `json:"event_id"`
	StartTime         time.Time       `json:"start_time"`
	StatusType        string          `json:"status_type,omitempty"`
	StatusDescription string          `json:"status_description,omitempty"`
	Home              string          `json:"home"`
	Away              string          `json:"away"`
	Tournament        string          `json:"tournament,omitempty"`
	Sport             string          `json:"sport,omitempty"`
	Venue             string          `json:"venue,omitempty"`
	HomeScore         *int            `json:"home_score,omitempty"`
	AwayScore         *int            `json:"away_score,omitempty"`
	CustomID          string          `json:"-"`
	Raw               json.RawMessage `json:"-"`
}

type SportCount struct {
	Slug  string `json:"slug"`
	Live  int    `json:"live"`
	Total int    `json:"total"`
}

type SportSectionDiscovery struct {
	Sport         string   `json:"sport"`
	SampleEventID int      `json:"sample_event_id"`
	Sections      []string `json:"sections"`
}

type ScheduledTournamentSummary struct {
	TournamentID         int            `json:"tournament_id"`
	UniqueTournamentID   int            `json:"unique_tournament_id"`
	Name                 string         `json:"name"`
	Slug                 string         `json:"slug,omitempty"`
	UniqueTournament     string         `json:"unique_tournament,omitempty"`
	UniqueTournamentSlug string         `json:"unique_tournament_slug,omitempty"`
	Category             string         `json:"category,omitempty"`
	Sport                string         `json:"sport,omitempty"`
	UTCEventCount        int            `json:"utc_event_count,omitempty"`
	TimezoneEventCount   map[string]int `json:"timezone_event_count,omitempty"`
}

type TournamentDetail struct {
	TournamentID int             `json:"tournament_id"`
	Name         string          `json:"name"`
	Slug         string          `json:"slug,omitempty"`
	Sport        string          `json:"sport,omitempty"`
	Category     string          `json:"category,omitempty"`
	Country      string          `json:"country,omitempty"`
	Raw          json.RawMessage `json:"-"`
}

type TournamentSeason struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Year string `json:"year,omitempty"`
}

type EventH2HEventsResult struct {
	Slug string
	Raw  json.RawMessage
}

type TeamInfoResult struct {
	Team          json.RawMessage
	FeaturedEvent json.RawMessage
	Performance   json.RawMessage
	Tournaments   json.RawMessage
}

type TeamStandingsSeason struct {
	TournamentID         int    `json:"tournament_id,omitempty"`
	TournamentName       string `json:"tournament_name,omitempty"`
	UniqueTournamentID   int    `json:"unique_tournament_id,omitempty"`
	UniqueTournamentName string `json:"unique_tournament_name,omitempty"`
	SeasonID             int    `json:"season_id"`
	SeasonName           string `json:"season_name,omitempty"`
	SeasonYear           string `json:"season_year,omitempty"`
}

type TeamStandingsResult struct {
	Season TeamStandingsSeason
	Raw    json.RawMessage
}

type UnsupportedTournamentEventsError struct {
	TournamentID int
	SeasonID     int
	Direction    string
}

func (e *UnsupportedTournamentEventsError) Error() string {
	return fmt.Sprintf("%s events are not available for tournament %d season %d", e.Direction, e.TournamentID, e.SeasonID)
}

type HTTPStatusError struct {
	StatusCode int
	URL        string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("request failed with HTTP %d for %s", e.StatusCode, e.URL)
}

func New(baseURL string, httpClient *http.Client) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		now:        time.Now,
	}
}

func (c *Client) Search(ctx context.Context, query string, page int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if page < 0 {
		return nil, fmt.Errorf("page must be zero or greater")
	}

	endpoint := c.baseURL + "/search/all?q=" + url.QueryEscape(query) + "&page=" + strconv.Itoa(page)
	var decoded searchResponse
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(decoded.Results))
	for _, item := range decoded.Results {
		result := SearchResult{
			Type:     item.Type,
			ID:       item.Entity.ID,
			Name:     item.Entity.Name,
			Sport:    item.Entity.Sport.Slug,
			Category: item.Entity.Category.Name,
			Score:    item.Score,
			Slug:     item.Entity.Slug,
		}
		if result.Sport == "" {
			result.Sport = item.Entity.Team.Sport.Slug
		}
		if result.Sport == "" {
			result.Sport = item.Entity.Category.Sport.Slug
		}
		if item.Entity.Country.Name != "" {
			result.Country = item.Entity.Country.Name
		}
		if result.Country == "" {
			result.Country = item.Entity.Category.Country.Name
		}
		if item.Entity.Team.Name != "" {
			result.Team = item.Entity.Team.Name
		}
		results = append(results, result)
	}

	return results, nil
}

func (c *Client) TeamEvents(ctx context.Context, teamID int, direction string, limit int) ([]EventSummary, error) {
	if teamID <= 0 {
		return nil, fmt.Errorf("team id must be positive")
	}
	if direction != "next" && direction != "last" {
		return nil, fmt.Errorf("direction must be next or last")
	}

	if direction == "next" {
		return c.nextTeamEvents(ctx, teamID, limit)
	}
	return c.lastTeamEvents(ctx, teamID, limit)
}

func (c *Client) nextTeamEvents(ctx context.Context, teamID int, limit int) ([]EventSummary, error) {
	capacity := 0
	if limit > 0 {
		capacity = limit
	}
	events := make([]EventSummary, 0, capacity)
	for page := 0; ; page++ {
		decoded, err := c.teamEventsPage(ctx, teamID, "next", page)
		if err != nil {
			if page == 0 {
				return c.nextTeamEventsFallback(ctx, teamID, err, limit)
			}
			return nil, err
		}

		for _, event := range decoded.Events {
			events = append(events, eventSummary(event))
		}

		if limit > 0 && len(events) >= limit {
			break
		}
		if !decoded.HasNextPage || len(decoded.Events) == 0 {
			break
		}
	}

	if len(events) == 0 {
		return c.nextTeamEventsFallback(ctx, teamID, nil, limit)
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].EventID < events[j].EventID
		}
		return events[i].StartTime.Before(events[j].StartTime)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) nextTeamEventsFallback(ctx context.Context, teamID int, nextErr error, limit int) ([]EventSummary, error) {
	// SofaScore sometimes returns 404 for tennis player "next" events even though the
	// same player id still has upcoming matches in the "last" feed. When that
	// happens, confirm the subject still exists, then reverse-scan the first "last"
	// page for not-started matches and sort them into upcoming order.
	if nextErr != nil {
		exists, existsErr := c.teamExists(ctx, teamID)
		if existsErr != nil {
			return nil, existsErr
		}
		if !exists {
			return nil, nextErr
		}
	}

	decoded, err := c.teamEventsPage(ctx, teamID, "last", 0)
	if err != nil {
		return nil, err
	}

	events := make([]EventSummary, 0, len(decoded.Events))
	for _, event := range decoded.Events {
		if event.Status.Type != "notstarted" {
			continue
		}
		events = append(events, eventSummary(event))
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].EventID < events[j].EventID
		}
		return events[i].StartTime.Before(events[j].StartTime)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) lastTeamEvents(ctx context.Context, teamID int, limit int) ([]EventSummary, error) {
	capacity := 0
	if limit > 0 {
		capacity = limit
	}
	events := make([]EventSummary, 0, capacity)
	for page := 0; ; page++ {
		decoded, err := c.teamEventsPage(ctx, teamID, "last", page)
		if err != nil {
			return nil, err
		}

		for _, event := range decoded.Events {
			if event.Status.Type == "notstarted" {
				continue
			}
			events = append(events, eventSummary(event))
		}

		if limit > 0 && len(events) >= limit {
			break
		}
		if !decoded.HasNextPage || len(decoded.Events) == 0 {
			break
		}
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].EventID > events[j].EventID
		}
		return events[i].StartTime.After(events[j].StartTime)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) teamEventsPage(ctx context.Context, teamID int, direction string, page int) (teamEventsResponse, error) {
	endpoint := c.baseURL + "/team/" + strconv.Itoa(teamID) + "/events/" + direction + "/" + strconv.Itoa(page)
	var decoded teamEventsResponse
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return teamEventsResponse{}, err
	}
	return decoded, nil
}

func (c *Client) playerEventsPage(ctx context.Context, playerID int, direction string, page int) (teamEventsResponse, error) {
	endpoint := c.baseURL + "/player/" + strconv.Itoa(playerID) + "/events/" + direction + "/" + strconv.Itoa(page)
	var decoded teamEventsResponse
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return teamEventsResponse{}, err
	}
	return decoded, nil
}

func (c *Client) teamExists(ctx context.Context, teamID int) (bool, error) {
	_, err := c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID))
	if err == nil {
		return true, nil
	}

	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, err
}

func eventSummary(event apiEvent) EventSummary {
	return EventSummary{
		EventID:           event.ID,
		StartTime:         time.Unix(int64(event.StartTimestamp), 0).UTC(),
		StatusType:        event.Status.Type,
		StatusDescription: event.Status.Description,
		Home:              coalesce(event.HomeTeam.Name, event.HomePlayer.Name, "TBD"),
		Away:              coalesce(event.AwayTeam.Name, event.AwayPlayer.Name, "TBD"),
		Tournament:        event.Tournament.Name,
		Sport:             event.Tournament.Category.Sport.Slug,
	}
}

func (c *Client) Event(ctx context.Context, eventID int) (EventDetail, error) {
	if eventID <= 0 {
		return EventDetail{}, fmt.Errorf("event id must be positive")
	}

	var response eventResponse
	if err := c.getJSON(ctx, c.baseURL+"/event/"+strconv.Itoa(eventID), &response); err != nil {
		return EventDetail{}, err
	}
	if len(response.Event) == 0 {
		return EventDetail{}, fmt.Errorf("event payload was empty")
	}

	var decoded apiEventDetail
	if err := json.Unmarshal(response.Event, &decoded); err != nil {
		return EventDetail{}, err
	}

	return EventDetail{
		EventID:           decoded.ID,
		StartTime:         time.Unix(int64(decoded.StartTimestamp), 0).UTC(),
		StatusType:        decoded.Status.Type,
		StatusDescription: decoded.Status.Description,
		Home:              coalesce(decoded.HomeTeam.Name, decoded.HomePlayer.Name, "TBD"),
		Away:              coalesce(decoded.AwayTeam.Name, decoded.AwayPlayer.Name, "TBD"),
		Tournament:        decoded.Tournament.Name,
		Sport:             decoded.Tournament.Category.Sport.Slug,
		Venue:             decoded.Venue.Name,
		HomeScore:         decoded.HomeScore.Current,
		AwayScore:         decoded.AwayScore.Current,
		CustomID:          decoded.CustomID,
		Raw:               response.Event,
	}, nil
}

func (c *Client) EventTVChannels(ctx context.Context, eventID int) (json.RawMessage, error) {
	if eventID <= 0 {
		return nil, fmt.Errorf("event id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/tv/event/"+strconv.Itoa(eventID)+"/country-channels")
}

func (c *Client) EventTVChannelVotes(ctx context.Context, channelID, eventID int) (json.RawMessage, error) {
	if channelID <= 0 {
		return nil, fmt.Errorf("channel id must be positive")
	}
	if eventID <= 0 {
		return nil, fmt.Errorf("event id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/tv/channel/"+strconv.Itoa(channelID)+"/event/"+strconv.Itoa(eventID)+"/votes")
}

func (c *Client) EventH2HEvents(ctx context.Context, eventID int) (EventH2HEventsResult, error) {
	event, err := c.Event(ctx, eventID)
	if err != nil {
		return EventH2HEventsResult{}, err
	}
	slug := strings.TrimSpace(event.CustomID)
	if slug == "" {
		return EventH2HEventsResult{}, fmt.Errorf("event %d did not include customId", eventID)
	}

	raw, err := c.getRawJSON(ctx, c.baseURL+"/event/"+url.PathEscape(slug)+"/h2h/events")
	if err != nil {
		return EventH2HEventsResult{}, err
	}
	return EventH2HEventsResult{
		Slug: slug,
		Raw:  raw,
	}, nil
}

func (c *Client) EventSection(ctx context.Context, eventID int, section string) (json.RawMessage, error) {
	if eventID <= 0 {
		return nil, fmt.Errorf("event id must be positive")
	}
	section = strings.TrimSpace(section)
	if section == "" {
		return nil, fmt.Errorf("section is required")
	}

	return c.getRawJSON(ctx, c.baseURL+"/event/"+strconv.Itoa(eventID)+"/"+escapeSectionPath(section))
}

func (c *Client) TeamInfo(ctx context.Context, teamID int) (TeamInfoResult, error) {
	if teamID <= 0 {
		return TeamInfoResult{}, fmt.Errorf("team id must be positive")
	}

	team, err := c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID))
	if err != nil {
		return TeamInfoResult{}, err
	}
	featuredEvent, err := c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/featured-event")
	if err != nil {
		return TeamInfoResult{}, err
	}
	performance, err := c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/performance")
	if err != nil {
		return TeamInfoResult{}, err
	}
	tournaments, err := c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/unique-tournaments/all")
	if err != nil {
		return TeamInfoResult{}, err
	}

	return TeamInfoResult{
		Team:          team,
		FeaturedEvent: featuredEvent,
		Performance:   performance,
		Tournaments:   tournaments,
	}, nil
}

func (c *Client) TeamTournaments(ctx context.Context, teamID int) (json.RawMessage, error) {
	if teamID <= 0 {
		return nil, fmt.Errorf("team id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/unique-tournaments/all")
}

func (c *Client) TeamStandingsSeasons(ctx context.Context, teamID int) ([]TeamStandingsSeason, error) {
	if teamID <= 0 {
		return nil, fmt.Errorf("team id must be positive")
	}

	var decoded teamStandingsSeasonsResponse
	if err := c.getJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/standings/seasons", &decoded); err != nil {
		return nil, err
	}

	seasons := make([]TeamStandingsSeason, 0)
	for _, item := range decoded.UniqueTournamentSeasons {
		for _, season := range item.Seasons {
			seasons = append(seasons, TeamStandingsSeason{
				UniqueTournamentID:   item.UniqueTournament.ID,
				UniqueTournamentName: item.UniqueTournament.Name,
				SeasonID:             season.ID,
				SeasonName:           season.Name,
				SeasonYear:           season.Year,
			})
		}
	}
	if len(seasons) > 0 {
		return seasons, nil
	}

	for _, item := range decoded.TournamentSeasons {
		for _, season := range item.Seasons {
			seasons = append(seasons, TeamStandingsSeason{
				TournamentID:         item.Tournament.ID,
				TournamentName:       item.Tournament.Name,
				UniqueTournamentID:   item.Tournament.UniqueTournament.ID,
				UniqueTournamentName: item.Tournament.UniqueTournament.Name,
				SeasonID:             season.ID,
				SeasonName:           season.Name,
				SeasonYear:           season.Year,
			})
		}
	}

	return seasons, nil
}

func (c *Client) TeamStandings(ctx context.Context, teamID, seasonID int) (TeamStandingsResult, error) {
	seasons, err := c.TeamStandingsSeasons(ctx, teamID)
	if err != nil {
		return TeamStandingsResult{}, err
	}
	if len(seasons) == 0 {
		return TeamStandingsResult{}, fmt.Errorf("no standings seasons found for team %d", teamID)
	}

	selected := seasons[0]
	if seasonID > 0 {
		found := false
		for _, item := range seasons {
			if item.SeasonID == seasonID {
				selected = item
				found = true
				break
			}
		}
		if !found {
			return TeamStandingsResult{}, fmt.Errorf("season %d not found for team %d standings", seasonID, teamID)
		}
	}
	if selected.UniqueTournamentID <= 0 {
		return TeamStandingsResult{}, fmt.Errorf("standings season %d did not include unique tournament id", selected.SeasonID)
	}

	raw, err := c.getRawJSON(ctx, c.baseURL+"/unique-tournament/"+strconv.Itoa(selected.UniqueTournamentID)+"/season/"+strconv.Itoa(selected.SeasonID)+"/standings/total")
	if err != nil {
		return TeamStandingsResult{}, err
	}
	return TeamStandingsResult{
		Season: selected,
		Raw:    raw,
	}, nil
}

func (c *Client) TeamTournamentStatistics(ctx context.Context, teamID, tournamentID, seasonID int) (json.RawMessage, error) {
	if teamID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("team, tournament, and season ids must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/statistics/overall")
}

func (c *Client) TeamFeaturedPlayers(ctx context.Context, teamID int) (json.RawMessage, error) {
	if teamID <= 0 {
		return nil, fmt.Errorf("team id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/featured-players")
}

func (c *Client) TeamMediaVideos(ctx context.Context, teamID int) (json.RawMessage, error) {
	if teamID <= 0 {
		return nil, fmt.Errorf("team id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/media/videos")
}

func (c *Client) TeamRankings(ctx context.Context, teamID int) (json.RawMessage, error) {
	if teamID <= 0 {
		return nil, fmt.Errorf("team id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/rankings")
}

func (c *Client) TeamTournamentRanks(ctx context.Context, teamID, tournamentID, seasonID int) (json.RawMessage, error) {
	if teamID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("team, tournament, and season ids must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/ranks/regularSeason")
}

func (c *Client) TeamTournamentTopPlayers(ctx context.Context, teamID, tournamentID, seasonID int) (json.RawMessage, error) {
	if teamID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("team, tournament, and season ids must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(teamID)+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/top-players/regularSeason")
}

func (c *Client) PlayerAttributeOverviews(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/attribute-overviews")
}

func (c *Client) PlayerCharacteristics(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/characteristics")
}

func (c *Client) PlayerNationalTeamStatistics(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/national-team-statistics")
}

func (c *Client) PlayerMedia(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/media")
}

func (c *Client) PlayerMediaVideos(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/media/videos")
}

func (c *Client) PlayerUniqueTournaments(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/unique-tournaments")
}

func (c *Client) PlayerStatisticsSeasons(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/statistics/seasons")
}

func (c *Client) PlayerSeasonStatistics(ctx context.Context, playerID, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	if playerID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("player, tournament, and season ids must be positive")
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return nil, fmt.Errorf("phase is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/statistics/"+url.PathEscape(phase))
}

func (c *Client) PlayerSeasonRatings(ctx context.Context, playerID, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	if playerID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("player, tournament, and season ids must be positive")
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return nil, fmt.Errorf("phase is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/ratings/"+url.PathEscape(phase))
}

func (c *Client) PlayerSeasonHeatmap(ctx context.Context, playerID, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	if playerID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("player, tournament, and season ids must be positive")
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return nil, fmt.Errorf("phase is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/heatmap/"+url.PathEscape(phase))
}

func (c *Client) PlayerPenaltyHistory(ctx context.Context, playerID, tournamentID, seasonID int) (json.RawMessage, error) {
	if playerID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("player, tournament, and season ids must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/penalty-history/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID))
}

func (c *Client) PlayerCareerStatistics(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/statistics")
}

func (c *Client) PlayerCareerStatisticsMatchType(ctx context.Context, playerID int, matchType string) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	matchType = strings.TrimSpace(matchType)
	if matchType == "" {
		return nil, fmt.Errorf("match type is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/statistics/match-type/"+url.PathEscape(matchType))
}

func (c *Client) PlayerShotActions(ctx context.Context, playerID, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	if playerID <= 0 || tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("player, tournament, and season ids must be positive")
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return nil, fmt.Errorf("phase is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/player/"+strconv.Itoa(playerID)+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/shot-actions/"+url.PathEscape(phase))
}

func (c *Client) PlayerShotActionAreas(ctx context.Context, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	if tournamentID <= 0 || seasonID <= 0 {
		return nil, fmt.Errorf("tournament and season ids must be positive")
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return nil, fmt.Errorf("phase is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/unique-tournament/"+strconv.Itoa(tournamentID)+"/season/"+strconv.Itoa(seasonID)+"/shot-action-areas/player/"+url.PathEscape(phase))
}

func (c *Client) PlayerLastEvents(ctx context.Context, playerID int, limit int) ([]EventSummary, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}

	capacity := 0
	if limit > 0 {
		capacity = limit
	}
	events := make([]EventSummary, 0, capacity)
	for page := 0; ; page++ {
		decoded, err := c.playerEventsPage(ctx, playerID, "last", page)
		if err != nil {
			return nil, err
		}

		for _, event := range decoded.Events {
			if event.Status.Type == "notstarted" {
				continue
			}
			events = append(events, eventSummary(event))
		}

		if limit > 0 && len(events) >= limit {
			break
		}
		if !decoded.HasNextPage || len(decoded.Events) == 0 {
			break
		}
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].EventID > events[j].EventID
		}
		return events[i].StartTime.After(events[j].StartTime)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) PlayerFeaturedEvent(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(playerID)+"/featured-event")
}

func (c *Client) PlayerStatisticsSeasonsTennis(ctx context.Context, playerID int) (json.RawMessage, error) {
	if playerID <= 0 {
		return nil, fmt.Errorf("player id must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(playerID)+"/team-statistics/seasons")
}

func (c *Client) PlayerYearStatistics(ctx context.Context, playerID, year int) (json.RawMessage, error) {
	if playerID <= 0 || year <= 0 {
		return nil, fmt.Errorf("player id and year must be positive")
	}
	return c.getRawJSON(ctx, c.baseURL+"/team/"+strconv.Itoa(playerID)+"/year-statistics/"+strconv.Itoa(year))
}

func (c *Client) SportLiveTournaments(ctx context.Context, sport string) (json.RawMessage, error) {
	sport = strings.TrimSpace(sport)
	if sport == "" {
		return nil, fmt.Errorf("sport is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/sport/"+url.PathEscape(sport)+"/live-tournaments")
}

func (c *Client) SportCategories(ctx context.Context, sport string) (json.RawMessage, error) {
	sport = strings.TrimSpace(sport)
	if sport == "" {
		return nil, fmt.Errorf("sport is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/sport/"+url.PathEscape(sport)+"/categories/all")
}

func (c *Client) SportTrendingTopPlayers(ctx context.Context, sport string) (json.RawMessage, error) {
	sport = strings.TrimSpace(sport)
	if sport == "" {
		return nil, fmt.Errorf("sport is required")
	}
	return c.getRawJSON(ctx, c.baseURL+"/sport/"+url.PathEscape(sport)+"/trending-top-players")
}

func (c *Client) ProbeEventSections(ctx context.Context, eventID int) ([]string, error) {
	if eventID <= 0 {
		return nil, fmt.Errorf("event id must be positive")
	}

	sections := make([]string, 0, len(candidateEventSections))
	for _, section := range candidateEventSections {
		endpoint := c.baseURL + "/event/" + strconv.Itoa(eventID) + "/" + escapeSectionPath(section)
		ok, err := c.probeEndpoint(ctx, endpoint, probeMethodHead)
		if err != nil {
			return nil, err
		}
		if ok {
			sections = append(sections, section)
		}
	}

	return sections, nil
}

func (c *Client) Sports(ctx context.Context) ([]SportCount, error) {
	var decoded map[string]sportEventCount
	if err := c.getJSON(ctx, c.baseURL+"/sport/3600/event-count", &decoded); err != nil {
		return nil, err
	}

	sports := make([]SportCount, 0, len(decoded))
	for slug, counts := range decoded {
		sports = append(sports, SportCount{
			Slug:  slug,
			Live:  counts.Live,
			Total: counts.Total,
		})
	}

	sort.Slice(sports, func(i, j int) bool {
		return sports[i].Slug < sports[j].Slug
	})

	return sports, nil
}

func (c *Client) SportEvents(ctx context.Context, sport, date string, limit int) ([]EventSummary, error) {
	sport = strings.TrimSpace(sport)
	date = strings.TrimSpace(date)
	if sport == "" {
		return nil, fmt.Errorf("sport is required")
	}
	if date == "" {
		return nil, fmt.Errorf("date is required")
	}

	var decoded scheduledEventsResponse
	endpoint := c.baseURL + "/sport/" + url.PathEscape(sport) + "/scheduled-events/" + date
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return nil, err
	}

	events := make([]EventSummary, 0, len(decoded.Events))
	for _, event := range decoded.Events {
		events = append(events, eventSummary(event))
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].EventID < events[j].EventID
		}
		return events[i].StartTime.Before(events[j].StartTime)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) SportScheduledTournaments(ctx context.Context, sport, date string, page int) ([]ScheduledTournamentSummary, bool, error) {
	sport = strings.TrimSpace(sport)
	date = strings.TrimSpace(date)
	if sport == "" {
		return nil, false, fmt.Errorf("sport is required")
	}
	if date == "" {
		return nil, false, fmt.Errorf("date is required")
	}
	if page <= 0 {
		return nil, false, fmt.Errorf("page must be positive")
	}

	var decoded scheduledTournamentsResponse
	endpoint := c.baseURL + "/sport/" + url.PathEscape(sport) + "/scheduled-tournaments/" + date + "/page/" + strconv.Itoa(page)
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return nil, false, err
	}

	tournaments := make([]ScheduledTournamentSummary, 0, len(decoded.Scheduled))
	for _, item := range decoded.Scheduled {
		timezoneEventCount := decodeTimezoneEventCount(item.TimezoneEventCount)
		tournaments = append(tournaments, ScheduledTournamentSummary{
			TournamentID:         item.Tournament.ID,
			UniqueTournamentID:   item.Tournament.UniqueTournament.ID,
			Name:                 item.Tournament.Name,
			Slug:                 item.Tournament.Slug,
			UniqueTournament:     item.Tournament.UniqueTournament.Name,
			UniqueTournamentSlug: item.Tournament.UniqueTournament.Slug,
			Category:             item.Tournament.Category.Name,
			Sport:                item.Tournament.Category.Sport.Slug,
			UTCEventCount:        utcEventCount(timezoneEventCount),
			TimezoneEventCount:   timezoneEventCount,
		})
	}

	return tournaments, decoded.HasNextPage, nil
}

func (c *Client) DetectCountryAlpha2(ctx context.Context) (string, error) {
	var decoded countryAlpha2Response
	if err := c.getJSON(ctx, c.baseURL+"/country/alpha2", &decoded); err != nil {
		return "", err
	}
	alpha2 := strings.ToUpper(strings.TrimSpace(decoded.Alpha2))
	if alpha2 == "" {
		return "", fmt.Errorf("country alpha2 payload was empty")
	}
	return alpha2, nil
}

func (c *Client) TournamentScheduledEvents(ctx context.Context, tournamentID int, date string, limit int) ([]EventSummary, error) {
	if tournamentID <= 0 {
		return nil, fmt.Errorf("tournament id must be positive")
	}
	date = strings.TrimSpace(date)
	if date == "" {
		return nil, fmt.Errorf("date is required")
	}

	var decoded scheduledEventsResponse
	endpoint := c.baseURL + "/unique-tournament/" + strconv.Itoa(tournamentID) + "/scheduled-events/" + date
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return nil, err
	}

	events := make([]EventSummary, 0, len(decoded.Events))
	for _, event := range decoded.Events {
		events = append(events, eventSummary(event))
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].EventID < events[j].EventID
		}
		return events[i].StartTime.Before(events[j].StartTime)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) TrendingEvents(ctx context.Context, country string, limit int) ([]TrendingEventSummary, error) {
	country = strings.ToUpper(strings.TrimSpace(country))
	if country == "" {
		return nil, fmt.Errorf("country is required")
	}

	var decoded trendingEventsResponse
	endpoint := c.baseURL + "/trending/events/" + url.PathEscape(country) + "/all"
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return nil, err
	}

	events := make([]TrendingEventSummary, 0, len(decoded.Events))
	for index, event := range decoded.Events {
		summary := eventSummary(event)
		events = append(events, TrendingEventSummary{
			Rank:              index + 1,
			EventID:           summary.EventID,
			StartTime:         summary.StartTime,
			StatusType:        summary.StatusType,
			StatusDescription: summary.StatusDescription,
			Home:              summary.Home,
			Away:              summary.Away,
			Tournament:        summary.Tournament,
			Sport:             summary.Sport,
		})
	}

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) SportSections(ctx context.Context, sport string) (SportSectionDiscovery, error) {
	sport = strings.TrimSpace(sport)
	if sport == "" {
		return SportSectionDiscovery{}, fmt.Errorf("sport is required")
	}

	sampleEventID, err := c.findSampleEventID(ctx, sport)
	if err != nil {
		return SportSectionDiscovery{}, err
	}

	sections, err := c.ProbeEventSections(ctx, sampleEventID)
	if err != nil {
		return SportSectionDiscovery{}, err
	}

	return SportSectionDiscovery{
		Sport:         sport,
		SampleEventID: sampleEventID,
		Sections:      sections,
	}, nil
}

func (c *Client) Tournament(ctx context.Context, tournamentID int) (TournamentDetail, error) {
	if tournamentID <= 0 {
		return TournamentDetail{}, fmt.Errorf("tournament id must be positive")
	}

	var response uniqueTournamentResponse
	endpoint := c.baseURL + "/unique-tournament/" + strconv.Itoa(tournamentID)
	if err := c.getJSON(ctx, endpoint, &response); err != nil {
		return TournamentDetail{}, err
	}
	if len(response.UniqueTournament) == 0 {
		return TournamentDetail{}, fmt.Errorf("tournament payload was empty")
	}

	var decoded apiUniqueTournament
	if err := json.Unmarshal(response.UniqueTournament, &decoded); err != nil {
		return TournamentDetail{}, err
	}

	return TournamentDetail{
		TournamentID: decoded.ID,
		Name:         decoded.Name,
		Slug:         decoded.Slug,
		Sport:        decoded.Category.Sport.Slug,
		Category:     decoded.Category.Name,
		Country:      decoded.Category.Country.Name,
		Raw:          response.UniqueTournament,
	}, nil
}

func (c *Client) TournamentSeasons(ctx context.Context, tournamentID int) ([]TournamentSeason, error) {
	if tournamentID <= 0 {
		return nil, fmt.Errorf("tournament id must be positive")
	}

	var decoded tournamentSeasonsResponse
	endpoint := c.baseURL + "/unique-tournament/" + strconv.Itoa(tournamentID) + "/seasons"
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return nil, err
	}

	seasons := make([]TournamentSeason, 0, len(decoded.Seasons))
	for _, season := range decoded.Seasons {
		seasons = append(seasons, TournamentSeason{
			ID:   season.ID,
			Name: season.Name,
			Year: season.Year,
		})
	}
	return seasons, nil
}

func (c *Client) TournamentSection(ctx context.Context, tournamentID, seasonID int, section string) (json.RawMessage, error) {
	if tournamentID <= 0 {
		return nil, fmt.Errorf("tournament id must be positive")
	}
	if seasonID <= 0 {
		return nil, fmt.Errorf("season id must be positive")
	}
	section = strings.TrimSpace(section)
	if section == "" {
		return nil, fmt.Errorf("section is required")
	}

	endpoint := c.tournamentSectionEndpoint(tournamentID, seasonID, section)
	return c.getRawJSON(ctx, endpoint)
}

func (c *Client) ProbeTournamentSections(ctx context.Context, tournamentID, seasonID int) ([]string, error) {
	if tournamentID <= 0 {
		return nil, fmt.Errorf("tournament id must be positive")
	}
	if seasonID <= 0 {
		return nil, fmt.Errorf("season id must be positive")
	}

	sections := make([]string, 0, len(candidateTournamentSections))
	for _, candidate := range candidateTournamentSections {
		endpoint := c.tournamentSectionEndpointByScope(tournamentID, seasonID, candidate.Name, candidate.Scope)
		ok, err := c.probeEndpoint(ctx, endpoint, candidate.Method)
		if err != nil {
			return nil, err
		}
		if ok {
			sections = append(sections, candidate.Name)
		}
	}

	return sections, nil
}

func (c *Client) TournamentEvents(ctx context.Context, tournamentID, seasonID int, mode string, round int, slug string, limit int) ([]EventSummary, error) {
	if tournamentID <= 0 {
		return nil, fmt.Errorf("tournament id must be positive")
	}
	if seasonID <= 0 {
		return nil, fmt.Errorf("season id must be positive")
	}

	switch mode {
	case "next":
		return c.tournamentPagedEvents(ctx, tournamentID, seasonID, mode, limit)
	case "last":
		return c.tournamentPagedEvents(ctx, tournamentID, seasonID, mode, limit)
	case "round":
		return c.tournamentRoundEvents(ctx, tournamentID, seasonID, round, slug, limit)
	default:
		return nil, fmt.Errorf("mode must be next, last, or round")
	}
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) error {
	body, err := c.getRawJSON(ctx, endpoint)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

func (c *Client) getRawJSON(ctx context.Context, endpoint string) (json.RawMessage, error) {
	body, err := c.doRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) head(ctx context.Context, endpoint string) error {
	_, err := c.doRequest(ctx, http.MethodHead, endpoint)
	return err
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SportsApp/1.0")
	req.Header.Set("Referer", "https://www.sofascore.com/")
	req.Header.Set("Origin", "https://www.sofascore.com")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, &HTTPStatusError{
			StatusCode: res.StatusCode,
			URL:        endpoint,
		}
	}

	if method == http.MethodHead {
		return nil, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) probeEndpoint(ctx context.Context, endpoint string, method probeMethod) (bool, error) {
	if method == probeMethodHead {
		if err := c.head(ctx, endpoint); err == nil {
			return true, nil
		} else {
			var statusErr *HTTPStatusError
			if !errors.As(err, &statusErr) {
				return false, err
			}
		}
	}

	_, err := c.getRawJSON(ctx, endpoint)
	if err == nil {
		return true, nil
	}

	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError, http.StatusMethodNotAllowed:
			return false, nil
		}
	}
	return false, err
}

func (c *Client) tournamentPagedEvents(ctx context.Context, tournamentID, seasonID int, direction string, limit int) ([]EventSummary, error) {
	capacity := 0
	if limit > 0 {
		capacity = limit
	}
	events := make([]EventSummary, 0, capacity)
	for page := 0; ; page++ {
		decoded, err := c.tournamentEventsPage(ctx, tournamentID, seasonID, direction, page)
		if err != nil {
			var statusErr *HTTPStatusError
			if page == 0 && direction == "last" && errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
				return nil, &UnsupportedTournamentEventsError{
					TournamentID: tournamentID,
					SeasonID:     seasonID,
					Direction:    direction,
				}
			}
			return nil, err
		}

		for _, event := range decoded.Events {
			events = append(events, eventSummary(event))
		}

		if limit > 0 && len(events) >= limit {
			break
		}
		if !decoded.HasNextPage || len(decoded.Events) == 0 {
			break
		}
	}

	switch direction {
	case "next":
		sort.Slice(events, func(i, j int) bool {
			if events[i].StartTime.Equal(events[j].StartTime) {
				return events[i].EventID < events[j].EventID
			}
			return events[i].StartTime.Before(events[j].StartTime)
		})
	case "last":
		sort.Slice(events, func(i, j int) bool {
			if events[i].StartTime.Equal(events[j].StartTime) {
				return events[i].EventID > events[j].EventID
			}
			return events[i].StartTime.After(events[j].StartTime)
		})
	}

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) tournamentRoundEvents(ctx context.Context, tournamentID, seasonID, round int, slug string, limit int) ([]EventSummary, error) {
	if round <= 0 {
		return nil, fmt.Errorf("round must be positive")
	}

	endpoint := c.baseURL + "/unique-tournament/" + strconv.Itoa(tournamentID) + "/season/" + strconv.Itoa(seasonID) + "/events/round/" + strconv.Itoa(round)
	if strings.TrimSpace(slug) != "" {
		endpoint += "/slug/" + url.PathEscape(strings.TrimSpace(slug))
	}

	var decoded teamEventsResponse
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return nil, err
	}

	events := make([]EventSummary, 0, len(decoded.Events))
	for _, event := range decoded.Events {
		events = append(events, eventSummary(event))
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].EventID < events[j].EventID
		}
		return events[i].StartTime.Before(events[j].StartTime)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func (c *Client) tournamentEventsPage(ctx context.Context, tournamentID, seasonID int, direction string, page int) (teamEventsResponse, error) {
	endpoint := c.baseURL + "/unique-tournament/" + strconv.Itoa(tournamentID) + "/season/" + strconv.Itoa(seasonID) + "/events/" + direction + "/" + strconv.Itoa(page)
	var decoded teamEventsResponse
	if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
		return teamEventsResponse{}, err
	}
	return decoded, nil
}

func (c *Client) tournamentSectionEndpoint(tournamentID, seasonID int, section string) string {
	if scope, ok := knownTournamentSectionScope(section); ok {
		return c.tournamentSectionEndpointByScope(tournamentID, seasonID, section, scope)
	}

	seasonEndpoint := c.tournamentSectionEndpointByScope(tournamentID, seasonID, section, "season")
	return seasonEndpoint
}

func (c *Client) tournamentSectionEndpointByScope(tournamentID, seasonID int, section, scope string) string {
	escaped := escapeSectionPath(section)
	if scope == "tournament" {
		return c.baseURL + "/unique-tournament/" + strconv.Itoa(tournamentID) + "/" + escaped
	}
	return c.baseURL + "/unique-tournament/" + strconv.Itoa(tournamentID) + "/season/" + strconv.Itoa(seasonID) + "/" + escaped
}

func knownTournamentSectionScope(section string) (string, bool) {
	switch {
	case section == "featured-events",
		section == "media",
		section == "player-news":
		return "tournament", true
	case section == "info",
		section == "rounds",
		strings.HasPrefix(section, "standings/"),
		section == "groups",
		section == "cuptrees",
		section == "venues",
		strings.HasPrefix(section, "player-statistics/"),
		strings.HasPrefix(section, "team-statistics/"),
		section == "player-of-the-season-race",
		strings.HasPrefix(section, "team-events/"),
		section == "draft",
		strings.HasPrefix(section, "team-of-the-week/"):
		return "season", true
	default:
		return "", false
	}
}

func (c *Client) findSampleEventID(ctx context.Context, sport string) (int, error) {
	for _, date := range c.discoveryDates() {
		var decoded scheduledEventsResponse
		endpoint := c.baseURL + "/sport/" + url.PathEscape(sport) + "/scheduled-events/" + date
		if err := c.getJSON(ctx, endpoint, &decoded); err != nil {
			return 0, err
		}
		if len(decoded.Events) > 0 {
			return decoded.Events[0].ID, nil
		}
	}

	return 0, fmt.Errorf("no sample events found for sport %q", sport)
}

func (c *Client) discoveryDates() []string {
	base := c.now().UTC()
	offsets := make([]int, 0, 61)
	offsets = append(offsets, 0)
	for i := 1; i <= 30; i++ {
		offsets = append(offsets, -i, i)
	}
	dates := make([]string, 0, len(offsets))
	for _, offset := range offsets {
		dates = append(dates, base.AddDate(0, 0, offset).Format("2006-01-02"))
	}
	return dates
}

func escapeSectionPath(section string) string {
	parts := strings.Split(section, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func coalesce(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

type searchResponse struct {
	Results []searchResult `json:"results"`
}

type searchResult struct {
	Entity searchEntity `json:"entity"`
	Score  float64      `json:"score"`
	Type   string       `json:"type"`
}

type searchEntity struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Slug     string          `json:"slug"`
	Sport    searchSport     `json:"sport"`
	Category searchCategory  `json:"category"`
	Country  searchCountry   `json:"country"`
	Team     searchTeamBrief `json:"team"`
}

type searchSport struct {
	Slug string `json:"slug"`
}

type searchCategory struct {
	Name    string        `json:"name"`
	Sport   searchSport   `json:"sport"`
	Country searchCountry `json:"country"`
}

type searchCountry struct {
	Name string `json:"name"`
}

type searchTeamBrief struct {
	Name  string      `json:"name"`
	Sport searchSport `json:"sport"`
}

type teamEventsResponse struct {
	Events      []apiEvent `json:"events"`
	HasNextPage bool       `json:"hasNextPage"`
}

type scheduledTournamentsResponse struct {
	Scheduled   []apiScheduledTournament `json:"scheduled"`
	HasNextPage bool                     `json:"hasNextPage"`
}

type scheduledEventsResponse struct {
	Events []apiEvent `json:"events"`
}

type trendingEventsResponse struct {
	Events []apiEvent `json:"events"`
}

type eventResponse struct {
	Event json.RawMessage `json:"event"`
}

type uniqueTournamentResponse struct {
	UniqueTournament json.RawMessage `json:"uniqueTournament"`
}

type countryAlpha2Response struct {
	Alpha2 string `json:"alpha2"`
}

type tournamentSeasonsResponse struct {
	Seasons []apiTournamentSeason `json:"seasons"`
}

type teamStandingsSeasonsResponse struct {
	TournamentSeasons       []apiTeamTournamentSeasons       `json:"tournamentSeasons"`
	UniqueTournamentSeasons []apiTeamUniqueTournamentSeasons `json:"uniqueTournamentSeasons"`
}

type apiEvent struct {
	ID             int           `json:"id"`
	StartTimestamp int           `json:"startTimestamp"`
	Status         apiStatus     `json:"status"`
	HomeTeam       apiCompetitor `json:"homeTeam"`
	AwayTeam       apiCompetitor `json:"awayTeam"`
	HomePlayer     apiCompetitor `json:"homePlayer"`
	AwayPlayer     apiCompetitor `json:"awayPlayer"`
	Tournament     apiTournament `json:"tournament"`
}

type apiScheduledTournament struct {
	Tournament         apiTournamentWithUnique `json:"tournament"`
	TimezoneEventCount json.RawMessage         `json:"timezoneEventCount"`
}

type apiEventDetail struct {
	ID             int           `json:"id"`
	CustomID       string        `json:"customId"`
	StartTimestamp int           `json:"startTimestamp"`
	Status         apiStatus     `json:"status"`
	HomeTeam       apiCompetitor `json:"homeTeam"`
	AwayTeam       apiCompetitor `json:"awayTeam"`
	HomePlayer     apiCompetitor `json:"homePlayer"`
	AwayPlayer     apiCompetitor `json:"awayPlayer"`
	Tournament     apiTournament `json:"tournament"`
	Venue          apiVenue      `json:"venue"`
	HomeScore      apiScore      `json:"homeScore"`
	AwayScore      apiScore      `json:"awayScore"`
}

type apiStatus struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type apiCompetitor struct {
	Name string `json:"name"`
}

type apiTournament struct {
	Name     string      `json:"name"`
	Category apiCategory `json:"category"`
}

type apiTournamentWithUnique struct {
	ID               int                 `json:"id"`
	Name             string              `json:"name"`
	Slug             string              `json:"slug"`
	Category         apiCategory         `json:"category"`
	UniqueTournament apiUniqueTournament `json:"uniqueTournament"`
}

type apiCategory struct {
	Name    string        `json:"name"`
	Sport   searchSport   `json:"sport"`
	Country searchCountry `json:"country"`
}

type apiUniqueTournament struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Slug     string      `json:"slug"`
	Category apiCategory `json:"category"`
}

type apiTournamentSeason struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Year string `json:"year"`
}

type apiTeamTournamentSeasons struct {
	Tournament apiTournamentWithUnique `json:"tournament"`
	Seasons    []apiTournamentSeason   `json:"seasons"`
}

type apiTeamUniqueTournamentSeasons struct {
	UniqueTournament apiUniqueTournament   `json:"uniqueTournament"`
	Seasons          []apiTournamentSeason `json:"seasons"`
}

type apiVenue struct {
	Name string `json:"name"`
}

type apiScore struct {
	Current *int `json:"current"`
}

type sportEventCount struct {
	Live  int `json:"live"`
	Total int `json:"total"`
}

func utcEventCount(counts map[string]int) int {
	if counts == nil {
		return 0
	}
	if value, ok := counts["0"]; ok {
		return value
	}
	maxValue := 0
	for _, value := range counts {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func cloneIntMap(source map[string]int) map[string]int {
	if source == nil {
		return nil
	}
	out := make(map[string]int, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}

func decodeTimezoneEventCount(raw json.RawMessage) map[string]int {
	if len(raw) == 0 {
		return nil
	}

	// SofaScore uses two wire shapes here. Scheduled-tournament payloads can expose
	// timezoneEventCount either as a plain object keyed by UTC offset or as an array
	// of {offset, count} objects, so accept both and normalize to one map shape.
	var object map[string]int
	if err := json.Unmarshal(raw, &object); err == nil {
		return cloneIntMap(object)
	}

	var array []struct {
		Offset string `json:"offset"`
		Count  int    `json:"count"`
	}
	if err := json.Unmarshal(raw, &array); err == nil {
		out := make(map[string]int, len(array))
		for _, item := range array {
			if strings.TrimSpace(item.Offset) == "" {
				continue
			}
			out[item.Offset] = item.Count
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}

	return nil
}
