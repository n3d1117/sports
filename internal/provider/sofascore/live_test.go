package sofascoreapi

import (
	"reflect"
	"testing"
)

func TestParseLivePatchSupportsIDAndEventID(t *testing.T) {
	t.Run("id", func(t *testing.T) {
		eventID, flat, expanded, err := parseLivePatch([]byte(`{"id":15636234,"status.description":"HT","homeScore.current":1}`))
		if err != nil {
			t.Fatalf("parseLivePatch returned error: %v", err)
		}
		if eventID != 15636234 {
			t.Fatalf("unexpected event id %d", eventID)
		}
		if _, ok := flat["id"]; ok {
			t.Fatalf("flat patch should not include id: %+v", flat)
		}
		if got := getString(expanded, "status", "description"); got != "HT" {
			t.Fatalf("unexpected expanded status: %q", got)
		}
		if got := getInt(expanded, "homeScore", "current"); got != 1 {
			t.Fatalf("unexpected expanded home score: %d", got)
		}
	})

	t.Run("eventId", func(t *testing.T) {
		eventID, _, _, err := parseLivePatch([]byte(`{"eventId":"15855340","status.description":"Set 2"}`))
		if err != nil {
			t.Fatalf("parseLivePatch returned error: %v", err)
		}
		if eventID != 15855340 {
			t.Fatalf("unexpected event id %d", eventID)
		}
	})
}

func TestDeepMergeObjectsWithExpandedPatch(t *testing.T) {
	base := map[string]any{
		"id": 15636234,
		"status": map[string]any{
			"type":        "inprogress",
			"description": "1st half",
		},
		"homeScore": map[string]any{
			"current": 0,
		},
	}
	patch := expandDottedKeys(map[string]any{
		"status.description": "HT",
		"homeScore.current":  1,
	})

	merged := deepMergeObjects(base, patch)
	if got := getString(merged, "status", "description"); got != "HT" {
		t.Fatalf("unexpected merged status: %q", got)
	}
	if got := getInt(merged, "homeScore", "current"); got != 1 {
		t.Fatalf("unexpected merged score: %d", got)
	}
	if got := getString(merged, "status", "type"); got != "inprogress" {
		t.Fatalf("deep merge dropped existing nested field: %q", got)
	}
}

func TestSummaryFromEventMapForDifferentSports(t *testing.T) {
	tests := []struct {
		name    string
		event   map[string]any
		want    WatchEventSummary
		homePtr *int
		awayPtr *int
	}{
		{
			name: "football",
			event: map[string]any{
				"id":             15636234,
				"startTimestamp": 1711274400,
				"status": map[string]any{
					"type":        "inprogress",
					"description": "2nd half",
				},
				"homeTeam": map[string]any{"name": "Torino"},
				"awayTeam": map[string]any{"name": "Parma"},
				"tournament": map[string]any{
					"name": "Serie A",
					"category": map[string]any{
						"sport": map[string]any{"slug": "football"},
					},
				},
				"homeScore": map[string]any{"current": 2},
				"awayScore": map[string]any{"current": 1},
			},
			want: WatchEventSummary{
				EventID:           15636234,
				StatusType:        "inprogress",
				StatusDescription: "2nd half",
				Home:              "Torino",
				Away:              "Parma",
				Tournament:        "Serie A",
				Sport:             "football",
			},
			homePtr: intPtr(2),
			awayPtr: intPtr(1),
		},
		{
			name: "tennis",
			event: map[string]any{
				"id": 15855340,
				"status": map[string]any{
					"type":        "inprogress",
					"description": "Set 2",
				},
				"homePlayer": map[string]any{"name": "Player A"},
				"awayPlayer": map[string]any{"name": "Player B"},
				"tournament": map[string]any{
					"name": "Indian Wells",
					"category": map[string]any{
						"sport": map[string]any{"slug": "tennis"},
					},
				},
			},
			want: WatchEventSummary{
				EventID:           15855340,
				StatusType:        "inprogress",
				StatusDescription: "Set 2",
				Home:              "Player A",
				Away:              "Player B",
				Tournament:        "Indian Wells",
				Sport:             "tennis",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summaryFromEventMap(tt.event)
			if got.EventID != tt.want.EventID || got.StatusType != tt.want.StatusType || got.StatusDescription != tt.want.StatusDescription || got.Home != tt.want.Home || got.Away != tt.want.Away || got.Tournament != tt.want.Tournament || got.Sport != tt.want.Sport {
				t.Fatalf("unexpected summary: got %+v want %+v", got, tt.want)
			}
			if !reflect.DeepEqual(got.HomeScore, tt.homePtr) || !reflect.DeepEqual(got.AwayScore, tt.awayPtr) {
				t.Fatalf("unexpected scores: got home=%v away=%v", got.HomeScore, got.AwayScore)
			}
		})
	}
}

func intPtr(value int) *int {
	return &value
}
