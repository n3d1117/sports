package mcpserver

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sports/internal/lookups"
	"sports/internal/provider/sofascore"
)

type fakeClient struct {
	searchResults                []sofascoreapi.SearchResult
	searchErr                    error
	events                       []sofascoreapi.EventSummary
	eventsErr                    error
	event                        sofascoreapi.EventDetail
	eventErr                     error
	eventTV                      json.RawMessage
	eventTVErr                   error
	eventTVChannel               json.RawMessage
	eventTVChannelErr            error
	eventH2H                     sofascoreapi.EventH2HEventsResult
	eventH2HErr                  error
	sportEvents                  []sofascoreapi.EventSummary
	sportEventsErr               error
	sportEventsSport             string
	sportEventsDate              string
	sportEventsLimit             int
	sportScheduledTournaments    []sofascoreapi.ScheduledTournamentSummary
	sportScheduledTournamentsErr error
	sportScheduledSport          string
	sportScheduledDate           string
	sportScheduledPage           int
	sportScheduledHasNext        bool
	detectedCountry              string
	detectCountryErr             error
	detectCountryCalls           int
	trendingEvents               []sofascoreapi.TrendingEventSummary
	trendingErr                  error
	trendingCountry              string
	trendingLimit                int
	tournamentScheduledEvents    []sofascoreapi.EventSummary
	tournamentScheduledEventsErr error
	tournamentScheduledID        int
	tournamentScheduledDate      string
	tournamentScheduledLimit     int
	eventSections                map[string]json.RawMessage
	sectionErrs                  map[string]error
	availableSections            []string
	probeErr                     error
	sports                       []sofascoreapi.SportCount
	sportsErr                    error
	discovery                    sofascoreapi.SportSectionDiscovery
	discoveryErr                 error
	tournament                   sofascoreapi.TournamentDetail
	tournamentErr                error
	tournamentSeasons            []sofascoreapi.TournamentSeason
	tournamentSeasonsErr         error
	tournamentSections           map[string]json.RawMessage
	tournamentSectionErrs        map[string]error
	tournamentAvailableSections  []string
	tournamentProbeErr           error
	tournamentEvents             []sofascoreapi.EventSummary
	tournamentEventsErr          error
	watchEventIDs                []int
	watchEventSections           []string
	watchEventAllSections        bool
	watchEventRecords            []sofascoreapi.WatchRecord
	watchEventsErr               error
	watchSportsArgs              []string
	watchSportRecords            []sofascoreapi.WatchRecord
	watchSportsErr               error
	teamInfo                     sofascoreapi.TeamInfoResult
	teamInfoErr                  error
	teamTournaments              json.RawMessage
	teamTournamentsErr           error
	teamStandings                sofascoreapi.TeamStandingsResult
	teamStandingsErr             error
	teamStats                    json.RawMessage
	teamStatsErr                 error
	teamPlayers                  json.RawMessage
	teamPlayersErr               error
	teamMedia                    json.RawMessage
	teamMediaErr                 error
	teamRankings                 json.RawMessage
	teamRankingsErr              error
	teamTournamentRanks          json.RawMessage
	teamTournamentRanksErr       error
	teamTopPlayers               json.RawMessage
	teamTopPlayersErr            error
	playerAttributes             json.RawMessage
	playerAttributesErr          error
	playerCharacteristics        json.RawMessage
	playerCharacteristicsErr     error
	playerNationalTeamStats      json.RawMessage
	playerNationalTeamStatsErr   error
	playerMedia                  json.RawMessage
	playerMediaErr               error
	playerMediaVideos            json.RawMessage
	playerMediaVideosErr         error
	playerTournaments            json.RawMessage
	playerTournamentsErr         error
	playerSeasons                json.RawMessage
	playerSeasonsErr             error
	playerSeasonsTennis          json.RawMessage
	playerSeasonsTennisErr       error
	playerSeasonStats            json.RawMessage
	playerSeasonStatsErr         error
	playerSeasonRatings          json.RawMessage
	playerSeasonRatingsErr       error
	playerSeasonHeatmap          json.RawMessage
	playerSeasonHeatmapErr       error
	playerPenaltyHistory         json.RawMessage
	playerPenaltyHistoryErr      error
	playerCareer                 json.RawMessage
	playerCareerErr              error
	playerCareerMatchType        json.RawMessage
	playerCareerMatchTypeErr     error
	playerShotActions            json.RawMessage
	playerShotActionsErr         error
	playerShotActionAreas        json.RawMessage
	playerShotActionAreasErr     error
	playerLastEvents             []sofascoreapi.EventSummary
	playerLastEventsErr          error
	playerFeaturedEvent          json.RawMessage
	playerFeaturedEventErr       error
	playerYearStats              json.RawMessage
	playerYearStatsErr           error
	sportLiveTournaments         json.RawMessage
	sportLiveTournamentsErr      error
	sportCategories              json.RawMessage
	sportCategoriesErr           error
	sportTopPlayers              json.RawMessage
	sportTopPlayersErr           error
}

func (f *fakeClient) Search(context.Context, string, int) ([]sofascoreapi.SearchResult, error) {
	return f.searchResults, f.searchErr
}

func (f *fakeClient) TeamEvents(context.Context, int, string, int) ([]sofascoreapi.EventSummary, error) {
	return f.events, f.eventsErr
}

func (f *fakeClient) Event(context.Context, int) (sofascoreapi.EventDetail, error) {
	return f.event, f.eventErr
}

func (f *fakeClient) EventTVChannels(context.Context, int) (json.RawMessage, error) {
	return f.eventTV, f.eventTVErr
}

func (f *fakeClient) EventTVChannelVotes(context.Context, int, int) (json.RawMessage, error) {
	return f.eventTVChannel, f.eventTVChannelErr
}

func (f *fakeClient) EventH2HEvents(context.Context, int) (sofascoreapi.EventH2HEventsResult, error) {
	return f.eventH2H, f.eventH2HErr
}

func (f *fakeClient) SportEvents(_ context.Context, sport, date string, limit int) ([]sofascoreapi.EventSummary, error) {
	f.sportEventsSport = sport
	f.sportEventsDate = date
	f.sportEventsLimit = limit
	return f.sportEvents, f.sportEventsErr
}

func (f *fakeClient) SportScheduledTournaments(_ context.Context, sport, date string, page int) ([]sofascoreapi.ScheduledTournamentSummary, bool, error) {
	f.sportScheduledSport = sport
	f.sportScheduledDate = date
	f.sportScheduledPage = page
	return f.sportScheduledTournaments, f.sportScheduledHasNext, f.sportScheduledTournamentsErr
}

func (f *fakeClient) DetectCountryAlpha2(context.Context) (string, error) {
	f.detectCountryCalls++
	return f.detectedCountry, f.detectCountryErr
}

func (f *fakeClient) TeamInfo(context.Context, int) (sofascoreapi.TeamInfoResult, error) {
	return f.teamInfo, f.teamInfoErr
}

func (f *fakeClient) TeamTournaments(context.Context, int) (json.RawMessage, error) {
	return f.teamTournaments, f.teamTournamentsErr
}

func (f *fakeClient) TeamStandings(context.Context, int, int) (sofascoreapi.TeamStandingsResult, error) {
	return f.teamStandings, f.teamStandingsErr
}

func (f *fakeClient) TeamTournamentStatistics(context.Context, int, int, int) (json.RawMessage, error) {
	return f.teamStats, f.teamStatsErr
}

func (f *fakeClient) TeamFeaturedPlayers(context.Context, int) (json.RawMessage, error) {
	return f.teamPlayers, f.teamPlayersErr
}

func (f *fakeClient) TeamMediaVideos(context.Context, int) (json.RawMessage, error) {
	return f.teamMedia, f.teamMediaErr
}

func (f *fakeClient) TeamRankings(context.Context, int) (json.RawMessage, error) {
	return f.teamRankings, f.teamRankingsErr
}

func (f *fakeClient) TeamTournamentRanks(context.Context, int, int, int) (json.RawMessage, error) {
	return f.teamTournamentRanks, f.teamTournamentRanksErr
}

func (f *fakeClient) TeamTournamentTopPlayers(context.Context, int, int, int) (json.RawMessage, error) {
	return f.teamTopPlayers, f.teamTopPlayersErr
}

func (f *fakeClient) PlayerAttributeOverviews(context.Context, int) (json.RawMessage, error) {
	return f.playerAttributes, f.playerAttributesErr
}

func (f *fakeClient) PlayerCharacteristics(context.Context, int) (json.RawMessage, error) {
	return f.playerCharacteristics, f.playerCharacteristicsErr
}

func (f *fakeClient) PlayerNationalTeamStatistics(context.Context, int) (json.RawMessage, error) {
	return f.playerNationalTeamStats, f.playerNationalTeamStatsErr
}

func (f *fakeClient) PlayerMedia(context.Context, int) (json.RawMessage, error) {
	return f.playerMedia, f.playerMediaErr
}

func (f *fakeClient) PlayerMediaVideos(context.Context, int) (json.RawMessage, error) {
	return f.playerMediaVideos, f.playerMediaVideosErr
}

func (f *fakeClient) PlayerUniqueTournaments(context.Context, int) (json.RawMessage, error) {
	return f.playerTournaments, f.playerTournamentsErr
}

func (f *fakeClient) PlayerStatisticsSeasons(context.Context, int) (json.RawMessage, error) {
	return f.playerSeasons, f.playerSeasonsErr
}

func (f *fakeClient) PlayerSeasonStatistics(context.Context, int, int, int, string) (json.RawMessage, error) {
	return f.playerSeasonStats, f.playerSeasonStatsErr
}

func (f *fakeClient) PlayerSeasonRatings(context.Context, int, int, int, string) (json.RawMessage, error) {
	return f.playerSeasonRatings, f.playerSeasonRatingsErr
}

func (f *fakeClient) PlayerSeasonHeatmap(context.Context, int, int, int, string) (json.RawMessage, error) {
	return f.playerSeasonHeatmap, f.playerSeasonHeatmapErr
}

func (f *fakeClient) PlayerPenaltyHistory(context.Context, int, int, int) (json.RawMessage, error) {
	return f.playerPenaltyHistory, f.playerPenaltyHistoryErr
}

func (f *fakeClient) PlayerCareerStatistics(context.Context, int) (json.RawMessage, error) {
	return f.playerCareer, f.playerCareerErr
}

func (f *fakeClient) PlayerCareerStatisticsMatchType(context.Context, int, string) (json.RawMessage, error) {
	return f.playerCareerMatchType, f.playerCareerMatchTypeErr
}

func (f *fakeClient) PlayerShotActions(context.Context, int, int, int, string) (json.RawMessage, error) {
	return f.playerShotActions, f.playerShotActionsErr
}

func (f *fakeClient) PlayerShotActionAreas(context.Context, int, int, string) (json.RawMessage, error) {
	return f.playerShotActionAreas, f.playerShotActionAreasErr
}

func (f *fakeClient) PlayerLastEvents(context.Context, int, int) ([]sofascoreapi.EventSummary, error) {
	return f.playerLastEvents, f.playerLastEventsErr
}

func (f *fakeClient) PlayerFeaturedEvent(context.Context, int) (json.RawMessage, error) {
	return f.playerFeaturedEvent, f.playerFeaturedEventErr
}

func (f *fakeClient) PlayerStatisticsSeasonsTennis(context.Context, int) (json.RawMessage, error) {
	return f.playerSeasonsTennis, f.playerSeasonsTennisErr
}

func (f *fakeClient) PlayerYearStatistics(context.Context, int, int) (json.RawMessage, error) {
	return f.playerYearStats, f.playerYearStatsErr
}

func (f *fakeClient) SportLiveTournaments(context.Context, string) (json.RawMessage, error) {
	return f.sportLiveTournaments, f.sportLiveTournamentsErr
}

func (f *fakeClient) SportCategories(context.Context, string) (json.RawMessage, error) {
	return f.sportCategories, f.sportCategoriesErr
}

func (f *fakeClient) SportTrendingTopPlayers(context.Context, string) (json.RawMessage, error) {
	return f.sportTopPlayers, f.sportTopPlayersErr
}

func (f *fakeClient) TournamentScheduledEvents(_ context.Context, tournamentID int, date string, limit int) ([]sofascoreapi.EventSummary, error) {
	f.tournamentScheduledID = tournamentID
	f.tournamentScheduledDate = date
	f.tournamentScheduledLimit = limit
	return f.tournamentScheduledEvents, f.tournamentScheduledEventsErr
}

func (f *fakeClient) TrendingEvents(_ context.Context, country string, limit int) ([]sofascoreapi.TrendingEventSummary, error) {
	f.trendingCountry = country
	f.trendingLimit = limit
	return f.trendingEvents, f.trendingErr
}

func (f *fakeClient) EventSection(_ context.Context, _ int, section string) (json.RawMessage, error) {
	if err := f.sectionErrs[section]; err != nil {
		return nil, err
	}
	return f.eventSections[section], nil
}

func (f *fakeClient) ProbeEventSections(context.Context, int) ([]string, error) {
	return f.availableSections, f.probeErr
}

func (f *fakeClient) Sports(context.Context) ([]sofascoreapi.SportCount, error) {
	return f.sports, f.sportsErr
}

func (f *fakeClient) SportSections(context.Context, string) (sofascoreapi.SportSectionDiscovery, error) {
	return f.discovery, f.discoveryErr
}

func (f *fakeClient) Tournament(context.Context, int) (sofascoreapi.TournamentDetail, error) {
	return f.tournament, f.tournamentErr
}

func (f *fakeClient) TournamentSeasons(context.Context, int) ([]sofascoreapi.TournamentSeason, error) {
	return f.tournamentSeasons, f.tournamentSeasonsErr
}

func (f *fakeClient) TournamentSection(_ context.Context, _, _ int, section string) (json.RawMessage, error) {
	if err := f.tournamentSectionErrs[section]; err != nil {
		return nil, err
	}
	return f.tournamentSections[section], nil
}

func (f *fakeClient) ProbeTournamentSections(context.Context, int, int) ([]string, error) {
	return f.tournamentAvailableSections, f.tournamentProbeErr
}

func (f *fakeClient) TournamentEvents(context.Context, int, int, string, int, string, int) ([]sofascoreapi.EventSummary, error) {
	return f.tournamentEvents, f.tournamentEventsErr
}

func (f *fakeClient) WatchEvents(_ context.Context, eventIDs []int, sections []string, allSections bool, emit func(sofascoreapi.WatchRecord) error) error {
	f.watchEventIDs = append([]int(nil), eventIDs...)
	f.watchEventSections = append([]string(nil), sections...)
	f.watchEventAllSections = allSections
	for _, record := range f.watchEventRecords {
		if err := emit(record); err != nil {
			return err
		}
	}
	return f.watchEventsErr
}

func (f *fakeClient) WatchSports(_ context.Context, sports []string, emit func(sofascoreapi.WatchRecord) error) error {
	f.watchSportsArgs = append([]string(nil), sports...)
	for _, record := range f.watchSportRecords {
		if err := emit(record); err != nil {
			return err
		}
	}
	return f.watchSportsErr
}

func connectSession(t *testing.T, svc lookups.Service) *mcp.ClientSession {
	return connectSessionWithOptions(t, svc, nil)
}

func connectSessionWithOptions(t *testing.T, svc lookups.Service, opts *mcp.ClientOptions) *mcp.ClientSession {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	server := New(svc)
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1.0.0"}, opts)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	t.Cleanup(func() {
		session.Close()
		cancel()
	})
	return session
}

func decodeStructured[T any](t *testing.T, value any) T {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var decoded T
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	return decoded
}

func TestListTools(t *testing.T) {
	session := connectSession(t, &fakeClient{})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}

	var names []string
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)

	want := []string{
		"event",
		"event_h2h_events",
		"event_tv",
		"event_tv_channel",
		"player_attributes",
		"player_career",
		"player_characteristics",
		"player_events_last",
		"player_featured_event",
		"player_media",
		"player_media_videos",
		"player_national_team_stats",
		"player_penalty_history",
		"player_season_heatmap",
		"player_season_ratings",
		"player_season_stats",
		"player_seasons",
		"player_shot_action_areas",
		"player_shot_actions",
		"player_tournaments",
		"player_year_stats",
		"search",
		"sport_categories",
		"sport_live_tournaments",
		"sport_top_players",
		"sports",
		"sports_events",
		"sports_tournaments",
		"team_events",
		"team_info",
		"team_media",
		"team_players",
		"team_rankings",
		"team_standings",
		"team_stats",
		"team_top_players",
		"team_tournaments",
		"tournament",
		"tournament_events",
		"tournament_scheduled_events",
		"tournament_seasons",
		"trending",
	}
	if !slices.Equal(names, want) {
		t.Fatalf("unexpected tools: got %v want %v", names, want)
	}
}

func TestSearchAndEventHappyPath(t *testing.T) {
	session := connectSession(t, &fakeClient{
		searchResults: []sofascoreapi.SearchResult{
			{Type: "team", ID: 2693, Name: "Fiorentina", Sport: "football"},
		},
		event: sofascoreapi.EventDetail{
			EventID:    13981704,
			StartTime:  time.Unix(1773690300, 0).UTC(),
			Home:       "Cremonese",
			Away:       "Fiorentina",
			Tournament: "Serie A",
			Sport:      "football",
			Raw:        json.RawMessage(`{"event":{"id":13981704}}`),
		},
		availableSections: []string{"statistics"},
	})

	searchResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "search",
		Arguments: SearchInput{Query: "fiorentina", Limit: 1},
	})
	if err != nil {
		t.Fatalf("search tool failed: %v", err)
	}
	if searchResult.IsError {
		t.Fatalf("search tool returned error result: %+v", searchResult)
	}
	search := decodeStructured[lookups.SearchResponse](t, searchResult.StructuredContent)
	if len(search.Results) != 1 || search.Results[0].ID != 2693 {
		t.Fatalf("unexpected search response: %+v", search)
	}

	eventResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "event",
		Arguments: EventInput{EventID: 13981704},
	})
	if err != nil {
		t.Fatalf("event tool failed: %v", err)
	}
	if eventResult.IsError {
		t.Fatalf("event tool returned error result: %+v", eventResult)
	}
	event := decodeStructured[lookups.EventResponse](t, eventResult.StructuredContent)
	if event.EventID != 13981704 || event.Home != "Cremonese" || event.Away != "Fiorentina" {
		t.Fatalf("unexpected event response: %+v", event)
	}
}

func TestValidationFailureReturnsToolError(t *testing.T) {
	session := connectSession(t, &fakeClient{})

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "team_events",
		Arguments: TeamEventsInput{TeamID: 2693},
	})
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error result")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "pass exactly one of --next or --last") {
		t.Fatalf("unexpected tool error text: %q", text)
	}
}

func TestSportsEventsAndTrendingHappyPath(t *testing.T) {
	client := &fakeClient{
		sportEvents: []sofascoreapi.EventSummary{
			{EventID: 11, StartTime: time.Date(2026, 3, 24, 18, 0, 0, 0, time.UTC), Home: "Torino", Away: "Parma"},
		},
		sportScheduledTournaments: []sofascoreapi.ScheduledTournamentSummary{
			{UniqueTournamentID: 696, UniqueTournament: "UEFA Women's Champions League", Name: "Knockout stage", UTCEventCount: 2},
		},
		sportScheduledHasNext: true,
		detectedCountry:       "dk",
		trendingEvents: []sofascoreapi.TrendingEventSummary{
			{Rank: 1, EventID: 21, StartTime: time.Date(2026, 3, 24, 19, 0, 0, 0, time.UTC), Sport: "football", Home: "A", Away: "B"},
		},
		tournamentScheduledEvents: []sofascoreapi.EventSummary{
			{EventID: 15471604, StartTime: time.Date(2026, 3, 24, 17, 45, 0, 0, time.UTC), Home: "Wolfsburg", Away: "Lyonnes"},
		},
	}
	session := connectSession(t, client)

	sportsEventsResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "sports_events",
		Arguments: SportsEventsInput{Sport: "football", Date: "2026-03-24", Limit: 5},
	})
	if err != nil {
		t.Fatalf("sports_events tool failed: %v", err)
	}
	if sportsEventsResult.IsError {
		t.Fatalf("sports_events tool returned error result: %+v", sportsEventsResult)
	}
	sportsEvents := decodeStructured[lookups.SportEventsResponse](t, sportsEventsResult.StructuredContent)
	if sportsEvents.Sport != "football" || sportsEvents.Date != "2026-03-24" || len(sportsEvents.Events) != 1 {
		t.Fatalf("unexpected sports_events response: %+v", sportsEvents)
	}

	sportsTournamentsResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "sports_tournaments",
		Arguments: SportsTournamentsInput{Sport: "football", Date: "2026-03-24", Page: 2},
	})
	if err != nil {
		t.Fatalf("sports_tournaments tool failed: %v", err)
	}
	if sportsTournamentsResult.IsError {
		t.Fatalf("sports_tournaments tool returned error result: %+v", sportsTournamentsResult)
	}
	sportsTournaments := decodeStructured[lookups.SportScheduledTournamentsResponse](t, sportsTournamentsResult.StructuredContent)
	if client.sportScheduledSport != "football" || client.sportScheduledDate != "2026-03-24" || client.sportScheduledPage != 2 {
		t.Fatalf("unexpected sports_tournaments args: sport=%q date=%q page=%d", client.sportScheduledSport, client.sportScheduledDate, client.sportScheduledPage)
	}
	if !sportsTournaments.HasNextPage || len(sportsTournaments.Tournaments) != 1 || sportsTournaments.Tournaments[0].UniqueTournamentID != 696 {
		t.Fatalf("unexpected sports_tournaments response: %+v", sportsTournaments)
	}

	trendingResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trending",
		Arguments: TrendingInput{Date: "2026-03-24", Limit: 5},
	})
	if err != nil {
		t.Fatalf("trending tool failed: %v", err)
	}
	if trendingResult.IsError {
		t.Fatalf("trending tool returned error result: %+v", trendingResult)
	}
	trending := decodeStructured[lookups.TrendingResponse](t, trendingResult.StructuredContent)
	if client.detectCountryCalls != 1 || client.trendingCountry != "DK" || client.trendingLimit != 0 {
		t.Fatalf("unexpected trending args: detectCalls=%d country=%q limit=%d", client.detectCountryCalls, client.trendingCountry, client.trendingLimit)
	}
	if trending.Country != "DK" || trending.Date != "2026-03-24" || len(trending.Events) != 1 || trending.Events[0].Rank != 1 {
		t.Fatalf("unexpected trending response: %+v", trending)
	}

	tournamentScheduledResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "tournament_scheduled_events",
		Arguments: TournamentScheduledEventsInput{TournamentID: 696, Date: "2026-03-24", Limit: 5},
	})
	if err != nil {
		t.Fatalf("tournament_scheduled_events tool failed: %v", err)
	}
	if tournamentScheduledResult.IsError {
		t.Fatalf("tournament_scheduled_events tool returned error result: %+v", tournamentScheduledResult)
	}
	tournamentScheduled := decodeStructured[lookups.TournamentScheduledEventsResponse](t, tournamentScheduledResult.StructuredContent)
	if client.tournamentScheduledID != 696 || client.tournamentScheduledDate != "2026-03-24" || client.tournamentScheduledLimit != 5 {
		t.Fatalf("unexpected tournament_scheduled_events args: id=%d date=%q limit=%d", client.tournamentScheduledID, client.tournamentScheduledDate, client.tournamentScheduledLimit)
	}
	if len(tournamentScheduled.Events) != 1 || tournamentScheduled.Events[0].EventID != 15471604 {
		t.Fatalf("unexpected tournament_scheduled_events response: %+v", tournamentScheduled)
	}
}

func TestNewLookupToolsHappyPath(t *testing.T) {
	session := connectSession(t, &fakeClient{
		eventTV:               json.RawMessage(`{"countryChannels":{"DK":[4024]}}`),
		eventTVChannel:        json.RawMessage(`{"tvChannelVotes":{"tvChannel":{"id":263,"name":"Viaplay"},"upvote":4,"downvote":0}}`),
		eventH2H:              sofascoreapi.EventH2HEventsResult{Slug: "utbsCtb", Raw: json.RawMessage(`{"events":[{"id":14442088}]}`)},
		teamInfo:              sofascoreapi.TeamInfoResult{Team: json.RawMessage(`{"team":{"id":3419}}`), FeaturedEvent: json.RawMessage(`{"featuredEvent":{"id":14442088}}`), Performance: json.RawMessage(`{"points":10}`), Tournaments: json.RawMessage(`{"uniqueTournaments":[]}`)},
		teamTournaments:       json.RawMessage(`{"uniqueTournaments":[{"id":132}]}`),
		teamStandings:         sofascoreapi.TeamStandingsResult{Season: sofascoreapi.TeamStandingsSeason{SeasonID: 80229}, Raw: json.RawMessage(`{"standings":[]}`)},
		teamStats:             json.RawMessage(`{"statistics":{"matches":10}}`),
		teamPlayers:           json.RawMessage(`{"featuredPlayers":[]}`),
		teamMedia:             json.RawMessage(`{"videos":[]}`),
		teamTournamentRanks:   json.RawMessage(`{"ranks":[]}`),
		teamTopPlayers:        json.RawMessage(`{"topPlayers":[]}`),
		playerAttributes:      json.RawMessage(`{"playerAttributeOverviews":[]}`),
		playerCareer:          json.RawMessage(`{"statistics":[]}`),
		playerCareerMatchType: json.RawMessage(`{"statistics":[]}`),
		playerCharacteristics: json.RawMessage(`{"items":[]}`),
		playerLastEvents: []sofascoreapi.EventSummary{
			{EventID: 100, StartTime: time.Date(2026, 3, 24, 19, 0, 0, 0, time.UTC), Home: "A", Away: "B"},
		},
		playerFeaturedEvent:     json.RawMessage(`{"featuredEvent":{"id":100}}`),
		playerMedia:             json.RawMessage(`{"media":[]}`),
		playerMediaVideos:       json.RawMessage(`{"videos":[]}`),
		playerNationalTeamStats: json.RawMessage(`{"statistics":[]}`),
		playerPenaltyHistory:    json.RawMessage(`{"penalties":[]}`),
		playerSeasonHeatmap:     json.RawMessage(`{"heatmap":[]}`),
		playerSeasonRatings:     json.RawMessage(`{"ratings":[]}`),
		playerSeasonStats:       json.RawMessage(`{"statistics":[]}`),
		playerSeasons:           json.RawMessage(`{"seasons":[]}`),
		playerShotActionAreas:   json.RawMessage(`{"areas":[]}`),
		playerShotActions:       json.RawMessage(`{"shotmap":[]}`),
		playerTournaments:       json.RawMessage(`{"uniqueTournaments":[]}`),
		playerYearStats:         json.RawMessage(`{"statistics":[]}`),
		sportLiveTournaments:    json.RawMessage(`{"liveTournaments":[]}`),
		sportCategories:         json.RawMessage(`{"categories":[]}`),
		sportTopPlayers:         json.RawMessage(`{"topPlayers":[]}`),
	})

	calls := []struct {
		name string
		args any
	}{
		{name: "event_tv", args: EventTVInput{EventID: 14442088}},
		{name: "event_tv_channel", args: EventTVChannelInput{EventID: 14442088, ChannelID: 263}},
		{name: "event_h2h_events", args: EventH2HEventsInput{EventID: 14442088}},
		{name: "team_info", args: TeamInfoInput{TeamID: 3419}},
		{name: "team_tournaments", args: TeamTournamentsInput{TeamID: 3419}},
		{name: "team_standings", args: TeamStandingsInput{TeamID: 3419}},
		{name: "team_stats", args: TeamStatsInput{TeamID: 3419, TournamentID: 132, SeasonID: 80229}},
		{name: "team_players", args: TeamPlayersInput{TeamID: 3419}},
		{name: "team_media", args: TeamMediaInput{TeamID: 3419}},
		{name: "team_rankings", args: TeamRankingsInput{TeamID: 3419, TournamentID: 132, SeasonID: 80229}},
		{name: "team_top_players", args: TeamTopPlayersInput{TeamID: 3419, TournamentID: 132, SeasonID: 80229}},
		{name: "player_attributes", args: PlayerAttributesInput{PlayerID: 829022}},
		{name: "player_media", args: PlayerMediaInput{PlayerID: 829022}},
		{name: "player_media_videos", args: PlayerMediaVideosInput{PlayerID: 829022}},
		{name: "player_events_last", args: PlayerLastEventsInput{PlayerID: 829022, Limit: 5}},
		{name: "player_seasons", args: PlayerSeasonsInput{PlayerID: 829022}},
		{name: "player_career", args: PlayerCareerInput{PlayerID: 829022}},
		{name: "player_season_stats", args: PlayerSeasonStatsInput{PlayerID: 817181, TournamentID: 132, SeasonID: 80229}},
		{name: "player_season_ratings", args: PlayerSeasonRatingsInput{PlayerID: 817181, TournamentID: 132, SeasonID: 80229}},
		{name: "player_characteristics", args: PlayerCharacteristicsInput{PlayerID: 829022}},
		{name: "player_national_team_stats", args: PlayerNationalTeamStatsInput{PlayerID: 829022}},
		{name: "player_tournaments", args: PlayerTournamentsInput{PlayerID: 829022}},
		{name: "player_season_heatmap", args: PlayerSeasonHeatmapInput{PlayerID: 829022, TournamentID: 23, SeasonID: 76457, Phase: "overall"}},
		{name: "player_penalty_history", args: PlayerPenaltyHistoryInput{PlayerID: 829022, TournamentID: 23, SeasonID: 76457}},
		{name: "player_shot_actions", args: PlayerShotActionsInput{PlayerID: 817181, TournamentID: 132, SeasonID: 80229, Phase: "regularSeason"}},
		{name: "player_shot_action_areas", args: PlayerShotActionAreasInput{PlayerID: 817181, TournamentID: 132, SeasonID: 80229, Phase: "regularSeason"}},
		{name: "player_year_stats", args: PlayerYearStatsInput{PlayerID: 206570, Year: 2026}},
		{name: "player_featured_event", args: PlayerFeaturedEventInput{PlayerID: 206570}},
		{name: "sport_live_tournaments", args: SportLiveTournamentsInput{Sport: "football"}},
		{name: "sport_categories", args: SportCategoriesInput{Sport: "football"}},
		{name: "sport_top_players", args: SportTopPlayersInput{Sport: "football"}},
	}

	for _, call := range calls {
		result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      call.name,
			Arguments: call.args,
		})
		if err != nil {
			t.Fatalf("%s failed: %v", call.name, err)
		}
		if result.IsError {
			t.Fatalf("%s returned error result: %+v", call.name, result)
		}
		if call.name == "event_tv_channel" {
			response := decodeStructured[lookups.EventTVChannelResponse](t, result.StructuredContent)
			if response.ChannelName != "Viaplay" || response.ChannelID != 263 {
				t.Fatalf("unexpected event_tv_channel response: %+v", response)
			}
		}
	}
}

func TestPartialSectionErrorsAreReturned(t *testing.T) {
	session := connectSession(t, &fakeClient{
		event: sofascoreapi.EventDetail{
			EventID:   11,
			StartTime: time.Unix(1773690300, 0).UTC(),
			Home:      "Torino",
			Away:      "Parma",
			Sport:     "football",
			Raw:       json.RawMessage(`{"event":{"id":11}}`),
		},
		availableSections: []string{"statistics", "lineups"},
		eventSections: map[string]json.RawMessage{
			"statistics": json.RawMessage(`{"shots":10}`),
		},
		sectionErrs: map[string]error{
			"lineups": &sofascoreapi.HTTPStatusError{StatusCode: 404, URL: "x"},
		},
		tournament: sofascoreapi.TournamentDetail{
			TournamentID: 17,
			Name:         "Premier League",
			Sport:        "football",
			Raw:          json.RawMessage(`{"tournament":{"id":17}}`),
		},
		tournamentSeasons:           []sofascoreapi.TournamentSeason{{ID: 99, Name: "2025/26"}},
		tournamentAvailableSections: []string{"standings/total", "info"},
		tournamentSections: map[string]json.RawMessage{
			"info": json.RawMessage(`{"name":"Premier League"}`),
		},
		tournamentSectionErrs: map[string]error{
			"standings/total": &sofascoreapi.HTTPStatusError{StatusCode: 404, URL: "x"},
		},
	})

	eventResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "event",
		Arguments: EventInput{EventID: 11, Sections: []string{"statistics", "lineups"}},
	})
	if err != nil {
		t.Fatalf("event tool failed: %v", err)
	}
	if eventResult.IsError {
		t.Fatalf("event tool returned error result: %+v", eventResult)
	}
	event := decodeStructured[lookups.EventResponse](t, eventResult.StructuredContent)
	if !event.Partial || event.SectionErrors["lineups"] == "" {
		t.Fatalf("unexpected partial event response: %+v", event)
	}

	tournamentResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "tournament",
		Arguments: TournamentInput{TournamentID: 17, Sections: []string{"standings/total", "info"}},
	})
	if err != nil {
		t.Fatalf("tournament tool failed: %v", err)
	}
	if tournamentResult.IsError {
		t.Fatalf("tournament tool returned error result: %+v", tournamentResult)
	}
	tournament := decodeStructured[lookups.TournamentResponse](t, tournamentResult.StructuredContent)
	if !tournament.Partial || tournament.SectionErrors["standings/total"] == "" {
		t.Fatalf("unexpected partial tournament response: %+v", tournament)
	}
}

func TestListResourceTemplatesIncludesLiveResources(t *testing.T) {
	session := connectSession(t, &fakeClient{})

	result, err := session.ListResourceTemplates(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListResourceTemplates returned error: %v", err)
	}

	var templates []string
	for _, item := range result.ResourceTemplates {
		templates = append(templates, item.URITemplate)
	}
	slices.Sort(templates)

	want := []string{
		"sports://live/event/{event_id}{?section,all_sections}",
		"sports://live/events/{event_ids}{?section,all_sections}",
		"sports://live/sport/{sport}",
		"sports://live/sports/{sports}",
	}
	if !slices.Equal(templates, want) {
		t.Fatalf("unexpected resource templates: got %v want %v", templates, want)
	}
}

func TestEventLiveResourceSubscribeAndRead(t *testing.T) {
	updates := make(chan string, 8)
	client := &fakeClient{
		watchEventRecords: []sofascoreapi.WatchRecord{
			{
				Type:      sofascoreapi.WatchRecordSnapshot,
				WatchKind: sofascoreapi.WatchKindEvent,
				EventID:   15636234,
				Sport:     "football",
				At:        "2026-03-24T10:00:00Z",
				Summary: &sofascoreapi.WatchEventSummary{
					EventID:           15636234,
					Sport:             "football",
					Home:              "Torino",
					Away:              "Parma",
					StatusDescription: "1st half",
				},
				Sections: map[string]any{
					"statistics": map[string]any{"shots": 10},
				},
			},
			{
				Type:          sofascoreapi.WatchRecordUpdate,
				WatchKind:     sofascoreapi.WatchKindEvent,
				EventID:       15636234,
				Sport:         "football",
				At:            "2026-03-24T10:00:05Z",
				ChangedFields: []string{"status.description"},
				Patch: map[string]any{
					"status.description": "HT",
				},
				Summary: &sofascoreapi.WatchEventSummary{
					EventID:           15636234,
					Sport:             "football",
					Home:              "Torino",
					Away:              "Parma",
					StatusDescription: "HT",
				},
			},
			{
				Type:      sofascoreapi.WatchRecordStatus,
				WatchKind: sofascoreapi.WatchKindEvent,
				At:        "2026-03-24T10:00:06Z",
				State:     "connected",
			},
		},
	}
	session := connectSessionWithOptions(t, client, &mcp.ClientOptions{
		ResourceUpdatedHandler: func(_ context.Context, req *mcp.ResourceUpdatedNotificationRequest) {
			updates <- req.Params.URI
		},
	})

	uri := "sports://live/event/15636234?section=statistics"
	if err := session.Subscribe(context.Background(), &mcp.SubscribeParams{URI: uri}); err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}

	select {
	case got := <-updates:
		if got != uri {
			t.Fatalf("unexpected update URI %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected resource update notification")
	}

	if got, want := client.watchEventIDs, []int{15636234}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("unexpected watch event ids: got %v want %v", got, want)
	}
	if got, want := client.watchEventSections, []string{"statistics"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("unexpected watch sections: got %v want %v", got, want)
	}

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri})
	if err != nil {
		t.Fatalf("ReadResource returned error: %v", err)
	}
	if len(result.Contents) != 1 || result.Contents[0].Text == "" {
		t.Fatalf("unexpected resource contents: %+v", result)
	}

	var state map[string]any
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &state); err != nil {
		t.Fatalf("resource is not valid JSON: %v", err)
	}
	if state["watch_kind"] != string(sofascoreapi.WatchKindEvent) {
		t.Fatalf("unexpected watch kind: %+v", state)
	}
	if state["connection_state"] != "connected" {
		t.Fatalf("unexpected connection state: %+v", state)
	}
}

func TestSportsLiveResourceSubscribeAndRead(t *testing.T) {
	client := &fakeClient{
		watchSportRecords: []sofascoreapi.WatchRecord{
			{
				Type:      sofascoreapi.WatchRecordSnapshot,
				WatchKind: sofascoreapi.WatchKindSport,
				Sport:     "football",
				At:        "2026-03-24T10:00:00Z",
				Events: []sofascoreapi.WatchEventSummary{
					{EventID: 15636234, Home: "Torino", Away: "Parma", StatusDescription: "1st half"},
				},
			},
			{
				Type:      sofascoreapi.WatchRecordSnapshot,
				WatchKind: sofascoreapi.WatchKindSport,
				Sport:     "tennis",
				At:        "2026-03-24T10:00:01Z",
				Events: []sofascoreapi.WatchEventSummary{
					{EventID: 15855340, Home: "Player A", Away: "Player B", StatusDescription: "Set 1"},
				},
			},
		},
	}
	session := connectSession(t, client)

	uri := "sports://live/sports/football,tennis"
	if err := session.Subscribe(context.Background(), &mcp.SubscribeParams{URI: uri}); err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for len(client.watchSportsArgs) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if got, want := client.watchSportsArgs, []string{"football", "tennis"}; !slices.Equal(got, want) {
		t.Fatalf("unexpected sports args: got %v want %v", got, want)
	}

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri})
	if err != nil {
		t.Fatalf("ReadResource returned error: %v", err)
	}

	var state map[string]any
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &state); err != nil {
		t.Fatalf("resource is not valid JSON: %v", err)
	}
	if state["watch_kind"] != string(sofascoreapi.WatchKindSport) {
		t.Fatalf("unexpected watch kind: %+v", state)
	}
	sportStates, ok := state["sport_states"].(map[string]any)
	if !ok || sportStates["football"] == nil || sportStates["tennis"] == nil {
		t.Fatalf("unexpected sport state payload: %+v", state)
	}
}
