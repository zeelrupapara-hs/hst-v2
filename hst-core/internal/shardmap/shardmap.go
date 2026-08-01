// Package shardmap decides which pod holds which account: login -> shard by a fixed hash, shard -> pod on a ring.
package shardmap

import (
	"fmt"
	"slices"
	"sort"
)

// Count is how many shards exist. Fixed forever: changing it would move every account.
const Count = 1024

// ShardOf maps a login to its shard. Hashed, not divided: logins are handed out in order.
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

// Map says which shards belong to this pod, given every pod that is alive.
type Map struct {
	me     string
	points []point
	mine   map[uint32]bool
}

type point struct {
	hash uint32
	pod  string
}

// replicas spreads each pod over the ring so the shares come out even.
const replicas = 100

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

func (m *Map) Mine() []uint32 {
	out := make([]uint32, 0, len(m.mine))
	for s := range m.mine {
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

func (m *Map) Holds(shard uint32) bool { return m.mine[shard] }

func (m *Map) HoldsLogin(login int64) bool { return m.Holds(ShardOf(login)) }

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

// hashString places a name on the ring; the avalanche keeps near-identical pod names apart.
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

	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33

	// #nosec G115 -- the fold to 32 bits is the point of the hash
	return uint32(h)
}
