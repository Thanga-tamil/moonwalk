package config

import (
	"os"
	"path/filepath"
	"testing"
	"moonwalk/pkg"
)

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	path := writeConfig(t, `{
		"sqlDriverName": "sqlite",
		"sqlDataSourceName": "/tmp/test.db",
		"logLevel": "INFO",
		"serverMode": "test"
	}`)

	c := LoadConfig(path)

	if c.SchedulerStrategy != pkg.StrategyAuto {
		t.Fatalf("expected schedulerStrategy default auto, got %q", c.SchedulerStrategy)
	}
	if c.DbMaxOpenConns == 0 {
		t.Fatal("expected dbMaxOpenConns default to be set")
	}
	if c.DbConnMaxLifetime == 0 {
		t.Fatal("expected dbConnMaxLifetime default to be set")
	}
}

func TestLoadConfigUsesProvidedStrategy(t *testing.T) {
	path := writeConfig(t, `{
		"sqlDriverName": "sqlite",
		"sqlDataSourceName": "/tmp/test.db",
		"logLevel": "INFO",
		"serverMode": "test",
		"schedulerStrategy": "fifo",
		"dbMaxOpenConns": 25
	}`)

	c := LoadConfig(path)

	if c.SchedulerStrategy != pkg.StrategyFIFO {
		t.Fatalf("expected schedulerStrategy fifo, got %q", c.SchedulerStrategy)
	}
	if c.DbMaxOpenConns != 25 {
		t.Fatalf("expected dbMaxOpenConns 25, got %d", c.DbMaxOpenConns)
	}
}
