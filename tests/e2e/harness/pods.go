package harness

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

var (
	podsMu sync.Mutex
	pods   = map[string]*exec.Cmd{}
)

// StartCore runs one more hst-core pod from source with its own name and health port.
func (e *Env) StartCore(t *testing.T, name string, healthPort int) {
	t.Helper()
	dir, _ := filepath.Abs(filepath.Join("..", "..", "..", "hst-core"))
	cmd := exec.Command("go", "run", "./cmd")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "POD_NAME="+name, fmt.Sprintf("HEALTH_PORT=%d", healthPort))
	logFile, _ := os.Create(filepath.Join("..", "logs", "core-"+name+".log"))
	cmd.Stdout, cmd.Stderr = logFile, logFile
	if err := cmd.Start(); err != nil {
		t.Fatalf("start core %s: %v", name, err)
	}
	podsMu.Lock()
	pods[name] = cmd
	podsMu.Unlock()
	t.Cleanup(func() { e.StopCore(name) })

	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		if resp, err := http.Get(fmt.Sprintf("http://localhost:%d/readyz", healthPort)); err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("core %s never became ready on %d", name, healthPort)
}

// StopCore sends SIGTERM to a pod started by StartCore and waits for it.
func (e *Env) StopCore(name string) {
	podsMu.Lock()
	cmd := pods[name]
	delete(pods, name)
	podsMu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(os.Interrupt)
	_ = cmd.Wait()
}

// RedisKeys lists keys by pattern, for shard and session assertions.
func (e *Env) RedisKeys(t *testing.T, pattern string) []string {
	t.Helper()
	keys, err := e.Redis.Keys(context.Background(), pattern).Result()
	if err != nil {
		t.Fatalf("redis keys %s: %v", pattern, err)
	}
	return keys
}
