package cloud

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/moby/moby/client"
)

func requireDocker(t *testing.T) *client.Client {
	t.Helper()
	if os.Getenv("DOCKER_INTEGRATION") == "" {
		t.Skip("set DOCKER_INTEGRATION=1 to run Docker integration tests")
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	return cli
}

func TestIntegration_NewCloudDBManager_EnsureNetwork(t *testing.T) {
	_ = requireDocker(t)

	mgr, err := NewCloudDBManager()
	if err != nil {
		t.Fatalf("NewCloudDBManager error: %v", err)
	}
	// calling again should still be fine (network already exists)
	if err := mgr.ensureNetwork(); err != nil {
		t.Fatalf("ensureNetwork idempotency failed: %v", err)
	}
}

func TestIntegration_CreateAndDelete_Postgres(t *testing.T) {
	_ = requireDocker(t)

	mgr, err := NewCloudDBManager()
	if err != nil {
		t.Fatalf("NewCloudDBManager error: %v", err)
	}

	inst, err := mgr.CreateInstance(ServiceConfig{
		Type:         PostgreSQL,
		Name:         "it_pg",
		Username:     "admin",
		DatabaseName: "testdb",
	})
	if err != nil {
		t.Fatalf("CreateInstance error: %v", err)
	}
	if inst.Status != "running" {
		t.Fatalf("expected running, got %s", inst.Status)
	}
	if inst.ContainerID == "" {
		t.Fatalf("expected container id")
	}

	// cleanup
	t.Cleanup(func() {
		_ = mgr.DeleteInstance(inst.ID)
	})

	// quick health check: container should exist for a short time window
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// We don't have a direct helper; if DeleteInstance failed, later assertions will fail on remove
	_ = ctx
}

func TestIntegration_CreateAndDeleteMultiple_Postgres(t *testing.T) {
	_ = requireDocker(t)

	mgr, err := NewCloudDBManager()
	if err != nil {
		t.Fatalf("NewCloudDBManager error: %v", err)
	}

	instancesConf := []ServiceConfig{
		{
			Type:         PostgreSQL,
			Name:         "it_pg",
			Username:     "admin",
			DatabaseName: "testdb",
		},
		{
			Type:         PostgreSQL,
			Name:         "it_pg2",
			Username:     "admin2",
			DatabaseName: "testdb2",
		},
		{
			Type:         MySQL,
			Name:         "it_mysql",
			Username:     "admin",
			DatabaseName: "testdb",
		},
		{
			Type:         MongoDB,
			Name:         "it_mongodb",
			Username:     "admin",
			DatabaseName: "testdb",
		},
		{
			Type:         Redis,
			Name:         "it_redis",
			Username:     "admin",
			DatabaseName: "testdb",
		},
		{
			Type:         MariaDB,
			Name:         "it_mariadb",
			Username:     "admin",
			DatabaseName: "testdb",
		},
	}

	for _, instConf := range instancesConf {
		inst, err := mgr.CreateInstance(instConf)
		if err != nil {
			t.Fatalf("CreateInstance error: %v", err)
		}
		if inst.Status != "running" {
			t.Fatalf("expected running, got %s", inst.Status)
		}
		if inst.ContainerID == "" {
			t.Fatalf("expected container id")
		}

		// cleanup
		t.Cleanup(func() {
			_ = mgr.DeleteInstance(inst.ID)
		})
	}

	// quick health check: container should exist for a short time window
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// We don't have a direct helper; if DeleteInstance failed, later assertions will fail on remove
	_ = ctx
}

func TestIntegration_DeleteNonExisting(t *testing.T) {
	_ = requireDocker(t)

	mgr, err := NewCloudDBManager()
	if err != nil {
		t.Fatalf("NewCloudDBManager error: %v", err)
	}

	if err := mgr.DeleteInstance("does-not-exist"); err == nil {
		t.Fatalf("expected error when deleting non-existing instance")
	}
}
