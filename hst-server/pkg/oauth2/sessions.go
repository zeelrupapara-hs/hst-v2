package oauth2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hstserver/model"
	"hstserver/pkg/cache"
	"hstserver/pkg/crypto"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
)

// ErrSessionNotFound means the session is gone, so the client should refresh.
var ErrSessionNotFound = errors.New("session not found")

// redis keys. Every key is scoped by one id, so moving to redis cluster later
// needs no application change.
var (
	KeySession = func(sid string) string { return "hst:auth:sess:" + sid }
	KeyRefresh = func(hash string) string { return "hst:auth:rt:" + hash }
	KeyFamily  = func(family string) string { return "hst:auth:fam:" + family }
	KeyUser    = func(login int64) string { return fmt.Sprintf("hst:auth:user:%d", login) }
	KeyFail    = func(login int64) string { return fmt.Sprintf("hst:auth:fail:%d", login) }
	KeyFailIP  = func(ip string) string { return "hst:auth:fail:ip:" + ip }
	KeyRTLock  = func(hash string) string { return "hst:auth:rtlock:" + hash }
)

// ChannelInvalidate carries revocations to every instance.
const ChannelInvalidate = "hst:auth:invalidate"

// RefreshRecord is what a refresh token hash resolves to.
type RefreshRecord struct {
	SessionId string `json:"sid"`
	FamilyId  string `json:"fid"`
	Login     int64  `json:"login"`
	ExpiresAt int64  `json:"exp"`
	// Used marks a rotated token. A second use is theft, not a mistake.
	Used bool `json:"used"`
}

// LoadSnapshot reads the session from redis. It never falls back to postgres:
// a missing session means refresh, and refresh is the only path allowed to
// read the database.
func (o *OAuth2) LoadSnapshot(ctx context.Context, sid string) (*cache.Snapshot, error) {
	raw, err := o.Redis.Client.Get(ctx, KeySession(sid)).Bytes()
	if err == redis.Nil {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	snap := &cache.Snapshot{}
	if err := json.Unmarshal(raw, snap); err != nil {
		return nil, err
	}

	return snap, nil
}

// SaveSnapshot writes the session to redis and warms the local cache.
func (o *OAuth2) SaveSnapshot(ctx context.Context, snap *cache.Snapshot) error {
	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}

	ttl := time.Duration(snap.ExpiresAt-time.Now().UnixNano()) * time.Nanosecond
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	pipe := o.Redis.Client.TxPipeline()
	pipe.Set(ctx, KeySession(snap.SessionId), raw, ttl)
	pipe.SAdd(ctx, KeyUser(snap.Login), snap.SessionId)
	pipe.Expire(ctx, KeyUser(snap.Login), o.Cfg.Auth.RefreshAbsoluteTTL)
	pipe.SAdd(ctx, KeyFamily(snap.SessionId), snap.SessionId)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	o.Cache.Put(snap)

	return nil
}

// SaveRefresh stores the refresh record under the hash of the token. The token
// itself is never written anywhere.
func (o *OAuth2) SaveRefresh(ctx context.Context, token string, rec *RefreshRecord) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	ttl := time.Duration(rec.ExpiresAt-time.Now().UnixNano()) * time.Nanosecond
	if ttl <= 0 {
		return fmt.Errorf("refresh token already expired")
	}

	pipe := o.Redis.Client.TxPipeline()
	pipe.Set(ctx, KeyRefresh(crypto.HashTokenHex(token)), raw, ttl)
	pipe.SAdd(ctx, KeyFamily(rec.FamilyId), rec.SessionId)
	pipe.Expire(ctx, KeyFamily(rec.FamilyId), o.Cfg.Auth.RefreshAbsoluteTTL)
	_, err = pipe.Exec(ctx)

	return err
}

// LoadRefresh resolves a refresh token to its record.
func (o *OAuth2) LoadRefresh(ctx context.Context, token string) (*RefreshRecord, error) {
	raw, err := o.Redis.Client.Get(ctx, KeyRefresh(crypto.HashTokenHex(token))).Bytes()
	if err == redis.Nil {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	rec := &RefreshRecord{}
	if err := json.Unmarshal(raw, rec); err != nil {
		return nil, err
	}

	return rec, nil
}

// MarkRefreshUsed leaves the old token behind as a tombstone rather than
// deleting it, so replaying it is detectable instead of merely failing.
func (o *OAuth2) MarkRefreshUsed(ctx context.Context, token string, rec *RefreshRecord) error {
	rec.Used = true

	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return o.Redis.Client.Set(ctx, KeyRefresh(crypto.HashTokenHex(token)), raw, time.Hour).Err()
}

// LockRefresh guards one rotation, so two parallel refreshes of the same token
// do not both succeed and then look like theft.
func (o *OAuth2) LockRefresh(ctx context.Context, token string) (bool, error) {
	return o.Redis.Client.SetNX(ctx, KeyRTLock(crypto.HashTokenHex(token)), 1, 5*time.Second).Result()
}

// RevokeSession kills one session everywhere.
func (o *OAuth2) RevokeSession(ctx context.Context, sid string, reason string) error {
	if _, err := o.DB.DB.Exec(ctx,
		`UPDATE hst.sessions SET revoked_at = $1, revoked_reason = $2
		  WHERE session_id = $3 AND revoked_at = 0`,
		time.Now().UnixNano(), reason, sid); err != nil {
		return err
	}

	if err := o.Redis.Client.Del(ctx, KeySession(sid)).Err(); err != nil {
		return err
	}

	o.Cache.Invalidate(sid)
	o.publish(ctx, invalidateMessage{Sid: sid, Reason: reason})

	return nil
}

// RevokeFamily kills every session in a rotation chain. This is the response to
// a replayed refresh token: the whole family is assumed compromised.
func (o *OAuth2) RevokeFamily(ctx context.Context, familyId string, reason string) error {
	if _, err := o.DB.DB.Exec(ctx,
		`UPDATE hst.sessions SET revoked_at = $1, revoked_reason = $2
		  WHERE family_id = $3 AND revoked_at = 0`,
		time.Now().UnixNano(), reason, familyId); err != nil {
		return err
	}

	sids, err := o.Redis.Client.SMembers(ctx, KeyFamily(familyId)).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	for _, sid := range sids {
		if err := o.Redis.Client.Del(ctx, KeySession(sid)).Err(); err != nil {
			o.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to drop session key",
				"session_id", sid, "error", err.Error())
		}
		o.Cache.Invalidate(sid)
		o.publish(ctx, invalidateMessage{Sid: sid, Reason: reason})
	}

	return o.Redis.Client.Del(ctx, KeyFamily(familyId)).Err()
}

// InvalidateLogin is called after any change to rights, group or a password.
// Sessions are dropped rather than rewritten, so the next request rebuilds from
// current data instead of trusting a snapshot that may already be wrong.
func (o *OAuth2) InvalidateLogin(ctx context.Context, login int64, reason string) error {
	sids, err := o.Redis.Client.SMembers(ctx, KeyUser(login)).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	for _, sid := range sids {
		if err := o.Redis.Client.Del(ctx, KeySession(sid)).Err(); err != nil {
			o.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to drop session key",
				"session_id", sid, "error", err.Error())
		}
	}

	if _, err := o.DB.DB.Exec(ctx,
		`UPDATE hst.sessions SET revoked_at = $1, revoked_reason = $2
		  WHERE login = $3 AND revoked_at = 0`,
		time.Now().UnixNano(), reason, login); err != nil {
		return err
	}

	o.Cache.InvalidateLogin(login)
	o.publish(ctx, invalidateMessage{Login: login, Reason: reason})

	return o.Redis.Client.Del(ctx, KeyUser(login)).Err()
}

// RegisterFailure counts a bad password and locks the login once the limit is
// reached. Returns true when the account is now locked.
func (o *OAuth2) RegisterFailure(ctx context.Context, login int64, ip string) (bool, error) {
	count, err := o.Redis.Client.Incr(ctx, KeyFail(login)).Result()
	if err != nil {
		return false, err
	}
	if err := o.Redis.Client.Expire(ctx, KeyFail(login), o.Cfg.Auth.LockoutDuration).Err(); err != nil {
		return false, err
	}

	if ip != "" {
		if err := o.Redis.Client.Incr(ctx, KeyFailIP(ip)).Err(); err == nil {
			_ = o.Redis.Client.Expire(ctx, KeyFailIP(ip), o.Cfg.Auth.LockoutDuration).Err()
		}
	}

	if int(count) < o.Cfg.Auth.MaxFailedAttempts {
		return false, nil
	}

	until := time.Now().Add(o.Cfg.Auth.LockoutDuration).UnixNano()
	_, err = o.DB.DB.Exec(ctx,
		`UPDATE hst.users SET failed_attempts = $1, locked_until = $2 WHERE login = $3`,
		count, until, login)

	return true, err
}

// ClearFailures resets the counters after a successful login.
func (o *OAuth2) ClearFailures(ctx context.Context, login int64, ip string) {
	keys := []string{KeyFail(login)}
	if ip != "" {
		keys = append(keys, KeyFailIP(ip))
	}
	_ = o.Redis.Client.Del(ctx, keys...).Err()
}

// IPThrottled reports whether this address has failed too often to keep trying.
func (o *OAuth2) IPThrottled(ctx context.Context, ip string) bool {
	if ip == "" {
		return false
	}

	count, err := o.Redis.Client.Get(ctx, KeyFailIP(ip)).Int()
	if err != nil {
		return false
	}

	return count >= o.Cfg.Auth.MaxFailedPerIP
}

// NewSnapshot builds the capability snapshot that every later request reads.
func NewSnapshot(sid string, u *model.User, mgr *model.Manager, cfg *Config, expiresAt int64) *cache.Snapshot {
	snap := &cache.Snapshot{
		SessionId:      sid,
		Login:          u.Login,
		ClientId:       u.ClientId,
		Group:          u.Group,
		Rights:         int64(u.Rights),
		Scope:          cfg.Scope,
		ConnectionType: cfg.ConnectionType,
		Restricted:     cfg.Restricted,
		Version:        1,
		CreatedAt:      time.Now().UnixNano(),
		ExpiresAt:      expiresAt,
	}

	if mgr != nil {
		snap.IsManager = true
		snap.ManagerRights = PackManagerRights(mgr)
	}

	return snap
}
