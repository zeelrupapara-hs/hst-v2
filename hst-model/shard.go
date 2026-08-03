package model

// The account partition. An engine pod holds a subset of accounts; this decides which.
//
// Defined once, here, because two copies of this hash in two modules would route the same login
// to two different pods the moment either drifted.
const ShardCount = 1024

// ShardOf maps a login to its shard, hashed so sequential logins do not land on one pod.
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
