package config_test

import (
	"os"
	"testing"

	"github.com/abemedia/appcast/internal/test"
	"github.com/abemedia/appcast/pkg/config"
	"github.com/google/go-cmp/cmp"
)

func TestSchema(t *testing.T) {
	got := config.Schema()
	want, _ := os.ReadFile("testdata/jsonschema.json")

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}

	if test.Update {
		os.WriteFile("testdata/jsonschema.json", got, 0o644)
	}
}