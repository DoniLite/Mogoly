package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchConfig_ReloadsOnChange(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "conf.yml")
	orig := []byte("server:\n  - name: a\n    protocol: http\n    host: 127.0.0.1\n    port: 8080\n")
	err := os.WriteFile(fp, orig, 0644)
	if err != nil {
		t.Fatalf("error writing original config file: %v", err)
	}

	reloaded := make(chan *Config, 1)
	if err := WatchConfig(fp, func(c *Config) { reloaded <- c }); err != nil {
		t.Fatalf("watch: %v", err)
	}
	time.Sleep(150 * time.Millisecond) // allow watcher to start

	// mutate
	updated := []byte("server:\n  - name: b\n    protocol: http\n    host: 127.0.0.1\n    port: 8081\n")
	if err := os.WriteFile(fp, updated, 0644); err != nil {
		t.Fatalf("error during file content writing: %v", err)
	}

	select {
	case cfg := <-reloaded:
		if len(cfg.Servers) != 1 || cfg.Servers[0].Name != "b" {
			t.Fatalf("unexpected reload: %#v", cfg)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting reload")
	}
}
