package wandb

import (
	"fmt"
	"strings"
	"time"
)

// Report is one W&B public workspace view (a report).
type Report struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description,omitempty"`
	EntityName  string    `json:"entityName"`
	ProjectName string    `json:"projectName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CoverURL    string    `json:"coverUrl,omitempty"`
	URL         string    `json:"url,omitempty"`
}

// Entity is a W&B user or team.
type Entity struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	IsTeam       bool      `json:"isTeam"`
	MemberCount  int       `json:"memberCount"`
	PhotoURL     string    `json:"photoUrl,omitempty"`
	ProjectCount int       `json:"projectCount"`
	CreatedAt    time.Time `json:"createdAt"`
	URL          string    `json:"url"`
}

// Project is a W&B project.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	EntityName  string    `json:"entityName"`
	Description string    `json:"description,omitempty"`
	IsPublic    bool      `json:"isPublic"`
	RunCount    int       `json:"runCount"`
	CreatedAt   time.Time `json:"createdAt"`
	URL         string    `json:"url"`
}

// reportURL builds the canonical wandb.ai URL for a report.
func reportURL(entityName, projectName, displayName, id string) string {
	slug := strings.ReplaceAll(strings.ToLower(displayName), " ", "-")
	if slug == "" {
		return fmt.Sprintf("https://wandb.ai/%s/%s/reports/%s", entityName, projectName, id)
	}
	return fmt.Sprintf("https://wandb.ai/%s/%s/reports/%s--%s", entityName, projectName, slug, id)
}

// --- raw GraphQL shapes ---

type rawView struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	EntityName  string    `json:"entityName"`
	ProjectName string    `json:"projectName"`
	CoverURL    *string   `json:"coverUrl"`
}

type rawViewEdge struct {
	Node   rawView `json:"node"`
	Cursor string  `json:"cursor"`
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type rawViewConnection struct {
	Edges    []rawViewEdge `json:"edges"`
	PageInfo pageInfo      `json:"pageInfo"`
}

type rawEntity struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	IsTeam       bool      `json:"isTeam"`
	MemberCount  int       `json:"memberCount"`
	PhotoURL     *string   `json:"photoUrl"`
	ProjectCount int       `json:"projectCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

type rawProject struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	EntityName  string    `json:"entityName"`
	Description *string   `json:"description"`
	IsPublic    bool      `json:"isPublic"`
	RunCount    int       `json:"runCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

func fromRawView(r rawView) Report {
	rpt := Report{
		ID:          r.ID,
		DisplayName: r.DisplayName,
		EntityName:  r.EntityName,
		ProjectName: r.ProjectName,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.Description != nil {
		rpt.Description = *r.Description
	}
	if r.CoverURL != nil {
		rpt.CoverURL = *r.CoverURL
	}
	rpt.URL = reportURL(r.EntityName, r.ProjectName, r.DisplayName, r.ID)
	return rpt
}

func fromRawEntity(r rawEntity) Entity {
	e := Entity{
		ID:           r.ID,
		Name:         r.Name,
		IsTeam:       r.IsTeam,
		MemberCount:  r.MemberCount,
		ProjectCount: r.ProjectCount,
		CreatedAt:    r.CreatedAt,
		URL:          "https://wandb.ai/" + r.Name,
	}
	if r.PhotoURL != nil {
		e.PhotoURL = *r.PhotoURL
	}
	return e
}

func fromRawProject(r rawProject) Project {
	p := Project{
		ID:         r.ID,
		Name:       r.Name,
		EntityName: r.EntityName,
		IsPublic:   r.IsPublic,
		RunCount:   r.RunCount,
		CreatedAt:  r.CreatedAt,
		URL:        fmt.Sprintf("https://wandb.ai/%s/%s", r.EntityName, r.Name),
	}
	if r.Description != nil {
		p.Description = *r.Description
	}
	return p
}
