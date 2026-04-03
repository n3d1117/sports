package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"sports/internal/buildinfo"
	"sports/internal/provider/sofascore"
)

type fakeClient struct {
	searchResults                []sofascoreapi.SearchResult
	searchErr                    error
	searchPage                   int
	events                       []sofascoreapi.EventSummary
	eventsErr                    error
	teamEventsCalls              int
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
	sportScheduledTournaments    []sofascoreapi.ScheduledTournamentSummary
	sportScheduledTournamentsErr error
	sportScheduledSport          string
	sportScheduledDate           string
	sportScheduledPage           int
	sportScheduledHasNext        bool
	detectedCountry              string
	detectCountryErr             error
	trendingEvents               []sofascoreapi.TrendingEventSummary
	trendingErr                  error
	tournamentScheduledEvents    []sofascoreapi.EventSummary
	tournamentScheduledEventsErr error
	tournamentScheduledID        int
	tournamentScheduledDate      string
	tournamentScheduledLimit     int
	eventSections                map[string]json.RawMessage
	sectionErr                   error
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
	tournamentSectionErr         error
	tournamentSectionErrs        map[string]error
	tournamentAvailableSections  []string
	tournamentProbeErr           error
	tournamentEvents             []sofascoreapi.EventSummary
	tournamentEventsErr          error
	tournamentMode               string
	tournamentRound              int
	tournamentSlug               string
	tournamentLimit              int
	tournamentSeasonID           int
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
	playerTournamentID           int
	playerSeasonID               int
	playerPhase                  string
	playerLimit                  int
	playerYear                   int
	sportLiveTournaments         json.RawMessage
	sportLiveTournamentsErr      error
	sportCategories              json.RawMessage
	sportCategoriesErr           error
	sportTopPlayers              json.RawMessage
	sportTopPlayersErr           error
}

func (f *fakeClient) Search(_ context.Context, _ string, page int) ([]sofascoreapi.SearchResult, error) {
	f.searchPage = page
	return f.searchResults, f.searchErr
}

func (f *fakeClient) TeamEvents(context.Context, int, string, int) ([]sofascoreapi.EventSummary, error) {
	f.teamEventsCalls++
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

func (f *fakeClient) SportEvents(_ context.Context, sport, date string, _ int) ([]sofascoreapi.EventSummary, error) {
	f.sportEventsSport = sport
	f.sportEventsDate = date
	return f.sportEvents, f.sportEventsErr
}

func (f *fakeClient) SportScheduledTournaments(_ context.Context, sport, date string, page int) ([]sofascoreapi.ScheduledTournamentSummary, bool, error) {
	f.sportScheduledSport = sport
	f.sportScheduledDate = date
	f.sportScheduledPage = page
	return f.sportScheduledTournaments, f.sportScheduledHasNext, f.sportScheduledTournamentsErr
}

func (f *fakeClient) DetectCountryAlpha2(context.Context) (string, error) {
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

func (f *fakeClient) PlayerSeasonStatistics(_ context.Context, _ int, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	f.playerTournamentID = tournamentID
	f.playerSeasonID = seasonID
	f.playerPhase = phase
	return f.playerSeasonStats, f.playerSeasonStatsErr
}

func (f *fakeClient) PlayerSeasonRatings(_ context.Context, _ int, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	f.playerTournamentID = tournamentID
	f.playerSeasonID = seasonID
	f.playerPhase = phase
	return f.playerSeasonRatings, f.playerSeasonRatingsErr
}

func (f *fakeClient) PlayerSeasonHeatmap(_ context.Context, _ int, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	f.playerTournamentID = tournamentID
	f.playerSeasonID = seasonID
	f.playerPhase = phase
	return f.playerSeasonHeatmap, f.playerSeasonHeatmapErr
}

func (f *fakeClient) PlayerPenaltyHistory(_ context.Context, _ int, tournamentID, seasonID int) (json.RawMessage, error) {
	f.playerTournamentID = tournamentID
	f.playerSeasonID = seasonID
	return f.playerPenaltyHistory, f.playerPenaltyHistoryErr
}

func (f *fakeClient) PlayerCareerStatistics(context.Context, int) (json.RawMessage, error) {
	return f.playerCareer, f.playerCareerErr
}

func (f *fakeClient) PlayerCareerStatisticsMatchType(context.Context, int, string) (json.RawMessage, error) {
	return f.playerCareerMatchType, f.playerCareerMatchTypeErr
}

func (f *fakeClient) PlayerShotActions(_ context.Context, _ int, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	f.playerTournamentID = tournamentID
	f.playerSeasonID = seasonID
	f.playerPhase = phase
	return f.playerShotActions, f.playerShotActionsErr
}

func (f *fakeClient) PlayerShotActionAreas(_ context.Context, tournamentID, seasonID int, phase string) (json.RawMessage, error) {
	f.playerTournamentID = tournamentID
	f.playerSeasonID = seasonID
	f.playerPhase = phase
	return f.playerShotActionAreas, f.playerShotActionAreasErr
}

func (f *fakeClient) PlayerLastEvents(_ context.Context, _ int, limit int) ([]sofascoreapi.EventSummary, error) {
	f.playerLimit = limit
	return f.playerLastEvents, f.playerLastEventsErr
}

func (f *fakeClient) PlayerFeaturedEvent(context.Context, int) (json.RawMessage, error) {
	return f.playerFeaturedEvent, f.playerFeaturedEventErr
}

func (f *fakeClient) PlayerStatisticsSeasonsTennis(context.Context, int) (json.RawMessage, error) {
	return f.playerSeasonsTennis, f.playerSeasonsTennisErr
}

func (f *fakeClient) PlayerYearStatistics(_ context.Context, _ int, year int) (json.RawMessage, error) {
	f.playerYear = year
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

func (f *fakeClient) TrendingEvents(_ context.Context, _ string, _ int) ([]sofascoreapi.TrendingEventSummary, error) {
	return f.trendingEvents, f.trendingErr
}

func (f *fakeClient) WatchEvents(_ context.Context, eventIDs []int, sections []string, allSections bool, emit func(sofascoreapi.WatchRecord) error) error {
	f.watchEventIDs = append([]int(nil), eventIDs...)
	f.watchEventSections = append([]string(nil), sections...)
	f.watchEventAllSections = allSections
	if len(f.watchEventRecords) == 0 {
		f.watchEventRecords = []sofascoreapi.WatchRecord{{
			Type:      sofascoreapi.WatchRecordStatus,
			WatchKind: sofascoreapi.WatchKindEvent,
			State:     "closed",
		}}
	}
	for _, record := range f.watchEventRecords {
		if err := emit(record); err != nil {
			return err
		}
	}
	return f.watchEventsErr
}

func (f *fakeClient) WatchSports(_ context.Context, sports []string, emit func(sofascoreapi.WatchRecord) error) error {
	f.watchSportsArgs = append([]string(nil), sports...)
	if len(f.watchSportRecords) == 0 {
		f.watchSportRecords = []sofascoreapi.WatchRecord{{
			Type:      sofascoreapi.WatchRecordStatus,
			WatchKind: sofascoreapi.WatchKindSport,
			State:     "closed",
		}}
	}
	for _, record := range f.watchSportRecords {
		if err := emit(record); err != nil {
			return err
		}
	}
	return f.watchSportsErr
}

func (f *fakeClient) EventSection(_ context.Context, _ int, section string) (json.RawMessage, error) {
	if err := f.sectionErrs[section]; err != nil {
		return nil, err
	}
	if f.sectionErr != nil {
		return nil, f.sectionErr
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

func (f *fakeClient) TournamentSection(_ context.Context, _, seasonID int, section string) (json.RawMessage, error) {
	f.tournamentSeasonID = seasonID
	if err := f.tournamentSectionErrs[section]; err != nil {
		return nil, err
	}
	if f.tournamentSectionErr != nil {
		return nil, f.tournamentSectionErr
	}
	return f.tournamentSections[section], nil
}

func (f *fakeClient) ProbeTournamentSections(_ context.Context, _, seasonID int) ([]string, error) {
	f.tournamentSeasonID = seasonID
	return f.tournamentAvailableSections, f.tournamentProbeErr
}

func (f *fakeClient) TournamentEvents(_ context.Context, _, seasonID int, mode string, round int, slug string, limit int) ([]sofascoreapi.EventSummary, error) {
	f.tournamentSeasonID = seasonID
	f.tournamentMode = mode
	f.tournamentRound = round
	f.tournamentSlug = slug
	f.tournamentLimit = limit
	return f.tournamentEvents, f.tournamentEventsErr
}

func testApp(client *fakeClient, tty bool) (*App, *bytes.Buffer, *bytes.Buffer) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	return &App{
		Client:  client,
		Stdout:  &stdout,
		Stderr:  &stderr,
		Context: context.Background(),
		isTerminal: func(io.Writer) bool {
			return tty
		},
	}, &stdout, &stderr
}

func requireExitCode(t *testing.T, err error, code int) {
	t.Helper()
	var coded *exitError
	if !errors.As(err, &coded) {
		t.Fatalf("expected exitError, got %v", err)
	}
	if coded.Code != code {
		t.Fatalf("expected code %d, got %d", code, coded.Code)
	}
}

func TestRootHelpAndMissingCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{}, true)

	err := app.Execute(nil)
	requireExitCode(t, err, 2)
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "football") {
		t.Fatalf("expected football command in help, got %q", stdout.String())
	}
}

func TestVersionFlag(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{}, true)

	if err := app.Execute([]string{"--version"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if strings.TrimSpace(stdout.String()) != buildinfo.Current() {
		t.Fatalf("expected version %q, got %q", buildinfo.Current(), stdout.String())
	}
}

func TestSearchTextOutputOnTTY(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		searchResults: []sofascoreapi.SearchResult{
			{Type: "team", ID: 2693, Name: "Fiorentina", Sport: "football", Country: "Italy"},
		},
	}, true)

	if err := app.Execute([]string{"search", "fiorentina"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Fiorentina") {
		t.Fatalf("expected Fiorentina in output, got %q", stdout.String())
	}
}

func TestSearchIDsStayTextWhenNotTTY(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		searchResults: []sofascoreapi.SearchResult{
			{Type: "team", ID: 2693, Name: "Fiorentina", Sport: "football"},
		},
	}, false)

	if err := app.Execute([]string{"search", "fiorentina", "--id"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if stdout.String() != "2693\n" {
		t.Fatalf("expected ids-only output, got %q", stdout.String())
	}
}

func TestSearchAutoJSONWhenNotTTY(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		searchResults: []sofascoreapi.SearchResult{
			{Type: "team", ID: 2693, Name: "Fiorentina", Sport: "football"},
		},
	}, false)

	if err := app.Execute([]string{"search", "fiorentina"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "\"query\": \"fiorentina\"") {
		t.Fatalf("expected JSON output, got %q", stdout.String())
	}
}

func TestListFiltersSupportedSports(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		sports: []sofascoreapi.SportCount{
			{Slug: "football", Live: 3, Total: 12},
			{Slug: "rugby", Live: 2, Total: 9},
			{Slug: "tennis", Live: 1, Total: 5},
		},
	}, false)

	if err := app.Execute([]string{"list"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "\"football\"") || !strings.Contains(stdout.String(), "\"tennis\"") {
		t.Fatalf("expected supported sports in JSON, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "\"rugby\"") {
		t.Fatalf("expected unsupported sport to be filtered, got %q", stdout.String())
	}
}

func TestFootballBaseCommandUsesSportEvents(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		sportEvents: []sofascoreapi.EventSummary{
			{EventID: 1, StartTime: time.Unix(1, 0).UTC(), Home: "A", Away: "B", Tournament: "Serie A", Sport: "football"},
		},
	}, false)

	if err := app.Execute([]string{"football", "--date", "2026-03-24"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "\"sport\": \"football\"") {
		t.Fatalf("expected football JSON output, got %q", stdout.String())
	}
}

func TestFootballSectionsCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		discovery: sofascoreapi.SportSectionDiscovery{
			Sport:         "football",
			SampleEventID: 13981714,
			Sections:      []string{"statistics", "lineups"},
		},
	}, false)

	if err := app.Execute([]string{"football", "sections", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "\"sample_event_id\": 13981714") {
		t.Fatalf("expected sections JSON, got %q", stdout.String())
	}
}

func TestEventBaseJSON(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		event: sofascoreapi.EventDetail{
			EventID:           13981714,
			StartTime:         time.Unix(1773690300, 0).UTC(),
			StatusType:        "notstarted",
			StatusDescription: "Not started",
			Home:              "Cremonese",
			Away:              "Fiorentina",
			Tournament:        "Serie A",
			Sport:             "football",
			Raw:               json.RawMessage(`{"id":13981714}`),
		},
		availableSections: []string{"statistics", "lineups"},
	}, false)

	if err := app.Execute([]string{"event", "13981714"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "\"event_id\": 13981714") {
		t.Fatalf("expected event JSON, got %q", stdout.String())
	}
}

func TestEventSectionsTextOutput(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		event: sofascoreapi.EventDetail{
			EventID: 13981714,
			Sport:   "football",
			Raw:     json.RawMessage(`{"id":13981714}`),
		},
		availableSections: []string{"statistics", "h2h", "lineups"},
	}, true)

	if err := app.Execute([]string{"event", "13981714", "sections"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "sports event 13981714 section h2h --json") {
		t.Fatalf("missing updated example in output %q", out)
	}
}

func TestEventSectionCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		event: sofascoreapi.EventDetail{
			EventID: 13981714,
			Sport:   "football",
			Raw:     json.RawMessage(`{"id":13981714}`),
		},
		availableSections: []string{"statistics", "lineups"},
		eventSections: map[string]json.RawMessage{
			"statistics": json.RawMessage(`{"shots":10}`),
			"lineups":    json.RawMessage(`{"home":{}}`),
		},
	}, false)

	if err := app.Execute([]string{"event", "13981714", "section", "statistics", "lineups", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "\"statistics\"") || !strings.Contains(stdout.String(), "\"lineups\"") {
		t.Fatalf("expected fetched sections in JSON, got %q", stdout.String())
	}
}

func TestEventWatchCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		watchEventRecords: []sofascoreapi.WatchRecord{{
			Type:      sofascoreapi.WatchRecordStatus,
			WatchKind: sofascoreapi.WatchKindEvent,
			State:     "closed",
		}},
	}, false)

	if err := app.Execute([]string{"event", "15636234", "watch", "--section", "statistics", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"watch_kind":"event"`) {
		t.Fatalf("expected NDJSON output, got %q", stdout.String())
	}
}

func TestTeamNextCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		events: []sofascoreapi.EventSummary{
			{EventID: 1, StartTime: time.Unix(1, 0).UTC(), Home: "A", Away: "B", Tournament: "Serie A"},
		},
	}, false)

	if err := app.Execute([]string{"team", "2693", "next", "--limit", "1", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"direction": "next"`) {
		t.Fatalf("expected next direction JSON, got %q", stdout.String())
	}
}

func TestTournamentBaseCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		tournament: sofascoreapi.TournamentDetail{
			TournamentID: 17,
			Name:         "Premier League",
			Sport:        "football",
			Category:     "England",
			Country:      "England",
			Raw:          json.RawMessage(`{"id":17}`),
		},
		tournamentSeasons:           []sofascoreapi.TournamentSeason{{ID: 99, Name: "2025/26"}},
		tournamentSeasonID:          99,
		availableSections:           []string{"info"},
		tournamentAvailableSections: []string{"info"},
	}, false)

	if err := app.Execute([]string{"tournament", "17", "--season", "99"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"season_id": 99`) {
		t.Fatalf("expected season id in JSON, got %q", stdout.String())
	}
}

func TestTournamentSectionCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		tournament: sofascoreapi.TournamentDetail{
			TournamentID: 17,
			Name:         "Premier League",
			Sport:        "football",
			Category:     "England",
			Country:      "England",
			Raw:          json.RawMessage(`{"id":17}`),
		},
		tournamentSeasons: []sofascoreapi.TournamentSeason{{ID: 99, Name: "2025/26"}},
		tournamentSections: map[string]json.RawMessage{
			"info": json.RawMessage(`{"name":"Premier League"}`),
		},
		tournamentAvailableSections: []string{"info"},
	}, false)

	if err := app.Execute([]string{"tournament", "17", "section", "info", "--season", "99", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"info"`) {
		t.Fatalf("expected section payload in JSON, got %q", stdout.String())
	}
}

func TestTournamentRoundCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		tournamentSeasons: []sofascoreapi.TournamentSeason{{ID: 99, Name: "2025/26"}},
		tournamentEvents: []sofascoreapi.EventSummary{
			{EventID: 1, StartTime: time.Unix(1, 0).UTC(), Home: "A", Away: "B", Tournament: "Premier League"},
		},
	}, false)

	if err := app.Execute([]string{"tournament", "17", "round", "5", "--slug", "round-of-16", "--limit", "2", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"mode": "round"`) {
		t.Fatalf("expected round mode in JSON, got %q", stdout.String())
	}
}

func TestTournamentScheduledRejectsSeason(t *testing.T) {
	app, _, _ := testApp(&fakeClient{}, true)

	err := app.Execute([]string{"tournament", "17", "scheduled", "--season", "99"})
	requireExitCode(t, err, 2)
}

func TestEventTVCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		eventTV: json.RawMessage(`{"countryChannels":{"DK":[4024]}}`),
	}, false)

	if err := app.Execute([]string{"event", "14442088", "tv", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"event_id": 14442088`) || !strings.Contains(stdout.String(), `"countryChannels"`) {
		t.Fatalf("expected event tv JSON, got %q", stdout.String())
	}
}

func TestEventTVChannelCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		eventTVChannel: json.RawMessage(`{"tvChannelVotes":{"tvChannel":{"id":263,"name":"Viaplay"},"upvote":4,"downvote":0}}`),
	}, false)

	if err := app.Execute([]string{"event", "15697200", "tv-channel", "263", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"channel_id": 263`) || !strings.Contains(stdout.String(), `"channel_name": "Viaplay"`) {
		t.Fatalf("expected event tv-channel JSON, got %q", stdout.String())
	}
}

func TestTeamRankingsTournamentCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		teamTournamentRanks: json.RawMessage(`{"ranks":[{"name":"Points per game","value":1}]}`),
	}, false)

	if err := app.Execute([]string{"team", "3419", "rankings", "--tournament", "132", "--season", "80229", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"tournament_id": 132`) || !strings.Contains(stdout.String(), `"season_id": 80229`) {
		t.Fatalf("expected tournament rankings JSON, got %q", stdout.String())
	}
}

func TestPlayerAttributesCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		playerAttributes: json.RawMessage(`{"averageAttributeOverviews":[],"playerAttributeOverviews":[]}`),
	}, false)

	if err := app.Execute([]string{"player", "829022", "attributes", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"player_id": 829022`) {
		t.Fatalf("expected player attributes JSON, got %q", stdout.String())
	}
}

func TestPlayerSeasonStatsCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		playerSeasonStats: json.RawMessage(`{"statistics":{"matches":10}}`),
	}, false)

	if err := app.Execute([]string{"player", "817181", "season-stats", "--tournament", "132", "--season", "80229", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"tournament_id": 132`) || !strings.Contains(stdout.String(), `"season_id": 80229`) {
		t.Fatalf("expected player season stats JSON, got %q", stdout.String())
	}
}

func TestPlayerShotActionsCommandRequiresPhase(t *testing.T) {
	app, _, _ := testApp(&fakeClient{}, false)

	err := app.Execute([]string{"player", "817181", "shot-actions", "--tournament", "132", "--season", "80229"})
	requireExitCode(t, err, 2)
}

func TestPlayerYearStatsCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		playerYearStats: json.RawMessage(`{"statistics":{"wins":1}}`),
	}, false)

	if err := app.Execute([]string{"player", "206570", "year-stats", "--year", "2026", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"year": 2026`) {
		t.Fatalf("expected player year stats JSON, got %q", stdout.String())
	}
}

func TestFootballLiveTournamentsCommand(t *testing.T) {
	app, stdout, _ := testApp(&fakeClient{
		sportLiveTournaments: json.RawMessage(`{"liveTournaments":[{"id":17,"name":"Premier League"}]}`),
	}, false)

	if err := app.Execute([]string{"football", "live-tournaments", "--json"}); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"sport": "football"`) || !strings.Contains(stdout.String(), `"liveTournaments"`) {
		t.Fatalf("expected live tournaments JSON, got %q", stdout.String())
	}
}
