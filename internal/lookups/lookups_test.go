package lookups

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"sports/internal/provider/sofascore"
)

type fakeClient struct {
	searchResults                 []sofascoreapi.SearchResult
	searchErr                     error
	events                        []sofascoreapi.EventSummary
	eventsErr                     error
	event                         sofascoreapi.EventDetail
	eventErr                      error
	eventTV                       json.RawMessage
	eventTVErr                    error
	eventTVChannel                json.RawMessage
	eventTVChannelErr             error
	eventH2H                      sofascoreapi.EventH2HEventsResult
	eventH2HErr                   error
	sportEvents                   []sofascoreapi.EventSummary
	sportEventsErr                error
	sportEventsSport              string
	sportEventsDate               string
	sportEventsLimit              int
	sportScheduledTournaments     []sofascoreapi.ScheduledTournamentSummary
	sportScheduledTournamentsErr  error
	sportScheduledSport           string
	sportScheduledDate            string
	sportScheduledPage            int
	sportScheduledHasNext         bool
	detectedCountry               string
	detectCountryErr              error
	detectCountryCalls            int
	trendingEvents                []sofascoreapi.TrendingEventSummary
	trendingErr                   error
	trendingCountry               string
	trendingLimit                 int
	tournamentScheduledEvents     []sofascoreapi.EventSummary
	tournamentScheduledEventsErr  error
	tournamentScheduledID         int
	tournamentScheduledDate       string
	tournamentScheduledLimit      int
	eventSections                 map[string]json.RawMessage
	sectionErrs                   map[string]error
	availableSections             []string
	probeErr                      error
	sports                        []sofascoreapi.SportCount
	sportsErr                     error
	discovery                     sofascoreapi.SportSectionDiscovery
	discoveryErr                  error
	tournament                    sofascoreapi.TournamentDetail
	tournamentErr                 error
	tournamentSeasons             []sofascoreapi.TournamentSeason
	tournamentSeasonsErr          error
	tournamentSections            map[string]json.RawMessage
	tournamentSectionErrs         map[string]error
	tournamentAvailableSections   []string
	tournamentProbeErr            error
	tournamentEvents              []sofascoreapi.EventSummary
	tournamentEventsErr           error
	teamInfo                      sofascoreapi.TeamInfoResult
	teamInfoErr                   error
	teamTournaments               json.RawMessage
	teamTournamentsErr            error
	teamStandings                 sofascoreapi.TeamStandingsResult
	teamStandingsErr              error
	teamStats                     json.RawMessage
	teamStatsErr                  error
	teamPlayers                   json.RawMessage
	teamPlayersErr                error
	teamMedia                     json.RawMessage
	teamMediaErr                  error
	teamRankings                  json.RawMessage
	teamRankingsErr               error
	teamTournamentRanks           json.RawMessage
	teamTournamentRanksErr        error
	teamTopPlayers                json.RawMessage
	teamTopPlayersErr             error
	playerAttributes              json.RawMessage
	playerAttributesErr           error
	playerCharacteristics         json.RawMessage
	playerCharacteristicsErr      error
	playerNationalTeamStats       json.RawMessage
	playerNationalTeamStatsErr    error
	playerMedia                   json.RawMessage
	playerMediaErr                error
	playerMediaVideos             json.RawMessage
	playerMediaVideosErr          error
	playerTournaments             json.RawMessage
	playerTournamentsErr          error
	playerSeasons                 json.RawMessage
	playerSeasonsErr              error
	playerSeasonsTennis           json.RawMessage
	playerSeasonsTennisErr        error
	playerSeasonStats             json.RawMessage
	playerSeasonStatsErr          error
	playerSeasonStatsByPhase      map[string]json.RawMessage
	playerSeasonStatsErrByPhase   map[string]error
	playerSeasonRatings           json.RawMessage
	playerSeasonRatingsErr        error
	playerSeasonRatingsByPhase    map[string]json.RawMessage
	playerSeasonRatingsErrByPhase map[string]error
	playerSeasonHeatmap           json.RawMessage
	playerSeasonHeatmapErr        error
	playerPenaltyHistory          json.RawMessage
	playerPenaltyHistoryErr       error
	playerCareer                  json.RawMessage
	playerCareerErr               error
	playerCareerMatchType         json.RawMessage
	playerCareerMatchTypeErr      error
	playerShotActions             json.RawMessage
	playerShotActionsErr          error
	playerShotActionAreas         json.RawMessage
	playerShotActionAreasErr      error
	playerLastEvents              []sofascoreapi.EventSummary
	playerLastEventsErr           error
	playerFeaturedEvent           json.RawMessage
	playerFeaturedEventErr        error
	playerYearStats               json.RawMessage
	playerYearStatsErr            error
	sportLiveTournaments          json.RawMessage
	sportLiveTournamentsErr       error
	sportCategories               json.RawMessage
	sportCategoriesErr            error
	sportTopPlayers               json.RawMessage
	sportTopPlayersErr            error
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

func (f *fakeClient) PlayerSeasonStatistics(_ context.Context, _ int, _ int, _ int, phase string) (json.RawMessage, error) {
	if raw, ok := f.playerSeasonStatsByPhase[strings.TrimSpace(phase)]; ok {
		return raw, f.playerSeasonStatsErrByPhase[strings.TrimSpace(phase)]
	}
	if err, ok := f.playerSeasonStatsErrByPhase[strings.TrimSpace(phase)]; ok {
		return nil, err
	}
	return f.playerSeasonStats, f.playerSeasonStatsErr
}

func (f *fakeClient) PlayerSeasonRatings(_ context.Context, _ int, _ int, _ int, phase string) (json.RawMessage, error) {
	if raw, ok := f.playerSeasonRatingsByPhase[strings.TrimSpace(phase)]; ok {
		return raw, f.playerSeasonRatingsErrByPhase[strings.TrimSpace(phase)]
	}
	if err, ok := f.playerSeasonRatingsErrByPhase[strings.TrimSpace(phase)]; ok {
		return nil, err
	}
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

func (f *fakeClient) WatchEvents(context.Context, []int, []string, bool, func(sofascoreapi.WatchRecord) error) error {
	return nil
}

func (f *fakeClient) WatchSports(context.Context, []string, func(sofascoreapi.WatchRecord) error) error {
	return nil
}

func TestSearchReturnsIDsAndFilters(t *testing.T) {
	resp, err := Search(context.Background(), &fakeClient{
		searchResults: []sofascoreapi.SearchResult{
			{Type: "team", ID: 1, Name: "Fiorentina", Sport: "football"},
			{Type: "uniqueTournament", ID: 2, Name: "Premier League", Sport: "football"},
		},
	}, SearchParams{
		Query:      "premier",
		ResultType: "tournament",
		IDOnly:     true,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(resp.Results) != 1 || resp.Results[0].ID != 2 {
		t.Fatalf("unexpected filtered results: %+v", resp.Results)
	}
	if len(resp.IDs) != 1 || resp.IDs[0] != 2 {
		t.Fatalf("unexpected ids: %+v", resp.IDs)
	}
}

func TestSportEventsDefaultsDate(t *testing.T) {
	originalNow := now
	now = func() time.Time { return time.Date(2026, 3, 24, 11, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = originalNow })

	resp, err := SportEvents(context.Background(), &fakeClient{
		sportEvents: []sofascoreapi.EventSummary{
			{EventID: 1, StartTime: time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC), Home: "Torino", Away: "Parma"},
		},
	}, SportEventsParams{
		Sport: "football",
	})
	if err != nil {
		t.Fatalf("SportEvents returned error: %v", err)
	}
	if resp.Date != "2026-03-24" || resp.Sport != "football" || len(resp.Events) != 1 {
		t.Fatalf("unexpected sport events response: %+v", resp)
	}
}

func TestTrendingDetectsCountryFiltersByDateAndAppliesLimit(t *testing.T) {
	originalNow := now
	now = func() time.Time { return time.Date(2026, 3, 24, 11, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = originalNow })

	client := &fakeClient{
		detectedCountry: "dk",
		trendingEvents: []sofascoreapi.TrendingEventSummary{
			{Rank: 1, EventID: 11, StartTime: time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC), Sport: "football"},
			{Rank: 2, EventID: 12, StartTime: time.Date(2026, 3, 25, 9, 0, 0, 0, time.UTC), Sport: "football"},
			{Rank: 3, EventID: 13, StartTime: time.Date(2026, 3, 24, 12, 0, 0, 0, time.UTC), Sport: "tennis"},
		},
	}

	resp, err := Trending(context.Background(), client, TrendingParams{
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("Trending returned error: %v", err)
	}
	if client.detectCountryCalls != 1 {
		t.Fatalf("expected one country detection call, got %d", client.detectCountryCalls)
	}
	if client.trendingCountry != "DK" {
		t.Fatalf("unexpected trending country: %q", client.trendingCountry)
	}
	if client.trendingLimit != 0 {
		t.Fatalf("expected trending fetch without upstream limit, got %d", client.trendingLimit)
	}
	if resp.Country != "DK" || resp.Date != "2026-03-24" || len(resp.Events) != 1 || resp.Events[0].EventID != 11 {
		t.Fatalf("unexpected trending response: %+v", resp)
	}
}

func TestTrendingExplicitCountrySkipsDetection(t *testing.T) {
	client := &fakeClient{
		trendingEvents: []sofascoreapi.TrendingEventSummary{
			{Rank: 1, EventID: 11, StartTime: time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC)},
		},
	}

	_, err := Trending(context.Background(), client, TrendingParams{
		Country: "us",
		Date:    "2026-03-24",
	})
	if err != nil {
		t.Fatalf("Trending returned error: %v", err)
	}
	if client.detectCountryCalls != 0 {
		t.Fatalf("did not expect country detection, got %d calls", client.detectCountryCalls)
	}
	if client.trendingCountry != "US" {
		t.Fatalf("unexpected trending country: %q", client.trendingCountry)
	}
}

func TestSportScheduledTournamentsDefaultsDateAndPage(t *testing.T) {
	originalNow := now
	now = func() time.Time { return time.Date(2026, 3, 24, 11, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = originalNow })

	client := &fakeClient{
		sportScheduledTournaments: []sofascoreapi.ScheduledTournamentSummary{
			{UniqueTournamentID: 696, UniqueTournament: "UEFA Women's Champions League", Name: "Knockout stage"},
		},
		sportScheduledHasNext: true,
	}
	resp, err := SportScheduledTournaments(context.Background(), client, SportScheduledTournamentsParams{
		Sport: "football",
	})
	if err != nil {
		t.Fatalf("SportScheduledTournaments returned error: %v", err)
	}
	if client.sportScheduledSport != "football" || client.sportScheduledDate != "2026-03-24" || client.sportScheduledPage != 1 {
		t.Fatalf("unexpected scheduled tournament args: sport=%q date=%q page=%d", client.sportScheduledSport, client.sportScheduledDate, client.sportScheduledPage)
	}
	if !resp.HasNextPage || len(resp.Tournaments) != 1 || resp.Tournaments[0].UniqueTournamentID != 696 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestTournamentScheduledEventsDefaultsDate(t *testing.T) {
	originalNow := now
	now = func() time.Time { return time.Date(2026, 3, 24, 11, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = originalNow })

	client := &fakeClient{
		tournamentScheduledEvents: []sofascoreapi.EventSummary{
			{EventID: 1, StartTime: time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC), Home: "A", Away: "B"},
		},
	}
	resp, err := TournamentScheduledEvents(context.Background(), client, TournamentScheduledEventsParams{
		TournamentID: 696,
		Limit:        5,
	})
	if err != nil {
		t.Fatalf("TournamentScheduledEvents returned error: %v", err)
	}
	if client.tournamentScheduledID != 696 || client.tournamentScheduledDate != "2026-03-24" || client.tournamentScheduledLimit != 5 {
		t.Fatalf("unexpected args: id=%d date=%q limit=%d", client.tournamentScheduledID, client.tournamentScheduledDate, client.tournamentScheduledLimit)
	}
	if resp.Date != "2026-03-24" || len(resp.Events) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestEventAllowsPartialSectionErrors(t *testing.T) {
	resp, err := Event(context.Background(), &fakeClient{
		event: sofascoreapi.EventDetail{
			EventID:   9,
			StartTime: time.Unix(1700000000, 0).UTC(),
			Home:      "Torino",
			Away:      "Parma",
			Sport:     "football",
			Raw:       json.RawMessage(`{"event":{"id":9}}`),
		},
		availableSections: []string{"statistics", "lineups"},
		eventSections: map[string]json.RawMessage{
			"statistics": json.RawMessage(`{"shots":10}`),
		},
		sectionErrs: map[string]error{
			"lineups": &sofascoreapi.HTTPStatusError{StatusCode: 404, URL: "x"},
		},
	}, EventParams{
		EventID:                   9,
		Sections:                  []string{"statistics", "lineups"},
		AllowPartialSectionErrors: true,
	})
	if err != nil {
		t.Fatalf("Event returned error: %v", err)
	}

	if !resp.Partial {
		t.Fatal("expected partial response")
	}
	if resp.Sections["statistics"] == nil {
		t.Fatalf("expected statistics payload, got %+v", resp.Sections)
	}
	if !strings.Contains(resp.SectionErrors["lineups"], `section "lineups" not found for event 9`) {
		t.Fatalf("unexpected section error: %+v", resp.SectionErrors)
	}
}

func TestTournamentSectionErrorWithoutPartialFails(t *testing.T) {
	_, err := Tournament(context.Background(), &fakeClient{
		tournament: sofascoreapi.TournamentDetail{
			TournamentID: 17,
			Name:         "Premier League",
			Sport:        "football",
			Raw:          json.RawMessage(`{"tournament":{"id":17}}`),
		},
		tournamentSeasons:           []sofascoreapi.TournamentSeason{{ID: 99, Name: "2025/26"}},
		tournamentAvailableSections: []string{"standings/total"},
		tournamentSectionErrs: map[string]error{
			"standings/total": &sofascoreapi.HTTPStatusError{StatusCode: 404, URL: "x"},
		},
	}, TournamentParams{
		TournamentID: 17,
		Sections:     []string{"standings/total"},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var lookupErr *Error
	if !errors.As(err, &lookupErr) || lookupErr.Kind != ErrorKindNotFound {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamEventsRequiresExactlyOneDirection(t *testing.T) {
	_, err := TeamEvents(context.Background(), &fakeClient{}, TeamEventsParams{
		TeamID: 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "pass exactly one of --next or --last" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrendingRejectsInvalidCountry(t *testing.T) {
	_, err := Trending(context.Background(), &fakeClient{}, TrendingParams{
		Country: "D1",
		Date:    "2026-03-24",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "country must be a 2-letter ISO alpha-2 code" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSportEventsRejectsInvalidDate(t *testing.T) {
	_, err := SportEvents(context.Background(), &fakeClient{}, SportEventsParams{
		Sport: "football",
		Date:  "2026/03/24",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "date must use YYYY-MM-DD" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEventTVDecodesChannelsPayload(t *testing.T) {
	resp, err := EventTV(context.Background(), &fakeClient{
		eventTV: json.RawMessage(`{"countryChannels":{"DK":[4024],"US":[6688,4024]}}`),
	}, EventTVParams{EventID: 14442088})
	if err != nil {
		t.Fatalf("EventTV returned error: %v", err)
	}

	if resp.EventID != 14442088 {
		t.Fatalf("unexpected event id: %+v", resp)
	}
	channels, ok := resp.Channels.(map[string]any)
	if !ok || channels["countryChannels"] == nil {
		t.Fatalf("unexpected channels payload: %#v", resp.Channels)
	}
}

func TestEventTVChannelNormalizesNameAndChannel(t *testing.T) {
	resp, err := EventTVChannel(context.Background(), &fakeClient{
		eventTVChannel: json.RawMessage(`{"tvChannelVotes":{"tvChannel":{"id":263,"name":"Viaplay","link":"https://example.com"},"upvote":4,"downvote":0}}`),
	}, EventTVChannelParams{EventID: 15697200, ChannelID: 263})
	if err != nil {
		t.Fatalf("EventTVChannel returned error: %v", err)
	}

	if resp.EventID != 15697200 || resp.ChannelID != 263 || resp.ChannelName != "Viaplay" {
		t.Fatalf("unexpected normalized response: %+v", resp)
	}
	channel, ok := resp.Channel.(map[string]any)
	if !ok || channel["id"] == nil {
		t.Fatalf("unexpected channel payload: %#v", resp.Channel)
	}
	votes, ok := resp.Votes.(map[string]any)
	if !ok || votes["tvChannelVotes"] == nil {
		t.Fatalf("unexpected votes payload: %#v", resp.Votes)
	}
}

func TestEventH2HEventsUsesDedicatedWrapper(t *testing.T) {
	resp, err := EventH2HEvents(context.Background(), &fakeClient{
		eventH2H: sofascoreapi.EventH2HEventsResult{
			Slug: "utbsCtb",
			Raw:  json.RawMessage(`{"events":[{"id":14442088}]}`),
		},
	}, EventH2HEventsParams{EventID: 14442088})
	if err != nil {
		t.Fatalf("EventH2HEvents returned error: %v", err)
	}

	if resp.Slug != "utbsCtb" {
		t.Fatalf("unexpected slug: %+v", resp)
	}
	events, ok := resp.Events.(map[string]any)
	if !ok || events["events"] == nil {
		t.Fatalf("unexpected events payload: %#v", resp.Events)
	}
}

func TestTeamRankingsSupportsDirectAndTournamentModes(t *testing.T) {
	directResp, err := TeamRankings(context.Background(), &fakeClient{
		teamRankings: json.RawMessage(`{"rankings":[{"type":"ATP","ranking":1}]}`),
	}, TeamRankingsParams{TeamID: 206570})
	if err != nil {
		t.Fatalf("direct TeamRankings returned error: %v", err)
	}
	if directResp.Sport != "tennis" || directResp.TournamentID != 0 || directResp.SeasonID != 0 {
		t.Fatalf("unexpected direct response: %+v", directResp)
	}

	tournamentResp, err := TeamRankings(context.Background(), &fakeClient{
		teamTournamentRanks: json.RawMessage(`{"ranks":[{"name":"Points per game","value":1}]}`),
	}, TeamRankingsParams{TeamID: 3419, TournamentID: 132, SeasonID: 80229})
	if err != nil {
		t.Fatalf("tournament TeamRankings returned error: %v", err)
	}
	if tournamentResp.TournamentID != 132 || tournamentResp.SeasonID != 80229 {
		t.Fatalf("unexpected tournament response: %+v", tournamentResp)
	}
}

func TestPlayerAttributesDecodesPayload(t *testing.T) {
	resp, err := PlayerAttributes(context.Background(), &fakeClient{
		playerAttributes: json.RawMessage(`{"averageAttributeOverviews":[],"playerAttributeOverviews":[]}`),
	}, PlayerAttributesParams{PlayerID: 829022})
	if err != nil {
		t.Fatalf("PlayerAttributes returned error: %v", err)
	}
	if resp.PlayerID != 829022 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestPlayerMediaVideosFallsBackToTeamRoute(t *testing.T) {
	resp, err := PlayerMediaVideos(context.Background(), &fakeClient{
		playerMediaVideosErr: &sofascoreapi.HTTPStatusError{StatusCode: 404},
		teamMedia:            json.RawMessage(`{"videos":[{"id":1}]}`),
	}, PlayerMediaVideosParams{PlayerID: 206570})
	if err != nil {
		t.Fatalf("PlayerMediaVideos returned error: %v", err)
	}
	if resp.PlayerID != 206570 || resp.Videos == nil {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestPlayerSeasonsFallsBackToTennisRoute(t *testing.T) {
	resp, err := PlayerSeasons(context.Background(), &fakeClient{
		playerSeasonsErr:       &sofascoreapi.HTTPStatusError{StatusCode: 404},
		playerSeasonsTennis:    json.RawMessage(`{"seasons":[{"year":2026}]}`),
		playerSeasonsTennisErr: nil,
	}, PlayerSeasonsParams{PlayerID: 206570})
	if err != nil {
		t.Fatalf("PlayerSeasons returned error: %v", err)
	}
	if resp.PlayerID != 206570 || resp.Seasons == nil {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestPlayerSeasonStatsDefaultsToRegularSeasonAfterOverall404(t *testing.T) {
	resp, err := PlayerSeasonStats(context.Background(), &fakeClient{
		playerSeasonStatsByPhase: map[string]json.RawMessage{
			"regularSeason": json.RawMessage(`{"statistics":{"matches":10}}`),
		},
		playerSeasonStatsErrByPhase: map[string]error{
			"overall": &sofascoreapi.HTTPStatusError{StatusCode: 404},
		},
	}, PlayerSeasonStatsParams{
		PlayerID:     817181,
		TournamentID: 132,
		SeasonID:     80229,
	})
	if err != nil {
		t.Fatalf("PlayerSeasonStats returned error: %v", err)
	}
	if resp.Phase != "regularSeason" || resp.TournamentID != 132 || resp.SeasonID != 80229 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
