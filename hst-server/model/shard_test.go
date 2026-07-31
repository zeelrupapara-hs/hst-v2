package model

import "testing"

// The engine computes the same shard from the same login, in its own copy of this function
// (hst-core/internal/shardmap). If the two ever drift, this server would publish a trade
// request to a subject no pod is listening on and the order would vanish silently.
//
// These values come from the engine's implementation. Changing either side must break this.
func TestShardOfMatchesTheEngine(t *testing.T) {
	want := map[int64]uint32{
		1:         932,
		1000:      732,
		1001:      477,
		1002:      650,
		500123:    854,
		999999999: 66,
	}

	for login, shard := range want {
		if got := ShardOf(login); got != shard {
			t.Errorf("login %d: this server says shard %d, the engine says %d", login, got, shard)
		}
	}
}

func TestSubjectCarriesTheShard(t *testing.T) {
	if got, want := SubjectShardRequest(1001), "core.s477.request.1001"; got != want {
		t.Errorf("subject = %q, want %q", got, want)
	}
}
