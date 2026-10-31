package docs

import (
	"go.yaml.in/yaml/v3"
	"testing"
)

func TestCronSpec(t *testing.T) {
	var doc struct {
		Paths      map[string]any `yaml:"paths"`
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/crons", "/crons/{cron}/runs", "/admin/crons", "/admin/crons/{cron}"} {
		if doc.Paths[path] == nil {
			t.Errorf("missing path %s", path)
		}
	}
	if doc.Components.Schemas["CronInput"] == nil {
		t.Fatal("missing cron input schema")
	}
}
