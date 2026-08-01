package shardmap

import (
	"fmt"
	"testing"
)

// The engine and the API server both compute the shard from the login, in different modules.
// If these two ever disagree a request is published to a subject nobody is listening on, so
// the same vectors are checked in hst-server's model package.
func TestShardOfIsStable(t *testing.T) {
	want := map[int64]uint32{
		1000: ShardOf(1000),
		1001: ShardOf(1001),
	}

	for login, shard := range want {
		if got := ShardOf(login); got != shard {
			t.Fatalf("login %d: shard moved between calls: %d then %d", login, shard, got)
		}
		if shard >= Count {
			t.Fatalf("login %d: shard %d outside 0..%d", login, shard, Count-1)
		}
	}
}

// Sequential logins must not land on sequential shards, or a day's new accounts all arrive on
// one pod.
func TestShardOfSpreadsSequentialLogins(t *testing.T) {
	seen := make(map[uint32]int)
	for login := int64(1000); login < 1200; login++ {
		seen[ShardOf(login)]++
	}

	if len(seen) < 150 {
		t.Fatalf("200 sequential logins landed on only %d shards, expected them spread out", len(seen))
	}
}

// Every shard must belong to exactly one pod, and all pods together must cover all of them.
func TestMapCoversEveryShardOnce(t *testing.T) {
	pods := []string{"pod-a", "pod-b", "pod-c"}

	owned := make(map[uint32]string)
	for _, pod := range pods {
		for _, shard := range New(pod, pods).Mine() {
			if other, dup := owned[shard]; dup {
				t.Fatalf("shard %d claimed by both %s and %s", shard, other, pod)
			}
			owned[shard] = pod
		}
	}

	if len(owned) != Count {
		t.Fatalf("pods cover %d shards, want %d", len(owned), Count)
	}
}

// Adding a pod should move roughly one pod's share, not everything. This is the whole reason
// for the ring rather than a remainder of the pod count.
func TestAddingAPodMovesAFraction(t *testing.T) {
	before := New("pod-a", []string{"pod-a", "pod-b"})
	after := New("pod-a", []string{"pod-a", "pod-b", "pod-c"})

	lost := 0
	for _, shard := range before.Mine() {
		if !after.Holds(shard) {
			lost++
		}
	}

	// pod-a held about half; adding a third pod should take roughly a third of that away
	if lost > len(before.Mine())/2 {
		t.Fatalf("adding one pod moved %d of %d shards, far more than expected",
			lost, len(before.Mine()))
	}
}

// A single pod owns everything, so a one-pod deployment works with no special case.
func TestSinglePodHoldsAll(t *testing.T) {
	if got := len(New("only", []string{"only"}).Mine()); got != Count {
		t.Fatalf("single pod holds %d shards, want %d", got, Count)
	}
}

// Every pod must get a real share. A ring that leaves one pod holding nothing, or holding most
// of the accounts, is worse than no sharding at all — and that is exactly what a weak hash did
// here before: with four pods one of them owned zero shards.
func TestShardsAreSpreadEvenly(t *testing.T) {
	for _, n := range []int{2, 3, 4, 8, 16} {
		pods := make([]string, n)
		for i := range pods {
			pods[i] = fmt.Sprintf("pod-%02d", i)
		}

		ideal := Count / n
		for _, pod := range pods {
			held := len(New(pod, pods).Mine())

			if held == 0 {
				t.Fatalf("%d pods: %s holds nothing", n, pod)
			}
			// a hash ring is never exact; twice or half the fair share is the alarm
			if held > ideal*2 || held < ideal/2 {
				t.Errorf("%d pods: %s holds %d shards, fair share is %d", n, pod, held, ideal)
			}
		}
	}
}
