// Package shardmap decides which pod holds which account.
//
// Two steps. A login maps to a shard by a hash that never changes, so an account stays in the
// same shard for life. Shards then map to pods, and only that second step moves when pods come
// and go. Hashing logins straight onto the pod count would move nearly every account every time
// the pod count changed.
package shardmap

import (
	"fmt"
	"sort"
)

// Count is how many shards exist. Fixed forever: changing it would move every account.
const Count = 1024

// ShardOf maps a login to its shard.
//
// The login is hashed rather than divided, because logins are handed out in order and a plain
// remainder would put every account created on the same day onto the same pod.
func ShardOf(login int64) uint32 {
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)

	// the same bits read as unsigned, so the hash can walk them byte by byte
	// #nosec G115 -- reinterpreting the bits, not narrowing a value
	u := uint64(login)

	h := offset64
	for i := 0; i < 8; i++ {
		h ^= (u >> (8 * i)) & 0xff
		h *= prime64
	}

	return uint32(h % uint64(Count))
}

// Subject is where a request for this login is published. The shard is in the subject, so a pod
// subscribes to the shards it owns and nothing has to look up an owner while a trade is waiting.
func Subject(login int64) string {
	return fmt.Sprintf("core.s%d.request.%d", ShardOf(login), login)
}

// SubjectFor is what a pod subscribes to for one shard it owns.
func SubjectFor(shard uint32) string { return fmt.Sprintf("core.s%d.>", shard) }

// Map says which shards belong to this pod, given every pod that is alive.
//
// Each shard goes to the pod whose name hashes closest above it on a ring. Adding or losing a
// pod therefore moves only the shards near that pod's place on the ring, roughly one pod's
// share, instead of reshuffling everything.
type Map struct {
	me     string
	points []point
	mine   map[uint32]bool
}

type point struct {
	hash uint32
	pod  string
}

// replicas spreads each pod over the ring. One point per pod would leave big gaps and some pods
// holding far more shards than others; a hundred smooths that out.
const replicas = 100

// New builds the map for this pod from the list of pods currently alive.
func New(me string, pods []string) *Map {
	m := &Map{me: me, mine: make(map[uint32]bool)}

	if len(pods) == 0 {
		pods = []string{me}
	}

	for _, pod := range pods {
		for i := 0; i < replicas; i++ {
			m.points = append(m.points, point{hash: hashString(fmt.Sprintf("%s#%d", pod, i)), pod: pod})
		}
	}
	sort.Slice(m.points, func(i, j int) bool { return m.points[i].hash < m.points[j].hash })

	for shard := uint32(0); shard < Count; shard++ {
		if m.podFor(shard) == me {
			m.mine[shard] = true
		}
	}

	return m
}

// Mine lists the shards this pod is responsible for, in order.
func (m *Map) Mine() []uint32 {
	out := make([]uint32, 0, len(m.mine))
	for s := range m.mine {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Holds reports whether this pod owns the shard.
func (m *Map) Holds(shard uint32) bool { return m.mine[shard] }

// HoldsLogin reports whether this pod owns the account.
func (m *Map) HoldsLogin(login int64) bool { return m.Holds(ShardOf(login)) }

// podFor walks the ring to the first point at or after the shard, wrapping at the end.
func (m *Map) podFor(shard uint32) string {
	if len(m.points) == 0 {
		return m.me
	}

	h := hashString(fmt.Sprintf("shard#%d", shard))
	i := sort.Search(len(m.points), func(i int) bool { return m.points[i].hash >= h })
	if i == len(m.points) {
		i = 0
	}

	return m.points[i].pod
}

// hashString is the same FNV-1a used for logins, over a string.
func hashString(s string) uint32 {
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)

	h := offset64
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}

	return uint32(h >> 32)
}
