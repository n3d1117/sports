package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/kong"

	"sports/internal/buildinfo"
	"sports/internal/lookups"
	"sports/internal/output"
	"sports/internal/provider/sofascore"
)

var supportedSportSlugs = map[string]struct{}{
	"football":   {},
	"basketball": {},
	"tennis":     {},
}

type SofaScoreService = lookups.Service

type terminalDetector func(io.Writer) bool

type App struct {
	Client     SofaScoreService
	Stdout     io.Writer
	Stderr     io.Writer
	Context    context.Context
	isTerminal terminalDetector
}

type exitError struct {
	Code    int
	Message string
}

type parserExit struct {
	Code int
}

type CLI struct {
	Version    kong.VersionFlag `name:"version" help:"Print version and exit."`
	Search     SearchCmd        `cmd:"" help:"Resolve ids from names."`
	List       ListCmd          `cmd:"" help:"List supported sports."`
	Football   FootballCmd      `cmd:"" help:"Football schedule, sections, and related feeds."`
	Basketball BasketballCmd    `cmd:"" help:"Basketball schedule, sections, and related feeds."`
	Tennis     TennisCmd        `cmd:"" help:"Tennis schedule, sections, and related feeds."`
	Event      EventCmd         `cmd:"" help:"Event details, sections, TV, H2H, and live updates."`
	Team       TeamCmd          `cmd:"" help:"Team schedules and related lookups."`
	Player     PlayerCmd        `cmd:"" help:"Player lookups."`
	Tournament TournamentCmd    `cmd:"" help:"Tournament metadata, seasons, sections, and events."`
	Trending   TrendingCmd      `cmd:"" help:"Trending events feed."`
}

type SearchCmd struct {
	Query []string `arg:"" name:"query" required:"" help:"Search query."`
	Type  string   `help:"Filter by result type."`
	Sport string   `help:"Filter by sport slug."`
	Page  int      `default:"0" help:"Search results page."`
	Limit int      `default:"10" help:"Limit results."`
	ID    bool     `help:"Print result ids only."`
	JSON  bool     `help:"Print JSON."`
}

type ListCmd struct {
	JSON bool `help:"Print JSON."`
}

type FootballCmd struct {
	Args []string `arg:"" optional:"" passthrough:"" help:"Optional football subcommand or flags."`
}

type BasketballCmd struct {
	Args []string `arg:"" optional:"" passthrough:"" help:"Optional basketball subcommand or flags."`
}

type TennisCmd struct {
	Args []string `arg:"" optional:"" passthrough:"" help:"Optional tennis subcommand or flags."`
}

type EventCmd struct {
	Args []string `arg:"" name:"event-id [subcommand]" required:"" passthrough:"" help:"Event id plus optional subcommand."`
}

type TeamCmd struct {
	Args []string `arg:"" name:"team-id next|last" required:"" passthrough:"" help:"Team id plus direction."`
}

type PlayerCmd struct {
	Args []string `arg:"" name:"player-id attributes" required:"" passthrough:"" help:"Player id plus subcommand."`
}

type TournamentCmd struct {
	Args []string `arg:"" name:"tournament-id [subcommand]" required:"" passthrough:"" help:"Tournament id plus optional subcommand."`
}

type TrendingCmd struct {
	Country string `help:"ISO alpha-2 country code."`
	Date    string `help:"UTC date in YYYY-MM-DD format."`
	Limit   int    `default:"10" help:"Limit results."`
	JSON    bool   `help:"Print JSON."`
}

func Run(args []string, stdout, stderr io.Writer) int {
	app := &App{
		Client:     sofascoreapi.New(""),
		Stdout:     stdout,
		Stderr:     stderr,
		Context:    context.Background(),
		isTerminal: defaultIsTerminal,
	}

	if err := app.Execute(args); err != nil {
		var coded *exitError
		if errors.As(err, &coded) {
			if coded.Message != "" {
				output.Errorf(stderr, "%s", coded.Message)
			}
			return coded.Code
		}

		output.Errorf(stderr, "%v", err)
		return 1
	}

	return 0
}

func (e *exitError) Error() string {
	return e.Message
}

func (a *App) Execute(args []string) (err error) {
	cli := &CLI{}
	parser, err := a.newParser(cli)
	if err != nil {
		return err
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			if exited, ok := recovered.(parserExit); ok {
				if exited.Code == 0 {
					err = nil
					return
				}
				err = &exitError{Code: exited.Code}
				return
			}
			panic(recovered)
		}
	}()

	if len(args) == 0 {
		a.printHelp(parser, []string{"--help"})
		return &exitError{Code: 2, Message: "missing command"}
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			return &exitError{Code: 2, Message: err.Error()}
		}
		return err
	}

	return ctx.Run(a)
}

func (a *App) newParser(cli *CLI) (*kong.Kong, error) {
	return kong.New(
		cli,
		kong.Name("sports"),
		kong.Description("Sports data CLI. Current scope: football, basketball, and tennis. Current provider: SofaScore."),
		kong.Vars{"version": buildinfo.Current()},
		kong.Writers(a.Stdout, a.Stderr),
		kong.Exit(func(code int) {
			panic(parserExit{Code: code})
		}),
	)
}

func (FootballCmd) Help() string {
	return sportHelp("football", true)
}

func (BasketballCmd) Help() string {
	return sportHelp("basketball", true)
}

func (TennisCmd) Help() string {
	return sportHelp("tennis", false)
}

func (EventCmd) Help() string {
	return strings.Join([]string{
		"Surface:",
		"  sports event <event-id> [--json]",
		"  sports event <event-id> sections [--json]",
		"  sports event <event-id> section <name>... [--json]",
		"  sports event <event-id> tv --json",
		"  sports event <event-id> tv-channel <channel-id> --json",
		"  sports event <event-id> h2h-events --json",
		"  sports event <event-id> watch [--section <name>] [--all-sections] [--json]",
	}, "\n")
}

func (TeamCmd) Help() string {
	return strings.Join([]string{
		"Surface:",
		"  sports team <team-id> next|last [--limit N] [--json]",
		"  sports team <team-id> info|tournaments|players|media [--json]",
		"  sports team <team-id> standings [--season <id>] [--json]",
		"  sports team <team-id> stats --tournament <id> --season <id> [--json]",
		"  sports team <team-id> rankings [--tournament <id> --season <id>] [--json]",
		"  sports team <team-id> top-players --tournament <id> --season <id> [--json]",
	}, "\n")
}

func (PlayerCmd) Help() string {
	return strings.Join([]string{
		"Surface:",
		"  sports player <player-id> attributes|media|media-videos|seasons|career [--json]",
		"  sports player <player-id> last [--limit N] [--json]",
		"  sports player <player-id> season-stats|season-ratings --tournament <id> --season <id> [--phase <name>] [--json]",
		"  sports player <player-id> characteristics|national-team|tournaments [--json]",
		"  sports player <player-id> season-heatmap --tournament <id> --season <id> --phase <name> [--json]",
		"  sports player <player-id> penalty-history --tournament <id> --season <id> [--json]",
		"  sports player <player-id> shot-actions|shot-action-areas --tournament <id> --season <id> --phase <name> [--json]",
		"  sports player <player-id> year-stats --year <yyyy> [--json]",
		"  sports player <player-id> featured-event [--json]",
	}, "\n")
}

func (TournamentCmd) Help() string {
	return strings.Join([]string{
		"Surface:",
		"  sports tournament <tournament-id> [--season <id>] [--json]",
		"  sports tournament <tournament-id> sections [--season <id>] [--json]",
		"  sports tournament <tournament-id> section <name>... [--season <id>] [--json]",
		"  sports tournament <tournament-id> seasons [--json]",
		"  sports tournament <tournament-id> next|last [--season <id>] [--limit N] [--json]",
		"  sports tournament <tournament-id> round <n> [--season <id>] [--slug <slug>] [--limit N] [--json]",
		"  sports tournament <tournament-id> scheduled [--date YYYY-MM-DD] [--limit N] [--json]",
	}, "\n")
}

func sportHelp(sport string, hasHomeFeeds bool) string {
	lines := []string{
		"Surface:",
		fmt.Sprintf("  sports %s [--date YYYY-MM-DD] [--json]", sport),
		fmt.Sprintf("  sports %s tournaments [--date YYYY-MM-DD] [--page N] [--json]", sport),
		fmt.Sprintf("  sports %s sections [--json]", sport),
		fmt.Sprintf("  sports %s watch [--json]", sport),
	}
	if hasHomeFeeds {
		lines = append(lines,
			fmt.Sprintf("  sports %s live-tournaments [--json]", sport),
			fmt.Sprintf("  sports %s categories [--json]", sport),
			fmt.Sprintf("  sports %s top-players [--json]", sport),
		)
	}
	return strings.Join(lines, "\n")
}

func (a *App) printHelp(parser *kong.Kong, args []string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if _, ok := recovered.(parserExit); ok {
				return
			}
			panic(recovered)
		}
	}()
	_, _ = parser.Parse(args)
}

func (a *App) outputJSON(force bool) bool {
	if force {
		return true
	}
	return !a.stdoutIsTerminal()
}

func (a *App) stdoutIsTerminal() bool {
	if a.isTerminal == nil {
		return defaultIsTerminal(a.Stdout)
	}
	return a.isTerminal(a.Stdout)
}

func defaultIsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func (c *SearchCmd) Run(a *App) error {
	if c.ID && c.JSON {
		return &exitError{Code: 2, Message: "search --id cannot be combined with --json"}
	}

	response, err := lookups.Search(a.Context, a.Client, lookups.SearchParams{
		Query:      strings.Join(c.Query, " "),
		ResultType: c.Type,
		Sport:      c.Sport,
		Page:       c.Page,
		Limit:      c.Limit,
		IDOnly:     c.ID,
	})
	if err != nil {
		return translateLookupError(err)
	}

	jsonOutput := c.JSON || (!a.stdoutIsTerminal() && !c.ID)
	if jsonOutput {
		return output.JSON(a.Stdout, response)
	}

	if c.ID {
		for _, id := range response.IDs {
			fmt.Fprintln(a.Stdout, id)
		}
		return nil
	}

	if len(response.Results) == 0 {
		fmt.Fprintln(a.Stdout, "No results.")
		return nil
	}

	tw := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tID\tSPORT\tCATEGORY\tNAME\tTEAM\tCOUNTRY")
	for _, result := range response.Results {
		fmt.Fprintf(
			tw,
			"%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			printable(result.Type),
			result.ID,
			printable(result.Sport),
			printable(result.Category),
			printable(result.Name),
			printable(result.Team),
			printable(result.Country),
		)
	}
	return tw.Flush()
}

func (c *ListCmd) Run(a *App) error {
	response, err := lookups.Sports(a.Context, a.Client, lookups.SportsParams{})
	if err != nil {
		return translateLookupError(err)
	}
	response.Sports = filterSupportedSports(response.Sports)

	if a.outputJSON(c.JSON) {
		return output.JSON(a.Stdout, response)
	}

	if len(response.Sports) == 0 {
		fmt.Fprintln(a.Stdout, "No sports.")
		return nil
	}

	tw := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SPORT\tLIVE\tTOTAL")
	for _, sport := range response.Sports {
		fmt.Fprintf(tw, "%s\t%d\t%d\n", sport.Slug, sport.Live, sport.Total)
	}
	return tw.Flush()
}

type sportBaseArgs struct {
	Date string `help:"UTC date in YYYY-MM-DD format."`
	JSON bool   `help:"Print JSON."`
}

type sportTournamentsArgs struct {
	Date string `help:"UTC date in YYYY-MM-DD format."`
	Page int    `default:"1" help:"1-based page number."`
	JSON bool   `help:"Print JSON."`
}

func (c *FootballCmd) Run(a *App) error   { return runSportCommand(a, "football", c.Args) }
func (c *BasketballCmd) Run(a *App) error { return runSportCommand(a, "basketball", c.Args) }
func (c *TennisCmd) Run(a *App) error     { return runSportCommand(a, "tennis", c.Args) }

type eventBaseArgs struct {
	JSON bool `help:"Print JSON."`
}

type eventSectionArgs struct {
	JSON  bool     `help:"Print JSON."`
	Names []string `arg:"" name:"section" required:"" help:"Section name."`
}

type eventWatchArgs struct {
	JSON        bool     `help:"Print NDJSON."`
	Section     []string `help:"Fetch and refresh an event section." xor:"event-sections"`
	AllSections bool     `name:"all-sections" help:"Fetch and refresh all discovered sections." xor:"event-sections"`
}

type eventTVChannelArgs struct {
	ChannelID int  `arg:"" name:"channel-id" required:"" help:"TV channel id."`
	JSON      bool `help:"Print JSON."`
}

type teamDirectionArgs struct {
	Limit int  `default:"10" help:"Limit results."`
	JSON  bool `help:"Print JSON."`
}

type teamBaseArgs struct {
	JSON bool `help:"Print JSON."`
}

type teamStandingsArgs struct {
	Season int  `help:"Use a specific season id."`
	JSON   bool `help:"Print JSON."`
}

type teamTournamentArgs struct {
	Tournament int  `required:"" help:"Use a specific unique tournament id."`
	Season     int  `required:"" help:"Use a specific season id."`
	JSON       bool `help:"Print JSON."`
}

type teamRankingsArgs struct {
	Tournament int  `help:"Use a specific unique tournament id."`
	Season     int  `help:"Use a specific season id."`
	JSON       bool `help:"Print JSON."`
}

type jsonOnlyArgs struct {
	JSON bool `help:"Print JSON."`
}

type playerDirectionArgs struct {
	Limit int  `default:"10" help:"Limit results."`
	JSON  bool `help:"Print JSON."`
}

type playerTournamentArgs struct {
	Tournament int    `required:"" help:"Use a specific unique tournament id."`
	Season     int    `required:"" help:"Use a specific season id."`
	Phase      string `help:"Use a specific phase name."`
	JSON       bool   `help:"Print JSON."`
}

type playerRequiredPhaseArgs struct {
	Tournament int    `required:"" help:"Use a specific unique tournament id."`
	Season     int    `required:"" help:"Use a specific season id."`
	Phase      string `required:"" help:"Use a specific phase name."`
	JSON       bool   `help:"Print JSON."`
}

type playerYearArgs struct {
	Year int  `required:"" help:"Use a specific calendar year."`
	JSON bool `help:"Print JSON."`
}

type tournamentBaseArgs struct {
	Season int  `help:"Use a specific season id."`
	JSON   bool `help:"Print JSON."`
}

type tournamentSectionArgs struct {
	Season int      `help:"Use a specific season id."`
	JSON   bool     `help:"Print JSON."`
	Names  []string `arg:"" name:"section" required:"" help:"Section name."`
}

type tournamentDirectionArgs struct {
	Season int  `help:"Use a specific season id."`
	Limit  int  `default:"10" help:"Limit results."`
	JSON   bool `help:"Print JSON."`
}

type tournamentRoundArgs struct {
	Season int    `help:"Use a specific season id."`
	JSON   bool   `help:"Print JSON."`
	Number int    `arg:"" name:"round" required:"" help:"Round number."`
	Slug   string `help:"Use the round slug endpoint."`
	Limit  int    `default:"10" help:"Limit results."`
}

type tournamentScheduledArgs struct {
	JSON  bool   `help:"Print JSON."`
	Date  string `help:"UTC date in YYYY-MM-DD format."`
	Limit int    `default:"10" help:"Limit results."`
}

func (c *EventCmd) Run(a *App) error {
	eventID, rest, err := parseIDAndRest(c.Args, "event")
	if err != nil {
		return err
	}

	if len(rest) == 0 || strings.HasPrefix(rest[0], "-") {
		flags := &eventBaseArgs{}
		handled, err := a.parseSubcommand("sports event <event-id>", rest, flags)
		if handled || err != nil {
			return err
		}
		return runEventBaseCommand(a, eventID, flags.JSON)
	}

	switch rest[0] {
	case "sections":
		flags := &eventBaseArgs{}
		handled, err := a.parseSubcommand("sports event <event-id> sections", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runEventSectionsCommand(a, eventID, flags.JSON)
	case "section":
		flags := &eventSectionArgs{}
		handled, err := a.parseSubcommand("sports event <event-id> section", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runEventSectionCommand(a, eventID, flags.Names, flags.JSON)
	case "watch":
		flags := &eventWatchArgs{}
		handled, err := a.parseSubcommand("sports event <event-id> watch", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runEventWatchCommand(a, eventID, flags.Section, flags.AllSections, flags.JSON)
	case "tv":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports event <event-id> tv", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runEventTVCommand(a, eventID)
	case "tv-channel":
		flags := &eventTVChannelArgs{}
		handled, err := a.parseSubcommand("sports event <event-id> tv-channel", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runEventTVChannelCommand(a, eventID, flags.ChannelID)
	case "h2h-events":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports event <event-id> h2h-events", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runEventH2HEventsCommand(a, eventID)
	default:
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown event subcommand %q", rest[0])}
	}
}

func (c *TeamCmd) Run(a *App) error {
	teamID, rest, err := parseIDAndRest(c.Args, "team")
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		return &exitError{Code: 2, Message: "missing team direction"}
	}

	flags := &teamDirectionArgs{}
	switch rest[0] {
	case "next":
		handled, err := a.parseSubcommand("sports team <team-id> next", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamDirectionCommand(a, teamID, flags.Limit, flags.JSON, true)
	case "last":
		handled, err := a.parseSubcommand("sports team <team-id> last", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamDirectionCommand(a, teamID, flags.Limit, flags.JSON, false)
	case "info":
		flags := &teamBaseArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> info", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamInfoCommand(a, teamID)
	case "tournaments":
		flags := &teamBaseArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> tournaments", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamTournamentsCommand(a, teamID)
	case "standings":
		flags := &teamStandingsArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> standings", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamStandingsCommand(a, teamID, flags.Season)
	case "stats":
		flags := &teamTournamentArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> stats", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamStatsCommand(a, teamID, flags.Tournament, flags.Season)
	case "players":
		flags := &teamBaseArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> players", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamPlayersCommand(a, teamID)
	case "media":
		flags := &teamBaseArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> media", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamMediaCommand(a, teamID)
	case "rankings":
		flags := &teamRankingsArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> rankings", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamRankingsCommand(a, teamID, flags.Tournament, flags.Season)
	case "top-players":
		flags := &teamTournamentArgs{}
		handled, err := a.parseSubcommand("sports team <team-id> top-players", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTeamTopPlayersCommand(a, teamID, flags.Tournament, flags.Season)
	default:
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown team subcommand %q", rest[0])}
	}
}

func (c *PlayerCmd) Run(a *App) error {
	playerID, rest, err := parseIDAndRest(c.Args, "player")
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		return &exitError{Code: 2, Message: "missing player subcommand"}
	}

	switch rest[0] {
	case "attributes":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> attributes", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerAttributesCommand(a, playerID)
	case "media":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> media", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerMediaCommand(a, playerID)
	case "media-videos":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> media-videos", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerMediaVideosCommand(a, playerID)
	case "last":
		flags := &playerDirectionArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> last", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerLastEventsCommand(a, playerID, flags.Limit)
	case "seasons":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> seasons", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerSeasonsCommand(a, playerID)
	case "career":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> career", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerCareerCommand(a, playerID)
	case "season-stats":
		flags := &playerTournamentArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> season-stats", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerSeasonStatsCommand(a, playerID, flags.Tournament, flags.Season, flags.Phase)
	case "season-ratings":
		flags := &playerTournamentArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> season-ratings", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerSeasonRatingsCommand(a, playerID, flags.Tournament, flags.Season, flags.Phase)
	case "characteristics":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> characteristics", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerCharacteristicsCommand(a, playerID)
	case "national-team":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> national-team", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerNationalTeamStatsCommand(a, playerID)
	case "tournaments":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> tournaments", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerTournamentsCommand(a, playerID)
	case "season-heatmap":
		flags := &playerRequiredPhaseArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> season-heatmap", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerSeasonHeatmapCommand(a, playerID, flags.Tournament, flags.Season, flags.Phase)
	case "penalty-history":
		flags := &playerTournamentArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> penalty-history", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerPenaltyHistoryCommand(a, playerID, flags.Tournament, flags.Season)
	case "shot-actions":
		flags := &playerRequiredPhaseArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> shot-actions", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerShotActionsCommand(a, playerID, flags.Tournament, flags.Season, flags.Phase)
	case "shot-action-areas":
		flags := &playerRequiredPhaseArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> shot-action-areas", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerShotActionAreasCommand(a, playerID, flags.Tournament, flags.Season, flags.Phase)
	case "year-stats":
		flags := &playerYearArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> year-stats", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerYearStatsCommand(a, playerID, flags.Year)
	case "featured-event":
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports player <player-id> featured-event", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runPlayerFeaturedEventCommand(a, playerID)
	default:
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown player subcommand %q", rest[0])}
	}
}

func (c *TournamentCmd) Run(a *App) error {
	tournamentID, rest, err := parseIDAndRest(c.Args, "tournament")
	if err != nil {
		return err
	}

	if len(rest) == 0 || strings.HasPrefix(rest[0], "-") {
		flags := &tournamentBaseArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id>", rest, flags)
		if handled || err != nil {
			return err
		}
		return runTournamentBaseCommand(a, tournamentID, flags.Season, flags.JSON)
	}

	switch rest[0] {
	case "sections":
		flags := &tournamentBaseArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id> sections", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTournamentSectionsCommand(a, tournamentID, flags.Season, flags.JSON)
	case "section":
		flags := &tournamentSectionArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id> section", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTournamentSectionCommand(a, tournamentID, flags.Season, flags.Names, flags.JSON)
	case "seasons":
		flags := &eventBaseArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id> seasons", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTournamentSeasonsCommand(a, tournamentID, flags.JSON)
	case "next":
		flags := &tournamentDirectionArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id> next", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTournamentEventsCommand(a, tournamentID, flags.Season, flags.JSON, true, false, 0, "", flags.Limit)
	case "last":
		flags := &tournamentDirectionArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id> last", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTournamentEventsCommand(a, tournamentID, flags.Season, flags.JSON, false, true, 0, "", flags.Limit)
	case "round":
		flags := &tournamentRoundArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id> round", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTournamentEventsCommand(a, tournamentID, flags.Season, flags.JSON, false, false, flags.Number, flags.Slug, flags.Limit)
	case "scheduled":
		flags := &tournamentScheduledArgs{}
		handled, err := a.parseSubcommand("sports tournament <tournament-id> scheduled", rest[1:], flags)
		if handled || err != nil {
			return err
		}
		return runTournamentScheduledCommand(a, tournamentID, flags.Date, flags.Limit, flags.JSON)
	default:
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown tournament subcommand %q", rest[0])}
	}
}

func (c *TrendingCmd) Run(a *App) error {
	response, err := lookups.Trending(a.Context, a.Client, lookups.TrendingParams{
		Country: c.Country,
		Date:    c.Date,
		Limit:   c.Limit,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(c.JSON) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "COUNTRY: %s\n", response.Country)
	fmt.Fprintf(a.Stdout, "DATE: %s\n", response.Date)
	if len(response.Events) == 0 {
		fmt.Fprintln(a.Stdout, "No events.")
		return nil
	}

	tw := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RANK\tEVENT ID\tSTART\tSPORT\tSTATUS\tMATCH\tTOURNAMENT")
	for _, event := range response.Events {
		fmt.Fprintf(
			tw,
			"%d\t%d\t%s\t%s\t%s\t%s vs %s\t%s\n",
			event.Rank,
			event.EventID,
			event.StartTime.Format(time.RFC3339),
			printable(event.Sport),
			printable(coalesce(event.StatusDescription, event.StatusType)),
			printable(event.Home),
			printable(event.Away),
			printable(event.Tournament),
		)
	}
	return tw.Flush()
}

func (a *App) parseSubcommand(name string, args []string, grammar any) (handled bool, err error) {
	parser, err := kong.New(
		grammar,
		kong.Name(name),
		kong.Writers(a.Stdout, a.Stderr),
		kong.Exit(func(code int) {
			panic(parserExit{Code: code})
		}),
	)
	if err != nil {
		return false, err
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			if exited, ok := recovered.(parserExit); ok {
				if exited.Code == 0 {
					handled = true
					err = nil
					return
				}
				err = &exitError{Code: exited.Code}
				return
			}
			panic(recovered)
		}
	}()

	if _, err := parser.Parse(args); err != nil {
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			return false, &exitError{Code: 2, Message: err.Error()}
		}
		return false, err
	}
	return false, nil
}

func parseIDAndRest(args []string, subject string) (int, []string, error) {
	if len(args) == 0 {
		return 0, nil, &exitError{Code: 2, Message: fmt.Sprintf("%s id is required", subject)}
	}
	id, err := strconv.Atoi(args[0])
	if err != nil || id <= 0 {
		return 0, nil, &exitError{Code: 2, Message: fmt.Sprintf("%s id must be a positive integer", subject)}
	}
	return id, args[1:], nil
}

func runSportCommand(a *App, sport string, args []string) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		flags := &sportBaseArgs{}
		handled, err := a.parseSubcommand("sports "+sport, args, flags)
		if handled || err != nil {
			return err
		}
		return runSportEventsCommand(a, sport, flags.Date, flags.JSON)
	}

	switch args[0] {
	case "tournaments":
		flags := &sportTournamentsArgs{}
		handled, err := a.parseSubcommand("sports "+sport+" tournaments", args[1:], flags)
		if handled || err != nil {
			return err
		}
		return runSportTournamentsCommand(a, sport, flags.Date, flags.JSON, flags.Page)
	case "sections":
		flags := &sportBaseArgs{}
		handled, err := a.parseSubcommand("sports "+sport+" sections", args[1:], flags)
		if handled || err != nil {
			return err
		}
		return runSportSectionsCommand(a, sport, flags.Date, flags.JSON)
	case "watch":
		flags := &sportBaseArgs{}
		handled, err := a.parseSubcommand("sports "+sport+" watch", args[1:], flags)
		if handled || err != nil {
			return err
		}
		return runSportWatchCommand(a, sport, flags.Date, flags.JSON)
	case "live-tournaments":
		if sport == "tennis" {
			return &exitError{Code: 2, Message: fmt.Sprintf("unknown %s subcommand %q", sport, args[0])}
		}
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports "+sport+" live-tournaments", args[1:], flags)
		if handled || err != nil {
			return err
		}
		return runSportLiveTournamentsCommand(a, sport)
	case "categories":
		if sport == "tennis" {
			return &exitError{Code: 2, Message: fmt.Sprintf("unknown %s subcommand %q", sport, args[0])}
		}
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports "+sport+" categories", args[1:], flags)
		if handled || err != nil {
			return err
		}
		return runSportCategoriesCommand(a, sport)
	case "top-players":
		if sport == "tennis" {
			return &exitError{Code: 2, Message: fmt.Sprintf("unknown %s subcommand %q", sport, args[0])}
		}
		flags := &jsonOnlyArgs{}
		handled, err := a.parseSubcommand("sports "+sport+" top-players", args[1:], flags)
		if handled || err != nil {
			return err
		}
		return runSportTopPlayersCommand(a, sport)
	default:
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown %s subcommand %q", sport, args[0])}
	}
}

func runEventBaseCommand(a *App, eventID int, jsonFlag bool) error {
	response, err := lookups.Event(a.Context, a.Client, lookups.EventParams{
		EventID:                   eventID,
		AllowPartialSectionErrors: a.outputJSON(jsonFlag),
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, eventJSONResponse(response))
	}

	fmt.Fprintf(a.Stdout, "EVENT ID: %d\n", response.EventID)
	fmt.Fprintf(a.Stdout, "SPORT: %s\n", printable(response.Sport))
	fmt.Fprintf(a.Stdout, "MATCH: %s vs %s\n", printable(response.Home), printable(response.Away))
	fmt.Fprintf(a.Stdout, "STATUS: %s\n", printable(coalesce(response.StatusDescription, response.StatusType)))
	fmt.Fprintf(a.Stdout, "START: %s\n", response.StartTime)
	fmt.Fprintf(a.Stdout, "SCORE: %s\n", printable(formatScore(response.HomeScore, response.AwayScore)))
	fmt.Fprintf(a.Stdout, "TOURNAMENT: %s\n", printable(response.Tournament))
	if strings.TrimSpace(response.Venue) != "" {
		fmt.Fprintf(a.Stdout, "VENUE: %s\n", response.Venue)
	}
	fmt.Fprintf(a.Stdout, "AVAILABLE SECTIONS: %s\n", printable(strings.Join(response.AvailableSections, ", ")))
	return nil
}

func runEventSectionsCommand(a *App, eventID int, jsonFlag bool) error {
	response, err := lookups.Event(a.Context, a.Client, lookups.EventParams{
		EventID:      eventID,
		SectionsOnly: true,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "EVENT ID: %d\n", eventID)
	fmt.Fprintf(a.Stdout, "SPORT: %s\n", printable(response.Sport))
	fmt.Fprintln(a.Stdout, "SECTIONS:")
	for _, section := range response.AvailableSections {
		fmt.Fprintf(a.Stdout, "- %s\n", printable(section))
	}
	fmt.Fprintln(a.Stdout)
	fmt.Fprintln(a.Stdout, "EXAMPLES:")
	fmt.Fprintf(a.Stdout, "  sports event %d section h2h --json\n", eventID)
	fmt.Fprintf(a.Stdout, "  sports event %d section lineups --json\n", eventID)
	return nil
}

func runEventSectionCommand(a *App, eventID int, names []string, jsonFlag bool) error {
	response, err := lookups.Event(a.Context, a.Client, lookups.EventParams{
		EventID:                   eventID,
		Sections:                  names,
		AllowPartialSectionErrors: a.outputJSON(jsonFlag),
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, eventJSONResponse(response))
	}

	fmt.Fprintf(a.Stdout, "EVENT ID: %d\n", response.EventID)
	fmt.Fprintf(a.Stdout, "SPORT: %s\n", printable(response.Sport))
	fmt.Fprintf(a.Stdout, "MATCH: %s vs %s\n", printable(response.Home), printable(response.Away))
	fmt.Fprintf(a.Stdout, "STATUS: %s\n", printable(coalesce(response.StatusDescription, response.StatusType)))
	fmt.Fprintf(a.Stdout, "START: %s\n", response.StartTime)
	fmt.Fprintf(a.Stdout, "SCORE: %s\n", printable(formatScore(response.HomeScore, response.AwayScore)))
	fmt.Fprintf(a.Stdout, "TOURNAMENT: %s\n", printable(response.Tournament))
	fmt.Fprintf(a.Stdout, "AVAILABLE SECTIONS: %s\n", printable(strings.Join(response.AvailableSections, ", ")))
	return nil
}

func runEventWatchCommand(a *App, eventID int, sections []string, allSections, jsonFlag bool) error {
	jsonOutput := a.outputJSON(jsonFlag)
	return a.Client.WatchEvents(a.Context, []int{eventID}, sections, allSections, func(record sofascoreapi.WatchRecord) error {
		return a.emitWatchRecord(record, jsonOutput)
	})
}

func runEventTVCommand(a *App, eventID int) error {
	response, err := lookups.EventTV(a.Context, a.Client, lookups.EventTVParams{
		EventID: eventID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runEventTVChannelCommand(a *App, eventID, channelID int) error {
	response, err := lookups.EventTVChannel(a.Context, a.Client, lookups.EventTVChannelParams{
		EventID:   eventID,
		ChannelID: channelID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runEventH2HEventsCommand(a *App, eventID int) error {
	response, err := lookups.EventH2HEvents(a.Context, a.Client, lookups.EventH2HEventsParams{
		EventID: eventID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTournamentBaseCommand(a *App, tournamentID, seasonID int, jsonFlag bool) error {
	response, err := lookups.Tournament(a.Context, a.Client, lookups.TournamentParams{
		TournamentID:              tournamentID,
		SeasonID:                  seasonID,
		AllowPartialSectionErrors: a.outputJSON(jsonFlag),
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, tournamentJSONResponse(response))
	}

	fmt.Fprintf(a.Stdout, "TOURNAMENT ID: %d\n", response.TournamentID)
	fmt.Fprintf(a.Stdout, "NAME: %s\n", printable(response.Name))
	fmt.Fprintf(a.Stdout, "SPORT: %s\n", printable(response.Sport))
	fmt.Fprintf(a.Stdout, "CATEGORY: %s\n", printable(response.Category))
	fmt.Fprintf(a.Stdout, "COUNTRY: %s\n", printable(response.Country))
	fmt.Fprintf(a.Stdout, "SEASON ID: %d\n", response.SeasonID)
	fmt.Fprintf(a.Stdout, "SEASON: %s\n", printable(response.SeasonName))
	fmt.Fprintf(a.Stdout, "AVAILABLE SECTIONS: %s\n", printable(strings.Join(response.AvailableSections, ", ")))
	return nil
}

func runTournamentSectionsCommand(a *App, tournamentID, seasonID int, jsonFlag bool) error {
	response, err := lookups.Tournament(a.Context, a.Client, lookups.TournamentParams{
		TournamentID: tournamentID,
		SeasonID:     seasonID,
		SectionsOnly: true,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "TOURNAMENT ID: %d\n", tournamentID)
	fmt.Fprintf(a.Stdout, "NAME: %s\n", printable(response.Name))
	fmt.Fprintf(a.Stdout, "SPORT: %s\n", printable(response.Sport))
	fmt.Fprintf(a.Stdout, "SEASON ID: %d\n", response.SeasonID)
	fmt.Fprintf(a.Stdout, "SEASON: %s\n", printable(response.SeasonName))
	fmt.Fprintln(a.Stdout, "SECTIONS:")
	for _, section := range response.AvailableSections {
		fmt.Fprintf(a.Stdout, "- %s\n", printable(section))
	}
	fmt.Fprintln(a.Stdout)
	fmt.Fprintln(a.Stdout, "EXAMPLES:")
	fmt.Fprintf(a.Stdout, "  sports tournament %d section standings/total --json\n", tournamentID)
	fmt.Fprintf(a.Stdout, "  sports tournament %d section info --season %d --json\n", tournamentID, response.SeasonID)
	return nil
}

func runTournamentSectionCommand(a *App, tournamentID, seasonID int, names []string, jsonFlag bool) error {
	response, err := lookups.Tournament(a.Context, a.Client, lookups.TournamentParams{
		TournamentID:              tournamentID,
		SeasonID:                  seasonID,
		Sections:                  names,
		AllowPartialSectionErrors: a.outputJSON(jsonFlag),
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, tournamentJSONResponse(response))
	}

	fmt.Fprintf(a.Stdout, "TOURNAMENT ID: %d\n", response.TournamentID)
	fmt.Fprintf(a.Stdout, "NAME: %s\n", printable(response.Name))
	fmt.Fprintf(a.Stdout, "SPORT: %s\n", printable(response.Sport))
	fmt.Fprintf(a.Stdout, "CATEGORY: %s\n", printable(response.Category))
	fmt.Fprintf(a.Stdout, "COUNTRY: %s\n", printable(response.Country))
	fmt.Fprintf(a.Stdout, "SEASON ID: %d\n", response.SeasonID)
	fmt.Fprintf(a.Stdout, "SEASON: %s\n", printable(response.SeasonName))
	fmt.Fprintf(a.Stdout, "AVAILABLE SECTIONS: %s\n", printable(strings.Join(response.AvailableSections, ", ")))
	return nil
}

func runTournamentSeasonsCommand(a *App, tournamentID int, jsonFlag bool) error {
	response, err := lookups.TournamentSeasons(a.Context, a.Client, lookups.TournamentSeasonsParams{
		TournamentID: tournamentID,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	if len(response.Seasons) == 0 {
		fmt.Fprintln(a.Stdout, "No seasons.")
		return nil
	}

	tw := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SEASON ID\tNAME\tYEAR")
	for _, season := range response.Seasons {
		fmt.Fprintf(tw, "%d\t%s\t%s\n", season.ID, printable(season.Name), printable(season.Year))
	}
	return tw.Flush()
}

func runTournamentScheduledCommand(a *App, tournamentID int, date string, limit int, jsonFlag bool) error {
	response, err := lookups.TournamentScheduledEvents(a.Context, a.Client, lookups.TournamentScheduledEventsParams{
		TournamentID: tournamentID,
		Date:         date,
		Limit:        limit,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "TOURNAMENT ID: %d\n", response.TournamentID)
	fmt.Fprintf(a.Stdout, "DATE: %s\n", response.Date)
	if len(response.Events) == 0 {
		fmt.Fprintln(a.Stdout, "No events.")
		return nil
	}
	return writeEventSummaryTable(a.Stdout, response.Events)
}

func runSportEventsCommand(a *App, sport, date string, jsonFlag bool) error {
	response, err := lookups.SportEvents(a.Context, a.Client, lookups.SportEventsParams{
		Sport: sport,
		Date:  date,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "SPORT: %s\n", response.Sport)
	fmt.Fprintf(a.Stdout, "DATE: %s\n", response.Date)
	if len(response.Events) == 0 {
		fmt.Fprintln(a.Stdout, "No events.")
		return nil
	}
	return writeEventSummaryTable(a.Stdout, response.Events)
}

func runSportTournamentsCommand(a *App, sport, date string, jsonFlag bool, page int) error {
	response, err := lookups.SportScheduledTournaments(a.Context, a.Client, lookups.SportScheduledTournamentsParams{
		Sport: sport,
		Date:  date,
		Page:  page,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "SPORT: %s\n", response.Sport)
	fmt.Fprintf(a.Stdout, "DATE: %s\n", response.Date)
	fmt.Fprintf(a.Stdout, "PAGE: %d\n", response.Page)
	fmt.Fprintf(a.Stdout, "HAS NEXT PAGE: %t\n", response.HasNextPage)
	if len(response.Tournaments) == 0 {
		fmt.Fprintln(a.Stdout, "No tournaments.")
		return nil
	}

	tw := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "UNIQUE TOURNAMENT ID\tTOURNAMENT\tSTAGE\tCATEGORY\tUTC EVENTS")
	for _, tournament := range response.Tournaments {
		fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%d\n",
			tournament.UniqueTournamentID,
			printable(tournament.UniqueTournament),
			printable(tournament.Name),
			printable(tournament.Category),
			tournament.UTCEventCount,
		)
	}
	return tw.Flush()
}

func runSportSectionsCommand(a *App, sport, date string, jsonFlag bool) error {
	if strings.TrimSpace(date) != "" {
		return &exitError{Code: 2, Message: "sections does not accept --date"}
	}

	response, err := lookups.Sports(a.Context, a.Client, lookups.SportsParams{
		Sport:    sport,
		Sections: true,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "SPORT: %s\n", response.Sport)
	fmt.Fprintf(a.Stdout, "SAMPLE EVENT ID: %d\n", response.SampleEventID)
	fmt.Fprintf(a.Stdout, "SECTIONS: %s\n", printable(strings.Join(response.Sections, ", ")))
	return nil
}

func runSportWatchCommand(a *App, sport, date string, jsonFlag bool) error {
	if strings.TrimSpace(date) != "" {
		return &exitError{Code: 2, Message: "watch does not accept --date"}
	}

	jsonOutput := a.outputJSON(jsonFlag)
	return a.Client.WatchSports(a.Context, []string{sport}, func(record sofascoreapi.WatchRecord) error {
		return a.emitWatchRecord(record, jsonOutput)
	})
}

func runSportLiveTournamentsCommand(a *App, sport string) error {
	response, err := lookups.SportLiveTournaments(a.Context, a.Client, lookups.SportLiveTournamentsParams{
		Sport: sport,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runSportCategoriesCommand(a *App, sport string) error {
	response, err := lookups.SportCategories(a.Context, a.Client, lookups.SportCategoriesParams{
		Sport: sport,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runSportTopPlayersCommand(a *App, sport string) error {
	response, err := lookups.SportTopPlayers(a.Context, a.Client, lookups.SportTopPlayersParams{
		Sport: sport,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamDirectionCommand(a *App, teamID, limit int, jsonFlag, next bool) error {
	response, err := lookups.TeamEvents(a.Context, a.Client, lookups.TeamEventsParams{
		TeamID: teamID,
		Next:   next,
		Last:   !next,
		Limit:  limit,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	if len(response.Events) == 0 {
		fmt.Fprintln(a.Stdout, "No events.")
		return nil
	}

	return writeEventSummaryTable(a.Stdout, response.Events)
}

func runTeamInfoCommand(a *App, teamID int) error {
	response, err := lookups.TeamInfo(a.Context, a.Client, lookups.TeamInfoParams{
		TeamID: teamID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamTournamentsCommand(a *App, teamID int) error {
	response, err := lookups.TeamTournaments(a.Context, a.Client, lookups.TeamTournamentsParams{
		TeamID: teamID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamStandingsCommand(a *App, teamID, seasonID int) error {
	response, err := lookups.TeamStandings(a.Context, a.Client, lookups.TeamStandingsParams{
		TeamID:   teamID,
		SeasonID: seasonID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamStatsCommand(a *App, teamID, tournamentID, seasonID int) error {
	response, err := lookups.TeamStats(a.Context, a.Client, lookups.TeamStatsParams{
		TeamID:       teamID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamPlayersCommand(a *App, teamID int) error {
	response, err := lookups.TeamPlayers(a.Context, a.Client, lookups.TeamPlayersParams{
		TeamID: teamID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamMediaCommand(a *App, teamID int) error {
	response, err := lookups.TeamMedia(a.Context, a.Client, lookups.TeamMediaParams{
		TeamID: teamID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamRankingsCommand(a *App, teamID, tournamentID, seasonID int) error {
	response, err := lookups.TeamRankings(a.Context, a.Client, lookups.TeamRankingsParams{
		TeamID:       teamID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTeamTopPlayersCommand(a *App, teamID, tournamentID, seasonID int) error {
	response, err := lookups.TeamTopPlayers(a.Context, a.Client, lookups.TeamTopPlayersParams{
		TeamID:       teamID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerAttributesCommand(a *App, playerID int) error {
	response, err := lookups.PlayerAttributes(a.Context, a.Client, lookups.PlayerAttributesParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerMediaCommand(a *App, playerID int) error {
	response, err := lookups.PlayerMedia(a.Context, a.Client, lookups.PlayerMediaParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerMediaVideosCommand(a *App, playerID int) error {
	response, err := lookups.PlayerMediaVideos(a.Context, a.Client, lookups.PlayerMediaVideosParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerLastEventsCommand(a *App, playerID, limit int) error {
	response, err := lookups.PlayerLastEvents(a.Context, a.Client, lookups.PlayerLastEventsParams{
		PlayerID: playerID,
		Limit:    limit,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerSeasonsCommand(a *App, playerID int) error {
	response, err := lookups.PlayerSeasons(a.Context, a.Client, lookups.PlayerSeasonsParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerCareerCommand(a *App, playerID int) error {
	response, err := lookups.PlayerCareer(a.Context, a.Client, lookups.PlayerCareerParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerSeasonStatsCommand(a *App, playerID, tournamentID, seasonID int, phase string) error {
	response, err := lookups.PlayerSeasonStats(a.Context, a.Client, lookups.PlayerSeasonStatsParams{
		PlayerID:     playerID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
		Phase:        phase,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerSeasonRatingsCommand(a *App, playerID, tournamentID, seasonID int, phase string) error {
	response, err := lookups.PlayerSeasonRatings(a.Context, a.Client, lookups.PlayerSeasonRatingsParams{
		PlayerID:     playerID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
		Phase:        phase,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerCharacteristicsCommand(a *App, playerID int) error {
	response, err := lookups.PlayerCharacteristics(a.Context, a.Client, lookups.PlayerCharacteristicsParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerNationalTeamStatsCommand(a *App, playerID int) error {
	response, err := lookups.PlayerNationalTeamStats(a.Context, a.Client, lookups.PlayerNationalTeamStatsParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerTournamentsCommand(a *App, playerID int) error {
	response, err := lookups.PlayerTournaments(a.Context, a.Client, lookups.PlayerTournamentsParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerSeasonHeatmapCommand(a *App, playerID, tournamentID, seasonID int, phase string) error {
	response, err := lookups.PlayerSeasonHeatmap(a.Context, a.Client, lookups.PlayerSeasonHeatmapParams{
		PlayerID:     playerID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
		Phase:        phase,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerPenaltyHistoryCommand(a *App, playerID, tournamentID, seasonID int) error {
	response, err := lookups.PlayerPenaltyHistory(a.Context, a.Client, lookups.PlayerPenaltyHistoryParams{
		PlayerID:     playerID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerShotActionsCommand(a *App, playerID, tournamentID, seasonID int, phase string) error {
	response, err := lookups.PlayerShotActions(a.Context, a.Client, lookups.PlayerShotActionsParams{
		PlayerID:     playerID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
		Phase:        phase,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerShotActionAreasCommand(a *App, playerID, tournamentID, seasonID int, phase string) error {
	response, err := lookups.PlayerShotActionAreas(a.Context, a.Client, lookups.PlayerShotActionAreasParams{
		PlayerID:     playerID,
		TournamentID: tournamentID,
		SeasonID:     seasonID,
		Phase:        phase,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerYearStatsCommand(a *App, playerID, year int) error {
	response, err := lookups.PlayerYearStats(a.Context, a.Client, lookups.PlayerYearStatsParams{
		PlayerID: playerID,
		Year:     year,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runPlayerFeaturedEventCommand(a *App, playerID int) error {
	response, err := lookups.PlayerFeaturedEvent(a.Context, a.Client, lookups.PlayerFeaturedEventParams{
		PlayerID: playerID,
	})
	if err != nil {
		return translateLookupError(err)
	}
	return writeStructuredOutput(a.Stdout, response)
}

func runTournamentEventsCommand(a *App, tournamentID, seasonID int, jsonFlag, next, last bool, round int, slug string, limit int) error {
	response, err := lookups.TournamentEvents(a.Context, a.Client, lookups.TournamentEventsParams{
		TournamentID: tournamentID,
		SeasonID:     seasonID,
		Next:         next,
		Last:         last,
		Round:        round,
		Slug:         slug,
		Limit:        limit,
	})
	if err != nil {
		return translateLookupError(err)
	}

	if a.outputJSON(jsonFlag) {
		return output.JSON(a.Stdout, response)
	}

	fmt.Fprintf(a.Stdout, "TOURNAMENT ID: %d\n", tournamentID)
	fmt.Fprintf(a.Stdout, "SEASON ID: %d\n", response.SeasonID)
	fmt.Fprintf(a.Stdout, "SEASON: %s\n", printable(response.SeasonName))
	if response.Mode == "round" {
		fmt.Fprintf(a.Stdout, "ROUND: %d\n", response.Round)
		if strings.TrimSpace(response.Slug) != "" {
			fmt.Fprintf(a.Stdout, "SLUG: %s\n", strings.TrimSpace(response.Slug))
		}
	} else {
		fmt.Fprintf(a.Stdout, "MODE: %s\n", response.Mode)
	}
	if len(response.Events) == 0 {
		fmt.Fprintln(a.Stdout, "No events.")
		return nil
	}

	return writeEventSummaryTable(a.Stdout, response.Events)
}

func eventJSONResponse(response lookups.EventResponse) any {
	return struct {
		OK                bool              `json:"ok"`
		Partial           bool              `json:"partial"`
		EventID           int               `json:"event_id"`
		Sport             string            `json:"sport"`
		AvailableSections []string          `json:"available_sections"`
		Event             any               `json:"event"`
		Sections          map[string]any    `json:"sections"`
		SectionErrors     map[string]string `json:"section_errors,omitempty"`
	}{
		OK:                response.OK,
		Partial:           response.Partial,
		EventID:           response.EventID,
		Sport:             response.Sport,
		AvailableSections: response.AvailableSections,
		Event:             response.Event,
		Sections:          emptyMap(response.Sections),
		SectionErrors:     response.SectionErrors,
	}
}

func tournamentJSONResponse(response lookups.TournamentResponse) any {
	return struct {
		OK                bool                            `json:"ok"`
		Partial           bool                            `json:"partial"`
		TournamentID      int                             `json:"tournament_id"`
		Sport             string                          `json:"sport"`
		Category          string                          `json:"category,omitempty"`
		Country           string                          `json:"country,omitempty"`
		SeasonID          int                             `json:"season_id"`
		SeasonName        string                          `json:"season_name,omitempty"`
		AvailableSections []string                        `json:"available_sections"`
		Seasons           []sofascoreapi.TournamentSeason `json:"seasons,omitempty"`
		Tournament        any                             `json:"tournament"`
		Sections          map[string]any                  `json:"sections"`
		SectionErrors     map[string]string               `json:"section_errors,omitempty"`
	}{
		OK:                response.OK,
		Partial:           response.Partial,
		TournamentID:      response.TournamentID,
		Sport:             response.Sport,
		Category:          response.Category,
		Country:           response.Country,
		SeasonID:          response.SeasonID,
		SeasonName:        response.SeasonName,
		AvailableSections: response.AvailableSections,
		Seasons:           response.Seasons,
		Tournament:        response.Tournament,
		Sections:          emptyMap(response.Sections),
		SectionErrors:     response.SectionErrors,
	}
}

func writeStructuredOutput(w io.Writer, value any) error {
	return output.JSON(w, value)
}

func filterSupportedSports(sports []sofascoreapi.SportCount) []sofascoreapi.SportCount {
	filtered := make([]sofascoreapi.SportCount, 0, len(sports))
	for _, sport := range sports {
		if _, ok := supportedSportSlugs[sport.Slug]; ok {
			filtered = append(filtered, sport)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Slug < filtered[j].Slug
	})
	return filtered
}

func translateLookupError(err error) error {
	var lookupErr *lookups.Error
	if !errors.As(err, &lookupErr) {
		return err
	}

	code := 3
	switch lookupErr.Kind {
	case lookups.ErrorKindInvalid:
		code = 2
	case lookups.ErrorKindNotFound:
		code = 4
	case lookups.ErrorKindUpstream:
		code = 3
	}

	return &exitError{
		Code:    code,
		Message: lookupErr.Message,
	}
}

func emptyMap[V any](value map[string]V) map[string]V {
	if value != nil {
		return value
	}
	return map[string]V{}
}

func printable(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func coalesce(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func formatScore(home, away *int) string {
	if home == nil || away == nil {
		return ""
	}
	return strconv.Itoa(*home) + "-" + strconv.Itoa(*away)
}

func writeEventSummaryTable(w io.Writer, events []sofascoreapi.EventSummary) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "EVENT ID\tSTART\tSTATUS\tMATCH\tTOURNAMENT")
	for _, event := range events {
		fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s vs %s\t%s\n",
			event.EventID,
			event.StartTime.Format(time.RFC3339),
			printable(coalesce(event.StatusDescription, event.StatusType)),
			printable(event.Home),
			printable(event.Away),
			printable(event.Tournament),
		)
	}
	return tw.Flush()
}
