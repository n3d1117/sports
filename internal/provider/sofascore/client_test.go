package sofascoreapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSearchNormalizesResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "SportsApp/1.0" {
			t.Fatalf("unexpected user agent %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"entity":{"id":2693,"name":"Fiorentina","slug":"fiorentina","sport":{"slug":"football"},"country":{"name":"Italy"},"team":{}},"score":1848375.8,"type":"team"}]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	results, err := client.Search(context.Background(), "fiorentina", 0)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != 2693 || results[0].Name != "Fiorentina" || results[0].Sport != "football" {
		t.Fatalf("unexpected result %+v", results[0])
	}
}

func TestSearchUsesPageParameter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("unexpected page %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if _, err := client.Search(context.Background(), "fiorentina", 2); err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
}

func TestSearchNormalizesTournamentCategoryFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"entity":{"id":17,"name":"Premier League","slug":"premier-league","category":{"name":"England","sport":{"slug":"football"},"country":{"name":"England"}}},"score":123.4,"type":"uniqueTournament"}]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	results, err := client.Search(context.Background(), "premier league", 0)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "uniqueTournament" || results[0].Sport != "football" || results[0].Category != "England" || results[0].Country != "England" {
		t.Fatalf("unexpected result %+v", results[0])
	}
}

func TestTeamEventsNormalizesResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"events":[{"id":13981704,"startTimestamp":1773690300,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Cremonese"},"awayTeam":{"name":"Fiorentina"},"tournament":{"name":"Serie A","category":{"sport":{"slug":"football"}}}}]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.TeamEvents(context.Background(), 2693, "next", 10)
	if err != nil {
		t.Fatalf("TeamEvents returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventID != 13981704 || events[0].Home != "Cremonese" || events[0].Away != "Fiorentina" {
		t.Fatalf("unexpected event %+v", events[0])
	}
}

func TestTeamEventsLastReturnsMostRecentFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/team/2693/events/last/0":
			w.Write([]byte(`{"events":[
				{"id":1,"startTimestamp":1735689600,"status":{"type":"finished","description":"Ended"},"homeTeam":{"name":"Older"},"awayTeam":{"name":"A"},"tournament":{"name":"Cup","category":{"sport":{"slug":"football"}}}},
				{"id":2,"startTimestamp":1735776000,"status":{"type":"finished","description":"Ended"},"homeTeam":{"name":"Newer"},"awayTeam":{"name":"B"},"tournament":{"name":"Cup","category":{"sport":{"slug":"football"}}}},
				{"id":4,"startTimestamp":1735862400,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Future"},"awayTeam":{"name":"D"},"tournament":{"name":"Cup","category":{"sport":{"slug":"football"}}}}
			],"hasNextPage":true}`))
		case "/team/2693/events/last/1":
			w.Write([]byte(`{"events":[
				{"id":3,"startTimestamp":1735603200,"status":{"type":"finished","description":"Ended"},"homeTeam":{"name":"Oldest"},"awayTeam":{"name":"C"},"tournament":{"name":"Cup","category":{"sport":{"slug":"football"}}}}
			],"hasNextPage":false}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.TeamEvents(context.Background(), 2693, "last", 3)
	if err != nil {
		t.Fatalf("TeamEvents returned error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].EventID != 2 || events[1].EventID != 1 || events[2].EventID != 3 {
		t.Fatalf("unexpected order %+v", events)
	}
}

func TestTeamEventsNextFallsBackToLastNotStartedEventsOnNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/team/206570/events/next/0":
			http.NotFound(w, r)
		case "/team/206570":
			w.Write([]byte(`{"team":{"id":206570,"name":"Jannik Sinner"}}`))
		case "/team/206570/events/last/0":
			w.Write([]byte(`{"events":[
				{"id":10,"startTimestamp":1736035200,"status":{"type":"finished","description":"Ended"},"homeTeam":{"name":"Past"},"awayTeam":{"name":"A"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}},
				{"id":12,"startTimestamp":1736208000,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Soon"},"awayTeam":{"name":"B"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}},
				{"id":11,"startTimestamp":1736294400,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Later"},"awayTeam":{"name":"C"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}}
			],"hasNextPage":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.TeamEvents(context.Background(), 206570, "next", 10)
	if err != nil {
		t.Fatalf("TeamEvents returned error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %+v", events)
	}
	if events[0].EventID != 12 || events[1].EventID != 11 {
		t.Fatalf("unexpected order %+v", events)
	}
}

func TestTeamEventsNextFallsBackToLastNotStartedEventsOnEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/team/206570/events/next/0":
			w.Write([]byte(`{"events":[],"hasNextPage":false}`))
		case "/team/206570/events/last/0":
			w.Write([]byte(`{"events":[
				{"id":21,"startTimestamp":1736208000,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Soon"},"awayTeam":{"name":"B"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}},
				{"id":20,"startTimestamp":1736121600,"status":{"type":"finished","description":"Ended"},"homeTeam":{"name":"Past"},"awayTeam":{"name":"A"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}}
			],"hasNextPage":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.TeamEvents(context.Background(), 206570, "next", 10)
	if err != nil {
		t.Fatalf("TeamEvents returned error: %v", err)
	}
	if len(events) != 1 || events[0].EventID != 21 {
		t.Fatalf("unexpected events %+v", events)
	}
}

func TestTeamEventsNextFallbackRespectsLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/team/206570/events/next/0":
			http.NotFound(w, r)
		case "/team/206570":
			w.Write([]byte(`{"team":{"id":206570,"name":"Jannik Sinner"}}`))
		case "/team/206570/events/last/0":
			w.Write([]byte(`{"events":[
				{"id":32,"startTimestamp":1736380800,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Latest"},"awayTeam":{"name":"C"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}},
				{"id":31,"startTimestamp":1736208000,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Earliest"},"awayTeam":{"name":"B"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}},
				{"id":30,"startTimestamp":1736121600,"status":{"type":"finished","description":"Ended"},"homeTeam":{"name":"Past"},"awayTeam":{"name":"A"},"tournament":{"name":"Cup","category":{"sport":{"slug":"tennis"}}}}
			],"hasNextPage":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.TeamEvents(context.Background(), 206570, "next", 1)
	if err != nil {
		t.Fatalf("TeamEvents returned error: %v", err)
	}
	if len(events) != 1 || events[0].EventID != 31 {
		t.Fatalf("unexpected events %+v", events)
	}
}

func TestTeamEventsNextKeepsNotFoundForInvalidTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.TeamEvents(context.Background(), 999999999, "next", 10)
	if err == nil {
		t.Fatal("expected error")
	}

	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected HTTP 404, got %v", err)
	}
}

func TestEventNormalizesSummaryAndRawPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"event":{"id":13981714,"startTimestamp":1773431100,"status":{"type":"inprogress","description":"1st half"},"homeTeam":{"name":"Torino"},"awayTeam":{"name":"Parma"},"homeScore":{"current":1},"awayScore":{"current":0},"tournament":{"name":"Serie A","category":{"sport":{"slug":"football"}}},"venue":{"name":"Olimpico Grande Torino"}}}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	event, err := client.Event(context.Background(), 13981714)
	if err != nil {
		t.Fatalf("Event returned error: %v", err)
	}
	if event.Home != "Torino" || event.Away != "Parma" || event.Sport != "football" {
		t.Fatalf("unexpected event %+v", event)
	}
	if len(event.Raw) == 0 {
		t.Fatal("expected raw payload")
	}
}

func TestSportsNormalizesEventCounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/sport/3600/event-count" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tennis":{"live":4,"total":80},"football":{"live":12,"total":300}}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	sports, err := client.Sports(context.Background())
	if err != nil {
		t.Fatalf("Sports returned error: %v", err)
	}
	if len(sports) != 2 || sports[0].Slug != "football" || sports[1].Slug != "tennis" {
		t.Fatalf("unexpected sports %+v", sports)
	}
}

func TestSportEventsNormalizesScheduledEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/sport/football/scheduled-events/2026-03-24" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"events":[{"id":13981704,"startTimestamp":1773690300,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Cremonese"},"awayTeam":{"name":"Fiorentina"},"tournament":{"name":"Serie A","category":{"sport":{"slug":"football"}}}}]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.SportEvents(context.Background(), "football", "2026-03-24", 10)
	if err != nil {
		t.Fatalf("SportEvents returned error: %v", err)
	}
	if len(events) != 1 || events[0].EventID != 13981704 || events[0].Home != "Cremonese" {
		t.Fatalf("unexpected events %+v", events)
	}
}

func TestSportScheduledTournamentsNormalizesGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/sport/football/scheduled-tournaments/2026-03-24/page/2" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"scheduled":[{"tournament":{"id":11392,"name":"UEFA Champions League, Women, Knockout stage","slug":"uefa-champions-league-women-knockout-stage","category":{"name":"Europe","sport":{"slug":"football"}},"uniqueTournament":{"id":696,"name":"UEFA Women's Champions League","slug":"uefa-womens-champions-league","category":{"name":"Europe","sport":{"slug":"football"}}}},"timezoneEventCount":{"0":2,"3600":2}}],"hasNextPage":true}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	tournaments, hasNextPage, err := client.SportScheduledTournaments(context.Background(), "football", "2026-03-24", 2)
	if err != nil {
		t.Fatalf("SportScheduledTournaments returned error: %v", err)
	}
	if !hasNextPage || len(tournaments) != 1 || tournaments[0].UniqueTournamentID != 696 || tournaments[0].UTCEventCount != 2 {
		t.Fatalf("unexpected tournaments %+v hasNext=%t", tournaments, hasNextPage)
	}
}

func TestDetectCountryAlpha2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/country/alpha2" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alpha2":"dk"}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	alpha2, err := client.DetectCountryAlpha2(context.Background())
	if err != nil {
		t.Fatalf("DetectCountryAlpha2 returned error: %v", err)
	}
	if alpha2 != "DK" {
		t.Fatalf("unexpected alpha2 %q", alpha2)
	}
}

func TestTournamentScheduledEventsNormalizesScheduledEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/unique-tournament/696/scheduled-events/2026-03-24" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"events":[{"id":15471604,"startTimestamp":1774374300,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"VfL Wolfsburg"},"awayTeam":{"name":"OL Lyonnes"},"tournament":{"name":"UEFA Champions League, Women, Knockout stage","category":{"sport":{"slug":"football"}}}}]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.TournamentScheduledEvents(context.Background(), 696, "2026-03-24", 10)
	if err != nil {
		t.Fatalf("TournamentScheduledEvents returned error: %v", err)
	}
	if len(events) != 1 || events[0].EventID != 15471604 || events[0].Home != "VfL Wolfsburg" {
		t.Fatalf("unexpected events %+v", events)
	}
}

func TestTrendingEventsPreservesFeedOrderAndRank(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/trending/events/DK/all" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"events":[
			{"id":11,"startTimestamp":1773766800,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"A"},"awayTeam":{"name":"B"},"tournament":{"name":"Serie A","category":{"sport":{"slug":"football"}}}},
			{"id":12,"startTimestamp":1773770400,"status":{"type":"notstarted","description":"Not started"},"homePlayer":{"name":"Player A"},"awayPlayer":{"name":"Player B"},"tournament":{"name":"Indian Wells","category":{"sport":{"slug":"tennis"}}}}
		]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	events, err := client.TrendingEvents(context.Background(), "DK", 0)
	if err != nil {
		t.Fatalf("TrendingEvents returned error: %v", err)
	}
	if len(events) != 2 || events[0].Rank != 1 || events[1].Rank != 2 {
		t.Fatalf("unexpected trending events %+v", events)
	}
	if events[0].EventID != 11 || events[1].EventID != 12 || events[1].Sport != "tennis" {
		t.Fatalf("unexpected trending events %+v", events)
	}
}

func TestSportSectionsDiscoversWorkingSections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/sport/football/scheduled-events/2026-03-13":
			w.Write([]byte(`{"events":[{"id":13981714}]}`))
		case "/event/13981714/statistics":
			w.Write([]byte(`{"groups":[]}`))
		case "/event/13981714/h2h":
			w.Write([]byte(`{"events":[]}`))
		case "/event/13981714/team-streaks":
			w.Write([]byte(`{"streaks":[]}`))
		case "/event/13981714/votes":
			w.Write([]byte(`{"home":0,"away":0}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	client.now = func() time.Time {
		return time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)
	}

	discovery, err := client.SportSections(context.Background(), "football")
	if err != nil {
		t.Fatalf("SportSections returned error: %v", err)
	}
	if discovery.SampleEventID != 13981714 {
		t.Fatalf("unexpected sample event id %d", discovery.SampleEventID)
	}
	if got := strings.Join(discovery.Sections, ","); got != "statistics,h2h,team-streaks,votes" {
		t.Fatalf("unexpected sections %q", got)
	}
}

func TestProbeEventSectionsDiscoversSupportedSections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/event/13981714/statistics":
			w.Write([]byte(`{"groups":[]}`))
		case "/event/13981714/h2h":
			w.Write([]byte(`{"events":[]}`))
		case "/event/13981714/team-streaks":
			w.Write([]byte(`{"streaks":[]}`))
		case "/event/13981714/votes":
			w.Write([]byte(`{"home":0,"away":0}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	sections, err := client.ProbeEventSections(context.Background(), 13981714)
	if err != nil {
		t.Fatalf("ProbeEventSections returned error: %v", err)
	}
	if got := strings.Join(sections, ","); got != "statistics,h2h,team-streaks,votes" {
		t.Fatalf("unexpected sections %q", got)
	}
}

func TestProbeEventSectionsFallsBackToGetWhenHeadIsNotAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/event/13981714/statistics":
			if r.Method == http.MethodHead {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Write([]byte(`{"groups":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	sections, err := client.ProbeEventSections(context.Background(), 13981714)
	if err != nil {
		t.Fatalf("ProbeEventSections returned error: %v", err)
	}
	if got := strings.Join(sections, ","); got != "statistics" {
		t.Fatalf("unexpected sections %q", got)
	}
}

func TestProbeEventSectionsSkipsForbiddenAndServerErrorSections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/event/13981714/statistics":
			w.Write([]byte(`{"groups":[]}`))
		case "/event/13981714/official-tweets":
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		case "/event/13981714/umpires":
			http.Error(w, `{"error":"backend failure"}`, http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	sections, err := client.ProbeEventSections(context.Background(), 13981714)
	if err != nil {
		t.Fatalf("ProbeEventSections returned error: %v", err)
	}
	if got := strings.Join(sections, ","); got != "statistics" {
		t.Fatalf("unexpected sections %q", got)
	}
}

func TestSportSectionsSearchesWiderDateWindow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/sport/football/scheduled-events/2026-03-13",
			"/sport/football/scheduled-events/2026-03-12",
			"/sport/football/scheduled-events/2026-03-14",
			"/sport/football/scheduled-events/2026-03-11",
			"/sport/football/scheduled-events/2026-03-15",
			"/sport/football/scheduled-events/2026-03-10",
			"/sport/football/scheduled-events/2026-03-16",
			"/sport/football/scheduled-events/2026-03-09",
			"/sport/football/scheduled-events/2026-03-17",
			"/sport/football/scheduled-events/2026-03-08",
			"/sport/football/scheduled-events/2026-03-18",
			"/sport/football/scheduled-events/2026-03-07",
			"/sport/football/scheduled-events/2026-03-19",
			"/sport/football/scheduled-events/2026-03-06",
			"/sport/football/scheduled-events/2026-03-20",
			"/sport/football/scheduled-events/2026-03-05",
			"/sport/football/scheduled-events/2026-03-21",
			"/sport/football/scheduled-events/2026-03-04",
			"/sport/football/scheduled-events/2026-03-22",
			"/sport/football/scheduled-events/2026-03-03":
			w.Write([]byte(`{"events":[]}`))
		case "/sport/football/scheduled-events/2026-03-23":
			w.Write([]byte(`{"events":[{"id":13981714}]}`))
		case "/event/13981714/statistics":
			w.Write([]byte(`{"groups":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	client.now = func() time.Time {
		return time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)
	}

	discovery, err := client.SportSections(context.Background(), "football")
	if err != nil {
		t.Fatalf("SportSections returned error: %v", err)
	}
	if discovery.SampleEventID != 13981714 {
		t.Fatalf("unexpected sample event id %d", discovery.SampleEventID)
	}
	if got := strings.Join(discovery.Sections, ","); got != "statistics" {
		t.Fatalf("unexpected sections %q", got)
	}
}

func TestEventSectionPreservesRawJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/event/13981714/best-players/summary" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"bestPlayers":[]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	payload, err := client.EventSection(context.Background(), 13981714, "best-players/summary")
	if err != nil {
		t.Fatalf("EventSection returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
}

func TestTournamentNormalizesDetailAndSeasons(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/unique-tournament/17":
			w.Write([]byte(`{"uniqueTournament":{"id":17,"name":"Premier League","slug":"premier-league","category":{"name":"England","sport":{"slug":"football"},"country":{"name":"England"}}}}`))
		case "/unique-tournament/17/seasons":
			w.Write([]byte(`{"seasons":[{"id":76986,"name":"Premier League 25/26","year":"25/26"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())

	tournament, err := client.Tournament(context.Background(), 17)
	if err != nil {
		t.Fatalf("Tournament returned error: %v", err)
	}
	if tournament.Name != "Premier League" || tournament.Sport != "football" || tournament.Category != "England" {
		t.Fatalf("unexpected tournament %+v", tournament)
	}

	seasons, err := client.TournamentSeasons(context.Background(), 17)
	if err != nil {
		t.Fatalf("TournamentSeasons returned error: %v", err)
	}
	if len(seasons) != 1 || seasons[0].ID != 76986 || seasons[0].Name != "Premier League 25/26" {
		t.Fatalf("unexpected seasons %+v", seasons)
	}
}

func TestProbeTournamentSectionsDiscoversMixedScopeSections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/unique-tournament/17/media":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Write([]byte(`{"media":[]}`))
		case "/unique-tournament/17/featured-events":
			w.Write([]byte(`{"events":[]}`))
		case "/unique-tournament/17/season/76986/standings/total":
			if r.Method == http.MethodHead {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Write([]byte(`{"standings":[]}`))
		case "/unique-tournament/17/season/76986/rounds":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Write([]byte(`{"rounds":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	sections, err := client.ProbeTournamentSections(context.Background(), 17, 76986)
	if err != nil {
		t.Fatalf("ProbeTournamentSections returned error: %v", err)
	}
	if got := strings.Join(sections, ","); got != "featured-events,media,rounds,standings/total" {
		t.Fatalf("unexpected sections %q", got)
	}
}

func TestTournamentSectionUsesScopeAwarePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/unique-tournament/17/media":
			w.Write([]byte(`{"media":[]}`))
		case "/unique-tournament/17/season/76986/team-of-the-week/periods":
			w.Write([]byte(`{"periods":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if _, err := client.TournamentSection(context.Background(), 17, 76986, "media"); err != nil {
		t.Fatalf("TournamentSection media returned error: %v", err)
	}
	if _, err := client.TournamentSection(context.Background(), 17, 76986, "team-of-the-week/periods"); err != nil {
		t.Fatalf("TournamentSection team-of-the-week returned error: %v", err)
	}
}

func TestTournamentEventsNormalizesNextRoundAndUnsupportedLast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/unique-tournament/17/season/76986/events/next/0":
			w.Write([]byte(`{"events":[{"id":1001,"startTimestamp":1773600000,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Arsenal"},"awayTeam":{"name":"Chelsea"},"tournament":{"name":"Premier League","category":{"sport":{"slug":"football"}}}}],"hasNextPage":false}`))
		case "/unique-tournament/17/season/76986/events/round/5/slug/round-of-16":
			w.Write([]byte(`{"events":[{"id":1002,"startTimestamp":1773686400,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Team A"},"awayTeam":{"name":"Team B"},"tournament":{"name":"Cup","category":{"sport":{"slug":"football"}}}}]}`))
		case "/unique-tournament/17/season/76986/events/last/0":
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())

	nextEvents, err := client.TournamentEvents(context.Background(), 17, 76986, "next", 0, "", 10)
	if err != nil {
		t.Fatalf("TournamentEvents next returned error: %v", err)
	}
	if len(nextEvents) != 1 || nextEvents[0].EventID != 1001 || nextEvents[0].Sport != "football" {
		t.Fatalf("unexpected next events %+v", nextEvents)
	}

	roundEvents, err := client.TournamentEvents(context.Background(), 17, 76986, "round", 5, "round-of-16", 10)
	if err != nil {
		t.Fatalf("TournamentEvents round returned error: %v", err)
	}
	if len(roundEvents) != 1 || roundEvents[0].EventID != 1002 {
		t.Fatalf("unexpected round events %+v", roundEvents)
	}

	_, err = client.TournamentEvents(context.Background(), 17, 76986, "last", 0, "", 10)
	if err == nil {
		t.Fatal("expected unsupported last error")
	}

	var unsupported *UnsupportedTournamentEventsError
	if !errors.As(err, &unsupported) {
		t.Fatalf("expected UnsupportedTournamentEventsError, got %v", err)
	}
}

func TestEventTVChannelsUsesTVPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/tv/event/14442088/country-channels" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"countryChannels":{"DK":[4024],"US":[6688,4024]}}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	raw, err := client.EventTVChannels(context.Background(), 14442088)
	if err != nil {
		t.Fatalf("EventTVChannels returned error: %v", err)
	}
	if !strings.Contains(string(raw), `"countryChannels"`) {
		t.Fatalf("unexpected payload %s", string(raw))
	}
}

func TestEventTVChannelVotesUsesVotesPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/tv/channel/263/event/14442088/votes" {
			t.Fatalf("unexpected path %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tvChannelVotes":{"tvChannel":{"id":263,"name":"Viaplay"},"upvote":4,"downvote":0}}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	raw, err := client.EventTVChannelVotes(context.Background(), 263, 14442088)
	if err != nil {
		t.Fatalf("EventTVChannelVotes returned error: %v", err)
	}
	if !strings.Contains(string(raw), `"Viaplay"`) {
		t.Fatalf("unexpected payload %s", string(raw))
	}
}

func TestEventH2HEventsUsesCustomID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/event/14442088":
			w.Write([]byte(`{"event":{"id":14442088,"customId":"utbsCtb","startTimestamp":1773690300,"status":{"type":"notstarted"},"homeTeam":{"name":"Indiana Pacers"},"awayTeam":{"name":"Los Angeles Lakers"},"tournament":{"name":"NBA","category":{"sport":{"slug":"basketball"}}},"venue":{"name":"Gainbridge Fieldhouse"}}}`))
		case "/event/utbsCtb/h2h/events":
			w.Write([]byte(`{"events":[{"id":14442088}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	result, err := client.EventH2HEvents(context.Background(), 14442088)
	if err != nil {
		t.Fatalf("EventH2HEvents returned error: %v", err)
	}
	if result.Slug != "utbsCtb" || !strings.Contains(string(result.Raw), `"events"`) {
		t.Fatalf("unexpected result %+v", result)
	}
}

func TestTeamStandingsUsesResolvedUniqueTournamentSeason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/team/3419/standings/seasons":
			w.Write([]byte(`{"uniqueTournamentSeasons":[{"uniqueTournament":{"id":132,"name":"NBA"},"seasons":[{"id":80229,"name":"NBA 25/26","year":"2025/26"}]}]}`))
		case "/unique-tournament/132/season/80229/standings/total":
			w.Write([]byte(`{"standings":[{"rows":[]}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	result, err := client.TeamStandings(context.Background(), 3419, 0)
	if err != nil {
		t.Fatalf("TeamStandings returned error: %v", err)
	}
	if result.Season.UniqueTournamentID != 132 || result.Season.SeasonID != 80229 {
		t.Fatalf("unexpected season %+v", result.Season)
	}
	if !strings.Contains(string(result.Raw), `"standings"`) {
		t.Fatalf("unexpected payload %s", string(result.Raw))
	}
}

func TestPlayerAttributeOverviewsAndSportCategoriesPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/player/829022/attribute-overviews":
			w.Write([]byte(`{"averageAttributeOverviews":[],"playerAttributeOverviews":[]}`))
		case "/sport/football/categories/all":
			w.Write([]byte(`{"categories":[{"slug":"england"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	attributes, err := client.PlayerAttributeOverviews(context.Background(), 829022)
	if err != nil {
		t.Fatalf("PlayerAttributeOverviews returned error: %v", err)
	}
	if !strings.Contains(string(attributes), `"playerAttributeOverviews"`) {
		t.Fatalf("unexpected attributes payload %s", string(attributes))
	}

	categories, err := client.SportCategories(context.Background(), "football")
	if err != nil {
		t.Fatalf("SportCategories returned error: %v", err)
	}
	if !strings.Contains(string(categories), `"categories"`) {
		t.Fatalf("unexpected categories payload %s", string(categories))
	}
}

func TestPlayerRoutePaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/player/851284/media":
			w.Write([]byte(`{"media":[]}`))
		case "/player/851284/media/videos":
			w.Write([]byte(`{"videos":[]}`))
		case "/player/851284/statistics/seasons":
			w.Write([]byte(`{"seasons":[]}`))
		case "/player/851284/unique-tournament/23/season/76457/statistics/overall":
			w.Write([]byte(`{"statistics":{"minutesPlayed":1}}`))
		case "/player/851284/unique-tournament/23/season/76457/ratings/overall":
			w.Write([]byte(`{"ratings":[7.1]}`))
		case "/player/851284/unique-tournament/23/season/76457/heatmap/overall":
			w.Write([]byte(`{"heatmap":[]}`))
		case "/player/851284/penalty-history/unique-tournament/23/season/76457":
			w.Write([]byte(`{"penalties":[]}`))
		case "/player/851284/statistics":
			w.Write([]byte(`{"statistics":[]}`))
		case "/player/851284/statistics/match-type/overall":
			w.Write([]byte(`{"statistics":[]}`))
		case "/player/817181/unique-tournament/132/season/80229/shot-actions/regularSeason":
			w.Write([]byte(`{"shotmap":[]}`))
		case "/unique-tournament/132/season/80229/shot-action-areas/player/regularSeason":
			w.Write([]byte(`{"areas":[]}`))
		case "/player/851284/events/last/0":
			w.Write([]byte(`{"events":[{"id":1,"startTimestamp":1735776000,"status":{"type":"finished","description":"Ended"},"homeTeam":{"name":"A"},"awayTeam":{"name":"B"},"tournament":{"name":"Serie A","category":{"sport":{"slug":"football"}}}}],"hasNextPage":false}`))
		case "/team/206570/featured-event":
			w.Write([]byte(`{"featuredEvent":{"id":99}}`))
		case "/team/206570/team-statistics/seasons":
			w.Write([]byte(`{"seasons":[{"year":2026}]}`))
		case "/team/206570/year-statistics/2026":
			w.Write([]byte(`{"statistics":{"wins":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	calls := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) { return client.PlayerMedia(context.Background(), 851284) },
		func() (json.RawMessage, error) { return client.PlayerMediaVideos(context.Background(), 851284) },
		func() (json.RawMessage, error) { return client.PlayerStatisticsSeasons(context.Background(), 851284) },
		func() (json.RawMessage, error) {
			return client.PlayerSeasonStatistics(context.Background(), 851284, 23, 76457, "overall")
		},
		func() (json.RawMessage, error) {
			return client.PlayerSeasonRatings(context.Background(), 851284, 23, 76457, "overall")
		},
		func() (json.RawMessage, error) {
			return client.PlayerSeasonHeatmap(context.Background(), 851284, 23, 76457, "overall")
		},
		func() (json.RawMessage, error) {
			return client.PlayerPenaltyHistory(context.Background(), 851284, 23, 76457)
		},
		func() (json.RawMessage, error) { return client.PlayerCareerStatistics(context.Background(), 851284) },
		func() (json.RawMessage, error) {
			return client.PlayerCareerStatisticsMatchType(context.Background(), 851284, "overall")
		},
		func() (json.RawMessage, error) {
			return client.PlayerShotActions(context.Background(), 817181, 132, 80229, "regularSeason")
		},
		func() (json.RawMessage, error) {
			return client.PlayerShotActionAreas(context.Background(), 132, 80229, "regularSeason")
		},
		func() (json.RawMessage, error) { return client.PlayerFeaturedEvent(context.Background(), 206570) },
		func() (json.RawMessage, error) {
			return client.PlayerStatisticsSeasonsTennis(context.Background(), 206570)
		},
		func() (json.RawMessage, error) {
			return client.PlayerYearStatistics(context.Background(), 206570, 2026)
		},
	}

	for i, call := range calls {
		raw, err := call()
		if err != nil {
			t.Fatalf("call %d returned error: %v", i, err)
		}
		if len(raw) == 0 {
			t.Fatalf("call %d returned empty payload", i)
		}
	}

	events, err := client.PlayerLastEvents(context.Background(), 851284, 10)
	if err != nil {
		t.Fatalf("PlayerLastEvents returned error: %v", err)
	}
	if len(events) != 1 || events[0].EventID != 1 {
		t.Fatalf("unexpected player events %+v", events)
	}
}

func TestPlayerMediaVideosReturnsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.PlayerMediaVideos(context.Background(), 851284)
	if err == nil {
		t.Fatal("expected error")
	}

	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected HTTP 404, got %v", err)
	}
}
