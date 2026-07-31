package model

import "fmt"

// The account partition. An engine pod holds a subset of accounts; this decides which.
//
// Two levels on purpose. A login maps to a virtual shard by a hash that never changes, so an
// account stays in the same shard for its whole life. Shards then map to pods, and only that
// second mapping moves when pods are added or lost. Rehashing logins directly onto pod count
// would move nearly every account on every scale event.
//
// This file is duplicated in hst-core (internal/shardmap). It is twenty lines, and a shared
// module for it would couple two services that are otherwise independent — but both copies must
// agree exactly, so they are covered by the same test vector.
const ShardCount = 1024

// ShardOf maps a login to its virtual shard.
//
// FNV-1a rather than the obvious modulo of the login itself: sequential logins would otherwise
// land in sequential shards, so a block of accounts created together would share a pod and the
// load would arrive unevenly.
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

	return uint32(h % uint64(ShardCount))
}

// SubjectShardRequest is where a trade request for this login is published. The shard is in the
// subject, so an engine pod subscribes only to the shards it owns and nothing looks up an owner
// on the hot path.
func SubjectShardRequest(login int64) string {
	return fmt.Sprintf("core.s%d.request.%d", ShardOf(login), login)
}

// SubjectShardAll is what a pod subscribes to for one shard it owns.
func SubjectShardAll(shard uint32) string { return fmt.Sprintf("core.s%d.>", shard) }
