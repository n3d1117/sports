package commands

import (
	"fmt"
	"sort"
	"strings"

	"sports/internal/output"
	"sports/internal/provider/sofascore"
)

func (a *App) emitWatchRecord(record sofascoreapi.WatchRecord, jsonOutput bool) error {
	if jsonOutput {
		return output.NDJSON(a.Stdout, record)
	}

	switch record.Type {
	case sofascoreapi.WatchRecordSnapshot:
		return a.emitWatchSnapshot(record)
	case sofascoreapi.WatchRecordUpdate:
		return a.emitWatchUpdate(record)
	case sofascoreapi.WatchRecordSectionRefresh:
		return a.emitWatchSectionRefresh(record)
	case sofascoreapi.WatchRecordStatus:
		_, err := fmt.Fprintf(a.Stderr, "[%s] %s\n", printable(record.At), printable(record.State))
		return err
	case sofascoreapi.WatchRecordError:
		if record.Error == "" {
			return nil
		}
		output.Errorf(a.Stderr, "%s", record.Error)
		return nil
	default:
		return nil
	}
}

func (a *App) emitWatchSnapshot(record sofascoreapi.WatchRecord) error {
	switch record.WatchKind {
	case sofascoreapi.WatchKindEvent:
		if record.Summary == nil {
			return nil
		}
		_, err := fmt.Fprintf(
			a.Stdout,
			"WATCH EVENT %d [%s]\nMATCH: %s vs %s\nSTATUS: %s\nSTART: %s\nSCORE: %s\nTOURNAMENT: %s\n",
			record.Summary.EventID,
			printable(record.Summary.Sport),
			printable(record.Summary.Home),
			printable(record.Summary.Away),
			printable(coalesce(record.Summary.StatusDescription, record.Summary.StatusType)),
			printable(record.Summary.StartTime),
			printable(formatScore(record.Summary.HomeScore, record.Summary.AwayScore)),
			printable(record.Summary.Tournament),
		)
		if err != nil {
			return err
		}
		if len(record.Sections) > 0 || len(record.SectionErrors) > 0 {
			sectionNames := make([]string, 0, len(record.Sections))
			for name := range record.Sections {
				sectionNames = append(sectionNames, name)
			}
			sort.Strings(sectionNames)
			if _, err := fmt.Fprintf(a.Stdout, "SECTIONS: %s\n", printable(strings.Join(sectionNames, ", "))); err != nil {
				return err
			}
			if len(record.SectionErrors) > 0 {
				errorNames := make([]string, 0, len(record.SectionErrors))
				for name := range record.SectionErrors {
					errorNames = append(errorNames, fmt.Sprintf("%s (%s)", name, record.SectionErrors[name]))
				}
				sort.Strings(errorNames)
				if _, err := fmt.Fprintf(a.Stdout, "SECTION ERRORS: %s\n", strings.Join(errorNames, ", ")); err != nil {
					return err
				}
			}
		}
		_, err = fmt.Fprintln(a.Stdout)
		return err
	case sofascoreapi.WatchKindSport:
		if _, err := fmt.Fprintf(a.Stdout, "WATCH SPORT %s (%d events)\n", printable(record.Sport), len(record.Events)); err != nil {
			return err
		}
		for _, event := range record.Events {
			if _, err := fmt.Fprintf(
				a.Stdout,
				"- %d %s %s vs %s [%s] %s\n",
				event.EventID,
				printable(event.StartTime),
				printable(event.Home),
				printable(event.Away),
				printable(coalesce(event.StatusDescription, event.StatusType)),
				printable(formatScore(event.HomeScore, event.AwayScore)),
			); err != nil {
				return err
			}
		}
		_, err := fmt.Fprintln(a.Stdout)
		return err
	default:
		return nil
	}
}

func (a *App) emitWatchUpdate(record sofascoreapi.WatchRecord) error {
	parts := make([]string, 0, len(record.ChangedFields))
	for _, key := range record.ChangedFields {
		parts = append(parts, fmt.Sprintf("%s=%v", key, record.Patch[key]))
	}
	label := record.Subject
	if record.Summary != nil {
		label = fmt.Sprintf("%d %s vs %s", record.Summary.EventID, printable(record.Summary.Home), printable(record.Summary.Away))
	}
	_, err := fmt.Fprintf(a.Stdout, "[%s] %s %s\n", printable(record.At), printable(label), strings.Join(parts, ", "))
	return err
}

func (a *App) emitWatchSectionRefresh(record sofascoreapi.WatchRecord) error {
	if record.SectionError != "" {
		_, err := fmt.Fprintf(a.Stdout, "[%s] %s section %s error: %s\n", printable(record.At), printable(record.Subject), printable(record.Section), record.SectionError)
		return err
	}
	_, err := fmt.Fprintf(a.Stdout, "[%s] %s section %s refreshed\n", printable(record.At), printable(record.Subject), printable(record.Section))
	return err
}
