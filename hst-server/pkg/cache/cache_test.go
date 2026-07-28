package cache

import (
	"fmt"
	"testing"
	"time"
)

func snap(sid string, login int64) *Snapshot {
	return &Snapshot{
		SessionId: sid,
		Login:     login,
		ExpiresAt: time.Now().Add(time.Hour).UnixNano(),
	}
}

func TestPutAndGet(t *testing.T) {
	d := New(0, 1, 100, 4, time.Minute)
	defer d.Close()

	d.Put(snap("s1", 1))

	got, ok := d.Get("s1")
	if !ok || got.Login != 1 {
		t.Fatalf("get: ok=%v snap=%+v", ok, got)
	}

	if _, ok := d.Get("missing"); ok {
		t.Fatal("returned a session that was never stored")
	}
}

// With one shard the instance owns everything; with two it owns about half.
func TestOwnsPartitionsTheKeyspace(t *testing.T) {
	single := New(0, 1, 100, 4, time.Minute)
	defer single.Close()

	for i := 0; i < 100; i++ {
		if !single.Owns(fmt.Sprintf("sid-%d", i)) {
			t.Fatal("a single shard must own every session")
		}
	}

	a := New(0, 2, 100, 4, time.Minute)
	b := New(1, 2, 100, 4, time.Minute)
	defer a.Close()
	defer b.Close()

	owned := 0
	for i := 0; i < 1000; i++ {
		sid := fmt.Sprintf("sid-%d", i)
		if a.Owns(sid) == b.Owns(sid) {
			t.Fatalf("session %s is owned by both shards or by neither", sid)
		}
		if a.Owns(sid) {
			owned++
		}
	}

	// the split need not be exact, but a wild skew means the hash is wrong
	if owned < 350 || owned > 650 {
		t.Fatalf("shard 0 owns %d of 1000, expected roughly half", owned)
	}
}

// Refusing foreign sessions is what bounds memory per instance.
func TestPutIgnoresForeignSessions(t *testing.T) {
	d := New(0, 2, 100, 4, time.Minute)
	defer d.Close()

	stored, skipped := 0, 0
	for i := 0; i < 200; i++ {
		sid := fmt.Sprintf("sid-%d", i)
		d.Put(snap(sid, int64(i)))

		if _, ok := d.Get(sid); ok {
			stored++
		} else {
			skipped++
		}
	}

	if skipped == 0 {
		t.Fatal("every session was cached, foreign shards are not being refused")
	}
	if stored == 0 {
		t.Fatal("no session was cached at all")
	}
}

func TestInvalidate(t *testing.T) {
	d := New(0, 1, 100, 4, time.Minute)
	defer d.Close()

	d.Put(snap("s1", 1))
	d.Invalidate("s1")

	if _, ok := d.Get("s1"); ok {
		t.Fatal("invalidated session is still cached")
	}
}

// A rights change must reach every session of the login, not just one.
func TestInvalidateLogin(t *testing.T) {
	d := New(0, 1, 100, 4, time.Minute)
	defer d.Close()

	d.Put(snap("s1", 42))
	d.Put(snap("s2", 42))
	d.Put(snap("s3", 99))

	d.InvalidateLogin(42)

	if _, ok := d.Get("s1"); ok {
		t.Fatal("s1 survived the login invalidation")
	}
	if _, ok := d.Get("s2"); ok {
		t.Fatal("s2 survived the login invalidation")
	}
	if _, ok := d.Get("s3"); !ok {
		t.Fatal("s3 belongs to another login and should have survived")
	}
}

// The LRU is the safety valve: a low cap must bound memory, not OOM.
func TestEvictionBoundsMemory(t *testing.T) {
	d := New(0, 1, 8, 1, time.Minute)
	defer d.Close()

	for i := 0; i < 500; i++ {
		d.Put(snap(fmt.Sprintf("sid-%d", i), int64(i)))
	}

	if got := d.Stats().Entries; got > 8 {
		t.Fatalf("cache holds %d entries, cap is 8", got)
	}
	if d.Stats().Evictions == 0 {
		t.Fatal("nothing was evicted despite exceeding the cap")
	}
}

func TestExpiryDropsEntry(t *testing.T) {
	d := New(0, 1, 100, 4, 10*time.Millisecond)
	defer d.Close()

	d.Put(snap("s1", 1))
	time.Sleep(50 * time.Millisecond)

	if _, ok := d.Get("s1"); ok {
		t.Fatal("an entry past its TTL is still being served")
	}
}
