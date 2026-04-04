package sofascoreapi

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	liveWebsocketURL        = "wss://ws.sofascore.com:9222"
	liveSectionsRefreshTick = 15 * time.Second
)

type WatchKind string

const (
	WatchKindEvent WatchKind = "event"
	WatchKindSport WatchKind = "sport"
)

type WatchRecordType string

const (
	WatchRecordSnapshot       WatchRecordType = "snapshot"
	WatchRecordUpdate         WatchRecordType = "update"
	WatchRecordSectionRefresh WatchRecordType = "section_refresh"
	WatchRecordStatus         WatchRecordType = "status"
	WatchRecordError          WatchRecordType = "error"
)

type WatchEventSummary struct {
	EventID           int    `json:"event_id"`
	StartTime         string `json:"start_time,omitempty"`
	StatusType        string `json:"status_type,omitempty"`
	StatusDescription string `json:"status_description,omitempty"`
	Home              string `json:"home,omitempty"`
	Away              string `json:"away,omitempty"`
	Tournament        string `json:"tournament,omitempty"`
	Sport             string `json:"sport,omitempty"`
	HomeScore         *int   `json:"home_score,omitempty"`
	AwayScore         *int   `json:"away_score,omitempty"`
}

type WatchRecord struct {
	Type          WatchRecordType     `json:"type"`
	WatchKind     WatchKind           `json:"watch_kind"`
	Subject       string              `json:"subject,omitempty"`
	EventID       int                 `json:"event_id,omitempty"`
	Sport         string              `json:"sport,omitempty"`
	At            string              `json:"at,omitempty"`
	State         string              `json:"state,omitempty"`
	Error         string              `json:"error,omitempty"`
	Summary       *WatchEventSummary  `json:"summary,omitempty"`
	Events        []WatchEventSummary `json:"events,omitempty"`
	Event         map[string]any      `json:"event,omitempty"`
	Sections      map[string]any      `json:"sections,omitempty"`
	SectionErrors map[string]string   `json:"section_errors,omitempty"`
	ChangedFields []string            `json:"changed_fields,omitempty"`
	Patch         map[string]any      `json:"patch,omitempty"`
	Section       string              `json:"section,omitempty"`
	SectionData   any                 `json:"section_data,omitempty"`
	SectionError  string              `json:"section_error,omitempty"`
}

type watchedEventState struct {
	mu            sync.RWMutex
	subject       string
	eventID       int
	event         map[string]any
	summary       WatchEventSummary
	sections      []string
	sectionData   map[string]any
	sectionErrors map[string]string
	sectionHashes map[string]string
}

type watchedSportState struct {
	mu        sync.RWMutex
	sport     string
	subject   string
	eventMaps map[int]map[string]any
	summaries map[int]WatchEventSummary
}

func (c *Client) WatchEvents(ctx context.Context, eventIDs []int, sections []string, allSections bool, emit func(WatchRecord) error) error {
	eventIDs = uniquePositiveInts(eventIDs)
	if len(eventIDs) == 0 {
		return fmt.Errorf("at least one event id is required")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	recordCh := make(chan WatchRecord, 256)
	emitErrCh := make(chan error, 1)
	go emitWatchRecords(ctx, cancel, recordCh, emit, emitErrCh)

	send := func(record WatchRecord) {
		record.At = time.Now().UTC().Format(time.RFC3339)
		select {
		case recordCh <- record:
		case <-ctx.Done():
		}
	}

	states := make(map[int]*watchedEventState, len(eventIDs))
	for _, eventID := range eventIDs {
		event, err := c.Event(ctx, eventID)
		if err != nil {
			return err
		}

		eventMap, err := decodeJSONObject(event.Raw)
		if err != nil {
			return err
		}

		requestedSections := append([]string(nil), sections...)
		if allSections {
			requestedSections, err = c.ProbeEventSections(ctx, eventID)
			if err != nil {
				return err
			}
		}
		requestedSections = uniqueStrings(requestedSections)

		sectionData, sectionErrors := c.loadEventSections(ctx, eventID, requestedSections)
		sectionHashes := make(map[string]string, len(sectionData))
		for name, value := range sectionData {
			sectionHashes[name] = mustJSON(value)
		}
		for name, value := range sectionErrors {
			sectionHashes[name] = "error:" + value
		}

		state := &watchedEventState{
			subject:       eventSubject(eventID),
			eventID:       eventID,
			event:         eventMap,
			summary:       summaryFromEventMap(eventMap),
			sections:      requestedSections,
			sectionData:   sectionData,
			sectionErrors: sectionErrors,
			sectionHashes: sectionHashes,
		}
		states[eventID] = state

		send(WatchRecord{
			Type:          WatchRecordSnapshot,
			WatchKind:     WatchKindEvent,
			Subject:       state.subject,
			EventID:       eventID,
			Sport:         state.summary.Sport,
			Summary:       cloneSummary(state.summary),
			Event:         cloneMap(state.event),
			Sections:      cloneMapAny(state.sectionData),
			SectionErrors: maps.Clone(state.sectionErrors),
		})
	}

	for _, state := range states {
		if len(state.sections) == 0 {
			continue
		}
		go c.runEventSectionRefreshLoop(ctx, state, send)
	}

	conn, err := c.newLiveConn(send)
	if err != nil {
		return err
	}
	defer conn.close()

	for _, state := range states {
		state := state
		if err := conn.subscribe(state.subject, func(payload []byte) {
			patchEventID, patch, expanded, err := parseLivePatch(payload)
			if err != nil {
				send(WatchRecord{
					Type:      WatchRecordError,
					WatchKind: WatchKindEvent,
					Subject:   state.subject,
					EventID:   state.eventID,
					Error:     err.Error(),
				})
				return
			}
			if patchEventID != state.eventID {
				return
			}

			changedFields := sortedKeys(patch)

			state.mu.Lock()
			state.event = deepMergeObjects(state.event, expanded)
			state.summary = summaryFromEventMap(state.event)
			summary := cloneSummary(state.summary)
			state.mu.Unlock()

			send(WatchRecord{
				Type:          WatchRecordUpdate,
				WatchKind:     WatchKindEvent,
				Subject:       state.subject,
				EventID:       state.eventID,
				Sport:         summary.Sport,
				ChangedFields: changedFields,
				Patch:         patch,
				Summary:       summary,
			})
		}); err != nil {
			return err
		}
	}

	return waitWatch(ctx, emitErrCh)
}

func (c *Client) WatchSports(ctx context.Context, sports []string, emit func(WatchRecord) error) error {
	sports = uniqueStrings(sports)
	if len(sports) == 0 {
		return fmt.Errorf("at least one sport slug is required")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	recordCh := make(chan WatchRecord, 256)
	emitErrCh := make(chan error, 1)
	go emitWatchRecords(ctx, cancel, recordCh, emit, emitErrCh)

	send := func(record WatchRecord) {
		record.At = time.Now().UTC().Format(time.RFC3339)
		select {
		case recordCh <- record:
		case <-ctx.Done():
		}
	}

	states := make(map[string]*watchedSportState, len(sports))
	for _, sport := range sports {
		scheduledEvents, err := c.sportScheduledEventMaps(ctx, sport, c.now().UTC().Format("2006-01-02"))
		if err != nil {
			return err
		}
		state := &watchedSportState{
			sport:     sport,
			subject:   sportSubject(sport),
			eventMaps: make(map[int]map[string]any, len(scheduledEvents)),
			summaries: make(map[int]WatchEventSummary, len(scheduledEvents)),
		}
		for eventID, eventMap := range scheduledEvents {
			state.eventMaps[eventID] = eventMap
			state.summaries[eventID] = summaryFromEventMap(eventMap)
		}
		states[sport] = state

		send(WatchRecord{
			Type:      WatchRecordSnapshot,
			WatchKind: WatchKindSport,
			Subject:   state.subject,
			Sport:     sport,
			Events:    sportSnapshot(state),
		})
	}

	conn, err := c.newLiveConn(send)
	if err != nil {
		return err
	}
	defer conn.close()

	for _, state := range states {
		state := state
		if err := conn.subscribe(state.subject, func(payload []byte) {
			eventID, patch, expanded, err := parseLivePatch(payload)
			if err != nil {
				send(WatchRecord{
					Type:      WatchRecordError,
					WatchKind: WatchKindSport,
					Subject:   state.subject,
					Sport:     state.sport,
					Error:     err.Error(),
				})
				return
			}

			state.mu.Lock()
			eventMap, ok := state.eventMaps[eventID]
			state.mu.Unlock()

			if !ok {
				event, err := c.Event(ctx, eventID)
				if err != nil {
					send(WatchRecord{
						Type:      WatchRecordError,
						WatchKind: WatchKindSport,
						Subject:   state.subject,
						Sport:     state.sport,
						EventID:   eventID,
						Error:     err.Error(),
					})
					return
				}
				eventMap, err = decodeJSONObject(event.Raw)
				if err != nil {
					send(WatchRecord{
						Type:      WatchRecordError,
						WatchKind: WatchKindSport,
						Subject:   state.subject,
						Sport:     state.sport,
						EventID:   eventID,
						Error:     err.Error(),
					})
					return
				}
			}

			state.mu.Lock()
			state.eventMaps[eventID] = deepMergeObjects(eventMap, expanded)
			state.summaries[eventID] = summaryFromEventMap(state.eventMaps[eventID])
			summary := cloneSummary(state.summaries[eventID])
			state.mu.Unlock()

			send(WatchRecord{
				Type:          WatchRecordUpdate,
				WatchKind:     WatchKindSport,
				Subject:       state.subject,
				Sport:         state.sport,
				EventID:       eventID,
				ChangedFields: sortedKeys(patch),
				Patch:         patch,
				Summary:       summary,
			})
		}); err != nil {
			return err
		}
	}

	return waitWatch(ctx, emitErrCh)
}

type liveConn struct {
	nc   *nats.Conn
	send func(WatchRecord)
}

func (c *Client) newLiveConn(send func(WatchRecord)) (*liveConn, error) {
	conn := &liveConn{send: send}
	nc, err := nats.Connect(
		liveWebsocketURL,
		nats.UserInfo("none", "none"),
		nats.Name("sofascore-live"),
		nats.Timeout(10*time.Second),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			conn.send(WatchRecord{
				Type:  WatchRecordStatus,
				State: "reconnecting",
				Error: errorString(err),
			})
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			conn.send(WatchRecord{
				Type:  WatchRecordStatus,
				State: "connected",
			})
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			conn.send(WatchRecord{
				Type:  WatchRecordStatus,
				State: "closed",
			})
		}),
	)
	if err != nil {
		return nil, err
	}
	conn.nc = nc
	send(WatchRecord{Type: WatchRecordStatus, State: "connected"})
	return conn, nil
}

func (c *liveConn) subscribe(subject string, handler func([]byte)) error {
	_, err := c.nc.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	if err != nil {
		return err
	}
	return c.nc.Flush()
}

func (c *liveConn) close() {
	if c.nc != nil {
		c.nc.Close()
	}
}

func emitWatchRecords(ctx context.Context, cancel context.CancelFunc, records <-chan WatchRecord, emit func(WatchRecord) error, emitErrCh chan<- error) {
	defer close(emitErrCh)
	for {
		select {
		case <-ctx.Done():
			return
		case record := <-records:
			if err := emit(record); err != nil {
				select {
				case emitErrCh <- err:
				default:
				}
				cancel()
				return
			}
		}
	}
}

func waitWatch(ctx context.Context, emitErrCh <-chan error) error {
	select {
	case <-ctx.Done():
		select {
		case err := <-emitErrCh:
			if err != nil {
				return err
			}
		default:
		}
		return nil
	case err := <-emitErrCh:
		return err
	}
}

func (c *Client) runEventSectionRefreshLoop(ctx context.Context, state *watchedEventState, send func(WatchRecord)) {
	ticker := time.NewTicker(liveSectionsRefreshTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			state.mu.RLock()
			live := isLiveStatus(state.summary.StatusType)
			eventID := state.eventID
			subject := state.subject
			sport := state.summary.Sport
			sections := append([]string(nil), state.sections...)
			state.mu.RUnlock()
			if !live {
				continue
			}

			for _, section := range sections {
				payload, err := c.EventSection(ctx, eventID, section)
				if err != nil {
					msg := err.Error()
					state.mu.Lock()
					if state.sectionHashes[section] == "error:"+msg {
						state.mu.Unlock()
						continue
					}
					state.sectionHashes[section] = "error:" + msg
					state.sectionErrors[section] = msg
					delete(state.sectionData, section)
					state.mu.Unlock()

					send(WatchRecord{
						Type:         WatchRecordSectionRefresh,
						WatchKind:    WatchKindEvent,
						Subject:      subject,
						EventID:      eventID,
						Sport:        sport,
						Section:      section,
						SectionError: msg,
					})
					continue
				}

				value, err := decodeJSONAny(payload)
				if err != nil {
					send(WatchRecord{
						Type:         WatchRecordSectionRefresh,
						WatchKind:    WatchKindEvent,
						Subject:      subject,
						EventID:      eventID,
						Sport:        sport,
						Section:      section,
						SectionError: err.Error(),
					})
					continue
				}
				hash := mustJSON(value)

				state.mu.Lock()
				if state.sectionHashes[section] == hash {
					state.mu.Unlock()
					continue
				}
				state.sectionHashes[section] = hash
				state.sectionData[section] = value
				delete(state.sectionErrors, section)
				state.mu.Unlock()

				send(WatchRecord{
					Type:        WatchRecordSectionRefresh,
					WatchKind:   WatchKindEvent,
					Subject:     subject,
					EventID:     eventID,
					Sport:       sport,
					Section:     section,
					SectionData: value,
				})
			}
		}
	}
}

func (c *Client) loadEventSections(ctx context.Context, eventID int, sections []string) (map[string]any, map[string]string) {
	sectionData := make(map[string]any, len(sections))
	sectionErrors := make(map[string]string)
	for _, section := range sections {
		payload, err := c.EventSection(ctx, eventID, section)
		if err != nil {
			sectionErrors[section] = err.Error()
			continue
		}
		value, err := decodeJSONAny(payload)
		if err != nil {
			sectionErrors[section] = err.Error()
			continue
		}
		sectionData[section] = value
	}
	return sectionData, sectionErrors
}

func (c *Client) sportScheduledEventMaps(ctx context.Context, sport, date string) (map[int]map[string]any, error) {
	var response struct {
		Events []json.RawMessage `json:"events"`
	}
	raw, err := c.getRawJSON(ctx, c.baseURL+"/sport/"+url.PathEscape(strings.TrimSpace(sport))+"/scheduled-events/"+date)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}

	events := make(map[int]map[string]any, len(response.Events))
	for _, item := range response.Events {
		eventMap, err := decodeJSONObject(item)
		if err != nil {
			return nil, err
		}
		eventID := getInt(eventMap, "id")
		if eventID <= 0 {
			continue
		}
		events[eventID] = eventMap
	}
	return events, nil
}

func parseLivePatch(payload []byte) (int, map[string]any, map[string]any, error) {
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return 0, nil, nil, err
	}

	eventID := 0
	switch value := decoded["id"].(type) {
	case float64:
		eventID = int(value)
	case string:
		eventID, _ = strconv.Atoi(value)
	}
	if eventID == 0 {
		switch value := decoded["eventId"].(type) {
		case float64:
			eventID = int(value)
		case string:
			eventID, _ = strconv.Atoi(value)
		}
	}
	if eventID <= 0 {
		return 0, nil, nil, fmt.Errorf("live patch missing event id")
	}

	delete(decoded, "id")
	delete(decoded, "eventId")

	flat := cloneMapAny(decoded)
	expanded := expandDottedKeys(decoded)
	return eventID, flat, expanded, nil
}

func expandDottedKeys(flat map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range flat {
		parts := strings.Split(key, ".")
		insertNested(result, parts, value)
	}
	return result
}

func insertNested(target map[string]any, parts []string, value any) {
	if len(parts) == 1 {
		target[parts[0]] = value
		return
	}

	child, _ := target[parts[0]].(map[string]any)
	if child == nil {
		child = make(map[string]any)
		target[parts[0]] = child
	}
	insertNested(child, parts[1:], value)
}

func deepMergeObjects(base, patch map[string]any) map[string]any {
	merged := cloneMap(base)
	for key, value := range patch {
		if patchMap, ok := value.(map[string]any); ok {
			if baseMap, ok := merged[key].(map[string]any); ok {
				merged[key] = deepMergeObjects(baseMap, patchMap)
				continue
			}
			merged[key] = deepMergeObjects(map[string]any{}, patchMap)
			continue
		}
		merged[key] = value
	}
	return merged
}

func summaryFromEventMap(event map[string]any) WatchEventSummary {
	startTimestamp := getInt(event, "startTimestamp")
	var startTime string
	if startTimestamp > 0 {
		startTime = time.Unix(int64(startTimestamp), 0).UTC().Format(time.RFC3339)
	}
	return WatchEventSummary{
		EventID:           getInt(event, "id"),
		StartTime:         startTime,
		StatusType:        getString(event, "status", "type"),
		StatusDescription: getString(event, "status", "description"),
		Home:              coalesce(getString(event, "homeTeam", "name"), getString(event, "homePlayer", "name"), "TBD"),
		Away:              coalesce(getString(event, "awayTeam", "name"), getString(event, "awayPlayer", "name"), "TBD"),
		Tournament:        getString(event, "tournament", "name"),
		Sport:             getString(event, "tournament", "category", "sport", "slug"),
		HomeScore:         getIntPtr(event, "homeScore", "current"),
		AwayScore:         getIntPtr(event, "awayScore", "current"),
	}
}

func sportSnapshot(state *watchedSportState) []WatchEventSummary {
	state.mu.RLock()
	defer state.mu.RUnlock()

	events := make([]WatchEventSummary, 0, len(state.summaries))
	for _, summary := range state.summaries {
		events = append(events, summary)
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime == events[j].StartTime {
			return events[i].EventID < events[j].EventID
		}
		return events[i].StartTime < events[j].StartTime
	})
	return events
}

func isLiveStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "inprogress")
}

func eventSubject(eventID int) string {
	return "event." + strconv.Itoa(eventID)
}

func sportSubject(sport string) string {
	return "sport." + strings.TrimSpace(sport)
}

func uniquePositiveInts(values []int) []int {
	seen := make(map[int]struct{}, len(values))
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
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
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
	return out
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func decodeJSONObject(data []byte) (map[string]any, error) {
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func decodeJSONAny(data []byte) (any, error) {
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func mustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	out := make(map[string]any, len(source))
	for key, value := range source {
		if child, ok := value.(map[string]any); ok {
			out[key] = cloneMap(child)
			continue
		}
		out[key] = value
	}
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

func cloneSummary(summary WatchEventSummary) *WatchEventSummary {
	copy := summary
	return &copy
}

func getMap(value map[string]any, key string) map[string]any {
	child, _ := value[key].(map[string]any)
	return child
}

func getAny(value map[string]any, path ...string) any {
	current := value
	for i, part := range path {
		item, ok := current[part]
		if !ok {
			return nil
		}
		if i == len(path)-1 {
			return item
		}
		next, ok := item.(map[string]any)
		if !ok {
			return nil
		}
		current = next
	}
	return nil
}

func getString(value map[string]any, path ...string) string {
	item := getAny(value, path...)
	switch typed := item.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func getInt(value map[string]any, path ...string) int {
	item := getAny(value, path...)
	switch typed := item.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		result, _ := strconv.Atoi(typed)
		return result
	default:
		return 0
	}
}

func getIntPtr(value map[string]any, path ...string) *int {
	item := getAny(value, path...)
	switch typed := item.(type) {
	case float64:
		result := int(typed)
		return &result
	case int:
		result := typed
		return &result
	case string:
		result, err := strconv.Atoi(typed)
		if err != nil {
			return nil
		}
		return &result
	default:
		return nil
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
