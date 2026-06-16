package wandb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/wandb-cli/wandb"
)

func TestDefaultConfig(t *testing.T) {
	cfg := wandb.DefaultConfig()
	if cfg.Rate <= 0 {
		t.Errorf("Rate = %v, want > 0", cfg.Rate)
	}
	if cfg.Retries <= 0 {
		t.Errorf("Retries = %d, want > 0", cfg.Retries)
	}
	if cfg.Timeout <= 0 {
		t.Errorf("Timeout = %v, want > 0", cfg.Timeout)
	}
	if cfg.UserAgent == "" {
		t.Error("UserAgent is empty")
	}
}

func TestNewClientNotNil(t *testing.T) {
	c := wandb.NewClient(wandb.DefaultConfig())
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestReportRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	want := wandb.Report{
		ID:          "VmlldzoxNDAxMTE=",
		DisplayName: "My Experiment Report",
		EntityName:  "myteam",
		ProjectName: "myproject",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got wandb.Report
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.DisplayName != want.DisplayName {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestEntityRoundTrip(t *testing.T) {
	want := wandb.Entity{
		ID:           "RW50aXR5OjM0OTM=",
		Name:         "stacey",
		IsTeam:       false,
		ProjectCount: 5,
	}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got wandb.Entity
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != want.Name || got.ProjectCount != want.ProjectCount {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestProjectRoundTrip(t *testing.T) {
	want := wandb.Project{
		ID:         "test-project-id",
		Name:       "wandb",
		EntityName: "wandb",
		IsPublic:   true,
		RunCount:   100,
	}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got wandb.Project
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != want.Name || got.RunCount != want.RunCount {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestErrSentinels(t *testing.T) {
	for name, err := range map[string]error{
		"ErrNotFound":    wandb.ErrNotFound,
		"ErrServerError": wandb.ErrServerError,
		"ErrRateLimited": wandb.ErrRateLimited,
	} {
		if err == nil {
			t.Errorf("%s is nil", name)
		}
		if err != nil && err.Error() == "" {
			t.Errorf("%s has empty message", name)
		}
	}
}

func TestGraphQLFromServer(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"entity": map[string]any{
				"id":   "RW50aXR5OjM0OTM=",
				"name": "stacey",
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type: application/json")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	// Directly test the HTTP transport
	resp, err := http.Post(srv.URL, "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["data"] == nil {
		t.Error("data field missing")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := wandb.DefaultConfig()
	cfg.Rate = 0
	cfg.Retries = 0
	c := wandb.NewClient(cfg)

	_, err := c.GetEntity(ctx, "test")
	if err == nil {
		t.Error("GetEntity with cancelled context returned nil error")
	}
}

func TestDefaultConstants(t *testing.T) {
	if wandb.DefaultUserAgent == "" {
		t.Error("DefaultUserAgent is empty")
	}
	if wandb.Host == "" {
		t.Error("Host is empty")
	}
	if wandb.APIEndpoint == "" {
		t.Error("APIEndpoint is empty")
	}
}

func TestReportURLField(t *testing.T) {
	r := wandb.Report{
		ID:          "VmlldzoxNDAxMTE=",
		DisplayName: "My Report",
		EntityName:  "myteam",
		ProjectName: "myproject",
		URL:         "https://wandb.ai/myteam/myproject/reports/my-report--VmlldzoxNDAxMTE=",
	}
	if r.URL == "" {
		t.Error("URL is empty")
	}
}
