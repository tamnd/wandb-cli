package wandb

import (
	"context"
	"encoding/json"
	"strings"
)

// ListReports returns public W&B workspace views.
// viewType is "runs" for workspace views or empty for all.
// limit caps the number of results (default 20, max 50).
// after is the pagination cursor (empty for first page).
func (c *Client) ListReports(ctx context.Context, viewType string, limit int, after string) ([]Report, string, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	vars := map[string]any{
		"first": limit,
	}
	if viewType != "" {
		vars["type"] = viewType
	}
	if after != "" {
		vars["after"] = after
	}

	data, err := c.postGQL(ctx, queryPublicViews, vars)
	if err != nil {
		return nil, "", false, err
	}

	var resp struct {
		PublicViews rawViewConnection `json:"publicViews"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", false, err
	}

	out := make([]Report, len(resp.PublicViews.Edges))
	for i, e := range resp.PublicViews.Edges {
		out[i] = fromRawView(e.Node)
	}
	return out, resp.PublicViews.PageInfo.EndCursor, resp.PublicViews.PageInfo.HasNextPage, nil
}

// FeaturedReport returns the pinned gallery report.
func (c *Client) FeaturedReport(ctx context.Context) (*Report, error) {
	data, err := c.postGQL(ctx, queryFeaturedReports, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		FeaturedReports *rawView `json:"featuredReports"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.FeaturedReports == nil {
		return nil, ErrNotFound
	}
	rpt := fromRawView(*resp.FeaturedReports)
	return &rpt, nil
}

// SearchReports does a full-text search across public report names.
// Returns null as an empty slice (server returns null for no results).
func (c *Client) SearchReports(ctx context.Context, query string) ([]Report, error) {
	data, err := c.postGQL(ctx, queryReportSearch, map[string]any{"query": query})
	if err != nil {
		return nil, err
	}

	var resp struct {
		ReportSearch *struct {
			Edges []rawViewEdge `json:"edges"`
		} `json:"reportSearch"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.ReportSearch == nil {
		return nil, nil // server returns null for no results
	}

	out := make([]Report, len(resp.ReportSearch.Edges))
	for i, e := range resp.ReportSearch.Edges {
		out[i] = fromRawView(e.Node)
	}
	return out, nil
}

// GetReport fetches a single public view by its ID from the first page of
// public views. This is a v1 limitation: the W&B API does not expose a
// view-by-ID query without entity/project context.
func (c *Client) GetReport(ctx context.Context, id string) (*Report, error) {
	// Fetch up to 50 public views and find the matching ID.
	reports, _, _, err := c.ListReports(ctx, "runs", 50, "")
	if err != nil {
		return nil, err
	}
	for _, r := range reports {
		if r.ID == id || strings.HasSuffix(r.URL, "--"+id) {
			return &r, nil
		}
	}
	return nil, ErrNotFound
}
