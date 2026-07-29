package model

type UsersConnectionTypes int32

// banded, clients 0..11 and staff 32 and above
const (
	UsersConnectionTypes_client            UsersConnectionTypes = 0
	UsersConnectionTypes_client_winmobile  UsersConnectionTypes = 1
	UsersConnectionTypes_client_winphone   UsersConnectionTypes = 2
	UsersConnectionTypes_client_api_web    UsersConnectionTypes = 3
	UsersConnectionTypes_client_iphone     UsersConnectionTypes = 4
	UsersConnectionTypes_client_android    UsersConnectionTypes = 5
	UsersConnectionTypes_client_blackberry UsersConnectionTypes = 6
	UsersConnectionTypes_client_web        UsersConnectionTypes = 11
	UsersConnectionTypes_admin             UsersConnectionTypes = 32
	UsersConnectionTypes_manager           UsersConnectionTypes = 33
	UsersConnectionTypes_manager_api       UsersConnectionTypes = 34
	UsersConnectionTypes_admin_api         UsersConnectionTypes = 36
	UsersConnectionTypes_manager_api_web   UsersConnectionTypes = 37
)

// Enum value maps for UsersConnectionTypes.
var (
	UsersConnectionTypes_name = map[int32]string{
		0:  "client",
		1:  "client_winmobile",
		2:  "client_winphone",
		3:  "client_api_web",
		4:  "client_iphone",
		5:  "client_android",
		6:  "client_blackberry",
		11: "client_web",
		32: "admin",
		33: "manager",
		34: "manager_api",
		36: "admin_api",
		37: "manager_api_web",
	}
	UsersConnectionTypes_value = map[string]int32{
		"client":            0,
		"client_winmobile":  1,
		"client_winphone":   2,
		"client_api_web":    3,
		"client_iphone":     4,
		"client_android":    5,
		"client_blackberry": 6,
		"client_web":        11,
		"admin":             32,
		"manager":           33,
		"manager_api":       34,
		"admin_api":         36,
		"manager_api_web":   37,
	}
)

// staffBand is the first connection type reserved for back office terminals.
const staffBand UsersConnectionTypes = 32

// IsStaff reports whether the connection came from an admin or manager terminal.
func (t UsersConnectionTypes) IsStaff() bool { return t >= staffBand }

// Session is the durable record of a login. Live state lives in redis.
type Session struct {
	SessionId      string               `db:"session_id" json:"session_id"`
	Login          int64                `db:"login" json:"login"`
	Scope          UsersPasswords       `db:"scope" json:"scope"`
	ConnectionType UsersConnectionTypes `db:"connection_type" json:"connection_type"`
	TokenHash      []byte               `db:"token_hash" json:"-"`
	Ip             string               `db:"ip" json:"ip"`
	UserAgent      string               `db:"user_agent" json:"user_agent"`
	CreatedAt      int64                `db:"created_at" json:"created_at"`
	LastSeenAt     int64                `db:"last_seen_at" json:"last_seen_at"`
	ExpiresAt      int64                `db:"expires_at" json:"expires_at"`
	RevokedAt      int64                `db:"revoked_at" json:"revoked_at"`
	RevokedReason  string               `db:"revoked_reason" json:"revoked_reason"`
	FamilyId       string               `db:"family_id" json:"family_id"`
	ParentId       string               `db:"parent_id" json:"parent_id"`
}

// revoked_reason values
const (
	SessionRevokedRotated       = "rotated"
	SessionRevokedLogout        = "logout"
	SessionRevokedReuseDetected = "reuse_detected"
	SessionRevokedRightsChanged = "rights_changed"
)

func (Session) TableName() string { return "hst.sessions" }
