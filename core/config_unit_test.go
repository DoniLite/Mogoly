package core

import (
	"testing"
)

func TestDiscoverConfigFormat(t *testing.T) {
	cases := []struct{ in, want string }{
		{"conf.json", "json"},
		{"conf.yml", "yaml"},
		{"conf.yaml", "yaml"},
	}
	for _, c := range cases {
		got, err := DiscoverConfigFormat(c.in)
		if err != nil || got != c.want {
			t.Fatalf("DiscoverConfigFormat(%s)=%q,%v want %q,nil", c.in, got, err, c.want)
		}
	}
	if _, err := DiscoverConfigFormat("conf.txt"); err == nil {
		t.Fatalf("expected error for unsupported extension")
	}
}

func TestParseConfig_JSON_YAML_AndBuildURL(t *testing.T) {
	j := []byte(`{"server":[{"name":"a","protocol":"http","host":"127.0.0.1","port":8080}]}`)
	y := []byte("server:\n  - name: b\n    protocol: http\n    host: 127.0.0.1\n    port: 8081\n")
	cfg, err := ParseConfig(j, "json")
	if err != nil || len(cfg.Servers) != 1 {
		t.Fatalf("json parse: %v %#v", err, cfg)
	}
	u, err := buildServerURL(cfg.Servers[0])
	if err != nil || u != "http://127.0.0.1:8080" {
		t.Fatalf("build url: got %q err %v", u, err)
	}
	cfg2, err := ParseConfig(y, "yaml")
	if err != nil || len(cfg2.Servers) != 1 {
		t.Fatalf("yaml parse: %v %#v", err, cfg2)
	}
	u2, err := buildServerURL(cfg2.Servers[0])
	if err != nil || u2 != "http://127.0.0.1:8081" {
		t.Fatalf("build url: got %q err %v", u2, err)
	}
}
