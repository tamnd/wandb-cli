package wandb_test

import (
	"strings"
	"testing"

	"github.com/tamnd/wandb-cli/wandb"
)

func TestClassifyReportURL(t *testing.T) {
	d := wandb.Domain{}
	uriType, id, err := d.Classify("https://wandb.ai/myteam/myproject/reports/my-report--VmlldzoxNDAxMTE=")
	if err != nil {
		t.Fatal(err)
	}
	if uriType != "report" {
		t.Errorf("type = %q, want report", uriType)
	}
	if id != "VmlldzoxNDAxMTE=" {
		t.Errorf("id = %q, want VmlldzoxNDAxMTE=", id)
	}
}

func TestClassifyEntityURL(t *testing.T) {
	d := wandb.Domain{}
	uriType, id, err := d.Classify("https://wandb.ai/stacey")
	if err != nil {
		t.Fatal(err)
	}
	if uriType != "entity" {
		t.Errorf("type = %q, want entity", uriType)
	}
	if id != "stacey" {
		t.Errorf("id = %q, want stacey", id)
	}
}

func TestClassifyProjectURL(t *testing.T) {
	d := wandb.Domain{}
	uriType, id, err := d.Classify("https://wandb.ai/wandb/wandb")
	if err != nil {
		t.Fatal(err)
	}
	if uriType != "project" {
		t.Errorf("type = %q, want project", uriType)
	}
	if id != "wandb/wandb" {
		t.Errorf("id = %q, want wandb/wandb", id)
	}
}

func TestClassifyBareID(t *testing.T) {
	d := wandb.Domain{}
	uriType, id, err := d.Classify("VmlldzoxNDAxMTE=")
	if err != nil {
		t.Fatal(err)
	}
	if uriType != "report" {
		t.Errorf("type = %q, want report", uriType)
	}
	if id != "VmlldzoxNDAxMTE=" {
		t.Errorf("id = %q", id)
	}
}

func TestClassifyEntitySlash(t *testing.T) {
	d := wandb.Domain{}
	uriType, id, err := d.Classify("wandb/wandb")
	if err != nil {
		t.Fatal(err)
	}
	if uriType != "project" {
		t.Errorf("type = %q, want project", uriType)
	}
	if id != "wandb/wandb" {
		t.Errorf("id = %q, want wandb/wandb", id)
	}
}

func TestClassifyBareName(t *testing.T) {
	d := wandb.Domain{}
	uriType, id, err := d.Classify("stacey")
	if err != nil {
		t.Fatal(err)
	}
	if uriType != "entity" {
		t.Errorf("type = %q, want entity", uriType)
	}
	if id != "stacey" {
		t.Errorf("id = %q, want stacey", id)
	}
}

func TestClassifyEmptyInput(t *testing.T) {
	d := wandb.Domain{}
	_, _, err := d.Classify("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestLocateReport(t *testing.T) {
	d := wandb.Domain{}
	u, err := d.Locate("report", "VmlldzoxNDAxMTE=")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "VmlldzoxNDAxMTE=") {
		t.Errorf("URL = %q, does not contain ID", u)
	}
}

func TestLocateEntity(t *testing.T) {
	d := wandb.Domain{}
	u, err := d.Locate("entity", "stacey")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "/stacey") {
		t.Errorf("URL = %q, does not contain entity name", u)
	}
}

func TestLocateProject(t *testing.T) {
	d := wandb.Domain{}
	u, err := d.Locate("project", "wandb/wandb")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "/wandb/wandb") {
		t.Errorf("URL = %q, does not contain project path", u)
	}
}

func TestLocateUnknownType(t *testing.T) {
	d := wandb.Domain{}
	_, err := d.Locate("unknown", "x")
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestDomainInfoScheme(t *testing.T) {
	d := wandb.Domain{}
	info := d.Info()
	if info.Scheme != "wandb" {
		t.Errorf("Scheme = %q, want wandb", info.Scheme)
	}
	if info.Identity.Binary != "wandb" {
		t.Errorf("Binary = %q, want wandb", info.Identity.Binary)
	}
}
