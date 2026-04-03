package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sports/internal/lookups"
	"sports/internal/provider/sofascore"
)

type liveResourceConfig struct {
	URI         string
	WatchKind   sofascoreapi.WatchKind
	EventIDs    []int
	Sports      []string
	Sections    []string
	AllSections bool
}

type liveEventSnapshot struct {
	Summary       *sofascoreapi.WatchEventSummary `json:"summary,omitempty"`
	Event         map[string]any                  `json:"event,omitempty"`
	Sections      map[string]any                  `json:"sections,omitempty"`
	SectionErrors map[string]string               `json:"section_errors,omitempty"`
	LastPatch     map[string]any                  `json:"last_patch,omitempty"`
	ChangedFields []string                        `json:"changed_fields,omitempty"`
	LastRecord    sofascoreapi.WatchRecordType    `json:"last_record,omitempty"`
	UpdatedAt     string                          `json:"updated_at,omitempty"`
}

type liveSportSnapshot struct {
	Sport         string                                    `json:"sport"`
	Events        map[string]sofascoreapi.WatchEventSummary `json:"events,omitempty"`
	LastEventID   int                                       `json:"last_event_id,omitempty"`
	LastPatch     map[string]any                            `json:"last_patch,omitempty"`
	ChangedFields []string                                  `json:"changed_fields,omitempty"`
	UpdatedAt     string                                    `json:"updated_at,omitempty"`
}

type liveResourceState struct {
	URI             string                       `json:"uri"`
	WatchKind       sofascoreapi.WatchKind       `json:"watch_kind"`
	EventIDs        []int                        `json:"event_ids,omitempty"`
	Sports          []string                     `json:"sports,omitempty"`
	Sections        []string                     `json:"sections,omitempty"`
	AllSections     bool                         `json:"all_sections,omitempty"`
	ConnectionState string                       `json:"connection_state,omitempty"`
	Error           string                       `json:"error,omitempty"`
	UpdatedAt       string                       `json:"updated_at,omitempty"`
	Events          map[string]liveEventSnapshot `json:"events,omitempty"`
	SportStates     map[string]liveSportSnapshot `json:"sport_states,omitempty"`
}

type liveResourceWatcher struct {
	mu     sync.RWMutex
	refs   int
	cancel context.CancelFunc
	state  liveResourceState
}

type liveHub struct {
	server *mcp.Server
	client lookups.Service

	mu             sync.Mutex
	watchers       map[string]*liveResourceWatcher
	sessionSubs    map[string]map[string]struct{}
	sessionWaiters map[string]struct{}
}

func newLiveHub(client lookups.Service) *liveHub {
	return &liveHub{
		client:         client,
		watchers:       map[string]*liveResourceWatcher{},
		sessionSubs:    map[string]map[string]struct{}{},
		sessionWaiters: map[string]struct{}{},
	}
}

func registerLiveResources(server *mcp.Server, hub *liveHub) {
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "live-event",
		Description: "Live state for one event. Subscribe, then read the resource to get the latest snapshot. Optional query params: section and all_sections=1.",
		MIMEType:    "application/json",
		URITemplate: "sports://live/event/{event_id}{?section,all_sections}",
	}, hub.readResource)

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "live-events",
		Description: "Live state for multiple events. Use comma-separated ids in the URI path.",
		MIMEType:    "application/json",
		URITemplate: "sports://live/events/{event_ids}{?section,all_sections}",
	}, hub.readResource)

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "live-sport",
		Description: "Live state for one sport. Subscribe, then read the resource to get the latest snapshot.",
		MIMEType:    "application/json",
		URITemplate: "sports://live/sport/{sport}",
	}, hub.readResource)

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "live-sports",
		Description: "Live state for multiple sports. Use comma-separated slugs in the URI path.",
		MIMEType:    "application/json",
		URITemplate: "sports://live/sports/{sports}",
	}, hub.readResource)
}

func (h *liveHub) subscribe(ctx context.Context, req *mcp.SubscribeRequest) error {
	config, err := parseLiveResourceURI(req.Params.URI)
	if err != nil {
		return err
	}

	sessionKey := resourceSessionKey(req.Session)

	h.mu.Lock()
	h.trackSessionLocked(sessionKey, req.Session)

	if watcher, ok := h.watchers[config.URI]; ok {
		watcher.refs++
		h.addSessionSubLocked(sessionKey, config.URI)
		h.mu.Unlock()
		return nil
	}

	watchCtx, cancel := context.WithCancel(context.Background())
	watcher := &liveResourceWatcher{
		refs:   1,
		cancel: cancel,
		state:  newLiveResourceState(config),
	}
	h.watchers[config.URI] = watcher
	h.addSessionSubLocked(sessionKey, config.URI)
	h.mu.Unlock()

	go h.runWatcher(watchCtx, watcher, config)
	return nil
}

func (h *liveHub) unsubscribe(_ context.Context, req *mcp.UnsubscribeRequest) error {
	config, err := parseLiveResourceURI(req.Params.URI)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	sessionKey := resourceSessionKey(req.Session)
	if subs := h.sessionSubs[sessionKey]; subs != nil {
		delete(subs, config.URI)
		if len(subs) == 0 {
			delete(h.sessionSubs, sessionKey)
		}
	}
	h.releaseWatcherLocked(config.URI)
	return nil
}

func (h *liveHub) readResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	config, err := parseLiveResourceURI(req.Params.URI)
	if err != nil {
		return nil, err
	}

	h.mu.Lock()
	watcher := h.watchers[config.URI]
	h.mu.Unlock()
	if watcher == nil {
		return nil, fmt.Errorf("live resource %s is not active; subscribe first", config.URI)
	}

	watcher.mu.RLock()
	state := watcher.state
	watcher.mu.RUnlock()

	data, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      config.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

func (h *liveHub) runWatcher(ctx context.Context, watcher *liveResourceWatcher, config liveResourceConfig) {
	emit := func(record sofascoreapi.WatchRecord) error {
		watcher.mu.Lock()
		applyWatchRecord(&watcher.state, record)
		watcher.mu.Unlock()

		return h.server.ResourceUpdated(context.Background(), &mcp.ResourceUpdatedNotificationParams{
			URI: config.URI,
		})
	}

	var err error
	switch config.WatchKind {
	case sofascoreapi.WatchKindEvent:
		err = h.client.WatchEvents(ctx, config.EventIDs, config.Sections, config.AllSections, emit)
	case sofascoreapi.WatchKindSport:
		err = h.client.WatchSports(ctx, config.Sports, emit)
	default:
		err = fmt.Errorf("unsupported live watch kind %q", config.WatchKind)
	}

	if err != nil && ctx.Err() == nil {
		watcher.mu.Lock()
		applyWatchRecord(&watcher.state, sofascoreapi.WatchRecord{
			Type:      sofascoreapi.WatchRecordError,
			WatchKind: config.WatchKind,
			Error:     err.Error(),
		})
		watcher.mu.Unlock()
		_ = h.server.ResourceUpdated(context.Background(), &mcp.ResourceUpdatedNotificationParams{URI: config.URI})
	}
}

func (h *liveHub) cleanupSession(sessionKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs := h.sessionSubs[sessionKey]
	for uri := range subs {
		h.releaseWatcherLocked(uri)
	}
	delete(h.sessionSubs, sessionKey)
	delete(h.sessionWaiters, sessionKey)
}

func (h *liveHub) trackSessionLocked(sessionKey string, session *mcp.ServerSession) {
	if session == nil {
		return
	}
	if _, ok := h.sessionWaiters[sessionKey]; ok {
		return
	}
	h.sessionWaiters[sessionKey] = struct{}{}
	go func() {
		_ = session.Wait()
		h.cleanupSession(sessionKey)
	}()
}

func (h *liveHub) addSessionSubLocked(sessionKey, uri string) {
	if sessionKey == "" {
		return
	}
	subs := h.sessionSubs[sessionKey]
	if subs == nil {
		subs = map[string]struct{}{}
		h.sessionSubs[sessionKey] = subs
	}
	subs[uri] = struct{}{}
}

func (h *liveHub) releaseWatcherLocked(uri string) {
	watcher := h.watchers[uri]
	if watcher == nil {
		return
	}
	watcher.refs--
	if watcher.refs > 0 {
		return
	}
	delete(h.watchers, uri)
	watcher.cancel()
}

func parseLiveResourceURI(raw string) (liveResourceConfig, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return liveResourceConfig{}, err
	}
	if parsed.Scheme != "sports" || parsed.Host != "live" {
		return liveResourceConfig{}, fmt.Errorf("unsupported live resource uri %q", raw)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 {
		return liveResourceConfig{}, fmt.Errorf("unsupported live resource path %q", parsed.Path)
	}

	config := liveResourceConfig{URI: raw}
	switch parts[0] {
	case "event":
		eventIDs, err := parseEventIDs(parts[1])
		if err != nil || len(eventIDs) != 1 {
			return liveResourceConfig{}, fmt.Errorf("live event uri requires one event id")
		}
		config.WatchKind = sofascoreapi.WatchKindEvent
		config.EventIDs = eventIDs
	case "events":
		eventIDs, err := parseEventIDs(parts[1])
		if err != nil {
			return liveResourceConfig{}, err
		}
		config.WatchKind = sofascoreapi.WatchKindEvent
		config.EventIDs = eventIDs
	case "sport":
		sports := parseCSV(parts[1])
		if len(sports) != 1 {
			return liveResourceConfig{}, fmt.Errorf("live sport uri requires one sport slug")
		}
		config.WatchKind = sofascoreapi.WatchKindSport
		config.Sports = sports
	case "sports":
		sports := parseCSV(parts[1])
		if len(sports) == 0 {
			return liveResourceConfig{}, fmt.Errorf("live sports uri requires at least one sport slug")
		}
		config.WatchKind = sofascoreapi.WatchKindSport
		config.Sports = sports
	default:
		return liveResourceConfig{}, fmt.Errorf("unsupported live resource kind %q", parts[0])
	}

	query := parsed.Query()
	config.Sections = uniqueStrings(query["section"])
	config.AllSections = parseBool(query.Get("all_sections"))
	if config.WatchKind != sofascoreapi.WatchKindEvent && (config.AllSections || len(config.Sections) > 0) {
		return liveResourceConfig{}, fmt.Errorf("section query parameters are only supported for event live resources")
	}
	if config.AllSections && len(config.Sections) > 0 {
		return liveResourceConfig{}, fmt.Errorf("all_sections cannot be combined with section")
	}

	return config, nil
}

func newLiveResourceState(config liveResourceConfig) liveResourceState {
	state := liveResourceState{
		URI:             config.URI,
		WatchKind:       config.WatchKind,
		EventIDs:        append([]int(nil), config.EventIDs...),
		Sports:          append([]string(nil), config.Sports...),
		Sections:        append([]string(nil), config.Sections...),
		AllSections:     config.AllSections,
		ConnectionState: "connecting",
	}
	if config.WatchKind == sofascoreapi.WatchKindEvent {
		state.Events = map[string]liveEventSnapshot{}
	}
	if config.WatchKind == sofascoreapi.WatchKindSport {
		state.SportStates = map[string]liveSportSnapshot{}
	}
	return state
}

func applyWatchRecord(state *liveResourceState, record sofascoreapi.WatchRecord) {
	state.UpdatedAt = record.At

	switch record.Type {
	case sofascoreapi.WatchRecordStatus:
		state.ConnectionState = record.State
		if record.Error != "" {
			state.Error = record.Error
		}
		return
	case sofascoreapi.WatchRecordError:
		state.Error = record.Error
		return
	}

	switch record.WatchKind {
	case sofascoreapi.WatchKindEvent:
		if state.Events == nil {
			state.Events = map[string]liveEventSnapshot{}
		}
		key := strconv.Itoa(record.EventID)
		snapshot := state.Events[key]
		snapshot.LastRecord = record.Type
		snapshot.UpdatedAt = record.At
		if record.Summary != nil {
			copy := *record.Summary
			snapshot.Summary = &copy
		}
		if record.Event != nil {
			snapshot.Event = cloneMapAny(record.Event)
		}
		if record.Sections != nil {
			snapshot.Sections = cloneMapAny(record.Sections)
		}
		if record.SectionErrors != nil {
			snapshot.SectionErrors = cloneStringMap(record.SectionErrors)
		}
		if record.Type == sofascoreapi.WatchRecordUpdate {
			snapshot.LastPatch = cloneMapAny(record.Patch)
			snapshot.ChangedFields = append([]string(nil), record.ChangedFields...)
		}
		if record.Type == sofascoreapi.WatchRecordSectionRefresh {
			if snapshot.Sections == nil {
				snapshot.Sections = map[string]any{}
			}
			if snapshot.SectionErrors == nil {
				snapshot.SectionErrors = map[string]string{}
			}
			if record.SectionError != "" {
				delete(snapshot.Sections, record.Section)
				snapshot.SectionErrors[record.Section] = record.SectionError
			} else {
				snapshot.Sections[record.Section] = record.SectionData
				delete(snapshot.SectionErrors, record.Section)
			}
		}
		state.Events[key] = snapshot
	case sofascoreapi.WatchKindSport:
		if state.SportStates == nil {
			state.SportStates = map[string]liveSportSnapshot{}
		}
		snapshot := state.SportStates[record.Sport]
		snapshot.Sport = record.Sport
		if snapshot.Events == nil {
			snapshot.Events = map[string]sofascoreapi.WatchEventSummary{}
		}
		snapshot.UpdatedAt = record.At
		if record.Type == sofascoreapi.WatchRecordSnapshot {
			snapshot.Events = map[string]sofascoreapi.WatchEventSummary{}
			for _, event := range record.Events {
				snapshot.Events[strconv.Itoa(event.EventID)] = event
			}
		}
		if record.Type == sofascoreapi.WatchRecordUpdate && record.Summary != nil {
			snapshot.Events[strconv.Itoa(record.Summary.EventID)] = *record.Summary
			snapshot.LastEventID = record.EventID
			snapshot.LastPatch = cloneMapAny(record.Patch)
			snapshot.ChangedFields = append([]string(nil), record.ChangedFields...)
		}
		state.SportStates[record.Sport] = snapshot
	}
}

func resourceSessionKey(session *mcp.ServerSession) string {
	if session == nil {
		return ""
	}
	return fmt.Sprintf("%p", session)
}

func parseEventIDs(value string) ([]int, error) {
	parts := parseCSV(value)
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		eventID, err := strconv.Atoi(part)
		if err != nil || eventID <= 0 {
			return nil, fmt.Errorf("invalid event id %q", part)
		}
		out = append(out, eventID)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one event id is required")
	}
	return uniquePositiveInts(out), nil
}

func parseCSV(value string) []string {
	parts := strings.Split(strings.TrimSpace(value), ",")
	return uniqueStrings(parts)
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func uniquePositiveInts(values []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func cloneMapAny(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	out := make(map[string]any, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}

func cloneStringMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	out := make(map[string]string, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
