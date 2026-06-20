package wandb

import (
	"context"
	"errors"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes W&B as a kit Domain: a driver that a multi-domain host
// (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/wandb-cli/wandb"
func init() { kit.Register(Domain{}) }

// Domain is the W&B driver. It carries no state; the per-run client is built
// by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, hostnames, and identity.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme:  "wandb",
		Aliases: []string{"wb"},
		Hosts:   []string{Host, "www.wandb.ai", "api.wandb.ai"},
		Identity: kit.Identity{
			Binary: "wandb",
			Short:  "Browse Weights and Biases public reports and runs",
			Long: `wandb turns wandb.ai into a fast, scriptable command line.

Read public experiment reports, search the gallery, and look up entity profiles
via the public GraphQL API at api.wandb.ai - no API key required for public data.

Quick start:
  wandb reports                         20 most recent public workspace views
  wandb reports -n 50                   50 views
  wandb trending                        the pinned gallery report
  wandb search "image classification"   search public reports
  wandb entity stacey                   stacey's public profile
  wandb project wandb/wandb             the wandb/wandb project`,
			Site: Host,
			Repo: "https://github.com/tamnd/wandb-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "reports",
		Group:   "list",
		Summary: "List recent public W&B workspace views",
	}, listReports)

	kit.Handle(app, kit.OpMeta{
		Name:     "report",
		Group:    "read",
		Single:   true,
		Resolver: true,
		URIType:  "report",
		Summary:  "Fetch a public view by ID",
		Args:     []kit.Arg{{Name: "id", Help: "view ID or wandb.ai URL"}},
	}, getReport)

	kit.Handle(app, kit.OpMeta{
		Name:    "trending",
		Group:   "list",
		Summary: "Show the featured report from the W&B gallery",
	}, trendingReport)

	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "read",
		Summary: "Search public reports by keyword",
		Args:    []kit.Arg{{Name: "query", Help: "search terms", Variadic: true}},
	}, searchReports)

	kit.Handle(app, kit.OpMeta{
		Name:     "entity",
		Group:    "read",
		Single:   true,
		Resolver: true,
		URIType:  "entity",
		Summary:  "Fetch an entity (user or team) profile",
		Args:     []kit.Arg{{Name: "name", Help: "entity name or wandb.ai URL"}},
	}, getEntity)

	kit.Handle(app, kit.OpMeta{
		Name:     "project",
		Group:    "read",
		Single:   true,
		Resolver: true,
		URIType:  "project",
		Summary:  "Fetch a public project",
		Args:     []kit.Arg{{Name: "ref", Help: "entity/project or wandb.ai URL"}},
	}, getProject)
}

// newClient builds the Client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- input structs ---

type reportsInput struct {
	Type   string  `kit:"flag" help:"view type (runs)" default:"runs"`
	Limit  int     `kit:"flag,inherit" help:"max results" default:"20"`
	Client *Client `kit:"inject"`
}

type reportInput struct {
	ID     string  `kit:"arg" help:"view ID or URL"`
	Client *Client `kit:"inject"`
}

type searchInput struct {
	Query  []string `kit:"arg,variadic" help:"search terms"`
	Limit  int      `kit:"flag,inherit" help:"max results" default:"20"`
	Client *Client  `kit:"inject"`
}

type entityInput struct {
	Name   string  `kit:"arg" help:"entity name or URL"`
	Client *Client `kit:"inject"`
}

type projectInput struct {
	Ref    string  `kit:"arg" help:"entity/project or URL"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listReports(ctx context.Context, in reportsInput, emit func(Report) error) error {
	reports, _, _, err := in.Client.ListReports(ctx, in.Type, in.Limit, "")
	if err != nil {
		return mapErr(err)
	}
	for _, r := range reports {
		if err := emit(r); err != nil {
			return err
		}
	}
	return nil
}

func getReport(ctx context.Context, in reportInput, emit func(*Report) error) error {
	id := parseReportID(in.ID)
	r, err := in.Client.GetReport(ctx, id)
	if err != nil {
		return mapErr(err)
	}
	return emit(r)
}

func trendingReport(ctx context.Context, in struct{ Client *Client `kit:"inject"` }, emit func(Report) error) error {
	r, err := in.Client.FeaturedReport(ctx)
	if err != nil {
		return mapErr(err)
	}
	return emit(*r)
}

func searchReports(ctx context.Context, in searchInput, emit func(Report) error) error {
	query := strings.Join(in.Query, " ")
	if query == "" {
		return errs.Usage("wandb search: query required")
	}
	results, err := in.Client.SearchReports(ctx, query)
	if err != nil {
		return mapErr(err)
	}
	n := 0
	for _, r := range results {
		if in.Limit > 0 && n >= in.Limit {
			break
		}
		if err := emit(r); err != nil {
			return err
		}
		n++
	}
	return nil
}

func getEntity(ctx context.Context, in entityInput, emit func(*Entity) error) error {
	name := parseEntityName(in.Name)
	e, err := in.Client.GetEntity(ctx, name)
	if err != nil {
		return mapErr(err)
	}
	return emit(e)
}

func getProject(ctx context.Context, in projectInput, emit func(*Project) error) error {
	entity, project := parseProjectRef(in.Ref)
	if entity == "" || project == "" {
		return errs.Usage("wandb project: expected entity/project, got %q", in.Ref)
	}
	p, err := in.Client.GetProject(ctx, entity, project)
	if err != nil {
		return mapErr(err)
	}
	return emit(p)
}

// --- Resolver ---

// Classify turns any accepted input into the canonical (uriType, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("wandb: empty input")
	}

	// wandb.ai URLs
	if strings.Contains(input, "wandb.ai/") {
		// Report: wandb.ai/<entity>/<project>/reports/<name>--<id>
		if strings.Contains(input, "/reports/") {
			parts := strings.SplitN(input, "/reports/", 2)
			if len(parts) == 2 {
				slug := strings.Split(parts[1], "?")[0]
				idx := strings.LastIndex(slug, "--")
				if idx >= 0 {
					return "report", slug[idx+2:], nil
				}
				return "report", slug, nil
			}
		}
		after := strings.SplitN(input, "wandb.ai/", 2)
		if len(after) == 2 {
			path := strings.Trim(after[1], "/")
			parts := strings.SplitN(path, "/", 2)
			if len(parts) == 2 && parts[1] != "" {
				return "project", parts[0] + "/" + parts[1], nil
			}
			if len(parts) >= 1 && parts[0] != "" {
				return "entity", parts[0], nil
			}
		}
	}

	// Bare base64 view ID (W&B IDs start with "Vmll")
	if strings.HasPrefix(input, "Vmll") {
		return "report", input, nil
	}

	// entity/project reference
	if strings.Contains(input, "/") {
		parts := strings.SplitN(input, "/", 2)
		return "project", parts[0] + "/" + parts[1], nil
	}

	// bare name: entity
	return "entity", input, nil
}

// Locate returns the canonical URL for a (uriType, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "report":
		return "https://wandb.ai/reports/" + id, nil
	case "entity":
		return "https://wandb.ai/" + id, nil
	case "project":
		entity, project, _ := strings.Cut(id, "/")
		return "https://wandb.ai/" + entity + "/" + project, nil
	default:
		return "", errs.Usage("wandb has no resource type %q", uriType)
	}
}

// --- helpers ---

func parseReportID(input string) string {
	if strings.Contains(input, "/reports/") {
		parts := strings.SplitN(input, "/reports/", 2)
		if len(parts) == 2 {
			slug := strings.Split(parts[1], "?")[0]
			idx := strings.LastIndex(slug, "--")
			if idx >= 0 {
				return slug[idx+2:]
			}
			return slug
		}
	}
	return strings.TrimSpace(input)
}

func parseEntityName(input string) string {
	if strings.Contains(input, "wandb.ai/") {
		after := strings.SplitN(input, "wandb.ai/", 2)
		if len(after) == 2 {
			return strings.Split(strings.Trim(after[1], "/"), "/")[0]
		}
	}
	return strings.TrimSpace(input)
}

func parseProjectRef(input string) (entity, project string) {
	if strings.Contains(input, "wandb.ai/") {
		after := strings.SplitN(input, "wandb.ai/", 2)
		if len(after) == 2 {
			parts := strings.SplitN(strings.Trim(after[1], "/"), "/", 2)
			if len(parts) == 2 {
				return parts[0], parts[1]
			}
		}
		return "", ""
	}
	input = strings.TrimSpace(input)
	parts := strings.SplitN(input, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

// mapErr converts library errors into kit error kinds.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return errs.NotFound("%s", err.Error())
	}
	if errors.Is(err, ErrRateLimited) || errors.Is(err, ErrServerError) {
		return errs.RateLimited("%s", err.Error())
	}
	return err
}
