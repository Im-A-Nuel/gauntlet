package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoadMissingFileIsConfigError(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "gauntlet init") {
		t.Fatalf("Load on missing config = %v, want an error suggesting `gauntlet init`", err)
	}
}

func TestWriteThenLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := Default()
	if err := Write(dir, c); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !Exists(dir) {
		t.Fatalf("Exists = false after Write")
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, c) {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, c)
	}
}

func TestValidateRejectsBadConfigs(t *testing.T) {
	cases := []struct {
		name string
		mut  func(c Config) Config
	}{
		{"empty baseRef", func(c Config) Config { c.BaseRef = ""; return c }},
		{"empty testCommand", func(c Config) Config { c.TestCommand = ""; return c }},
		{"minScore below 0", func(c Config) Config { c.MinScore = -1; return c }},
		{"minScore above 100", func(c Config) Config { c.MinScore = 101; return c }},
		{"zero concurrency", func(c Config) Config { c.Concurrency = 0; return c }},
		{"negative concurrency", func(c Config) Config { c.Concurrency = -1; return c }},
		{"empty include", func(c Config) Config { c.Include = nil; return c }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.mut(Default())
			if err := c.Validate(); err == nil {
				t.Fatalf("Validate() = nil, want error for %s (config: %+v)", tc.name, c)
			}
		})
	}
}

func TestValidateAcceptsDefault(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("Default().Validate() = %v, want nil", err)
	}
}

func TestLoadRejectsInvalidConfigOnDisk(t *testing.T) {
	dir := t.TempDir()
	bad := Default()
	bad.TestCommand = ""
	if err := Write(dir, bad); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatalf("Load() = nil, want validation error")
	}
}
