package model

import "encoding/json"

// JournalMode is which entries a journal query returns, filtering on severity.
type JournalMode int32

const (
	JournalMode_full           JournalMode = 0
	JournalMode_without_logins JournalMode = 1
	JournalMode_errors_only    JournalMode = 2
)

var (
	JournalMode_name = map[int32]string{
		0: "full",
		1: "without_logins",
		2: "errors_only",
	}
	JournalMode_value = map[string]int32{
		"full":           0,
		"without_logins": 1,
		"errors_only":    2,
	}
)

// JournalType is the module an entry belongs to.
type JournalType int32

const (
	JournalType_system    JournalType = 0
	JournalType_accounts  JournalType = 1
	JournalType_clients   JournalType = 2
	JournalType_groups    JournalType = 3
	JournalType_symbols   JournalType = 4
	JournalType_trade     JournalType = 5
	JournalType_managers  JournalType = 6
	JournalType_routing   JournalType = 7
	JournalType_datafeeds JournalType = 8
	JournalType_mail      JournalType = 9
	JournalType_holidays  JournalType = 10
	JournalType_leverage  JournalType = 11
	JournalType_auth      JournalType = 12
)

var (
	JournalType_name = map[int32]string{
		0:  "system",
		1:  "accounts",
		2:  "clients",
		3:  "groups",
		4:  "symbols",
		5:  "trade",
		6:  "managers",
		7:  "routing",
		8:  "datafeeds",
		9:  "mail",
		10: "holidays",
		11: "leverage",
		12: "auth",
	}
	JournalType_value = map[string]int32{
		"system":    0,
		"accounts":  1,
		"clients":   2,
		"groups":    3,
		"symbols":   4,
		"trade":     5,
		"managers":  6,
		"routing":   7,
		"datafeeds": 8,
		"mail":      9,
		"holidays":  10,
		"leverage":  11,
		"auth":      12,
	}
)

// Journal is one entry of the server journal. Type is the module taxonomy
// (JournalType), code is the severity defined as logger.Code in pkg/logger.
type Journal struct {
	JournalId int64           `db:"journal_id" json:"journal_id"`
	CreatedAt int64           `db:"created_at" json:"created_at" format:"int64"`
	Type      int32           `db:"type" json:"type"`
	Code      int32           `db:"code" json:"code"`
	Login     int64           `db:"login" json:"login"`
	Ip        string          `db:"ip" json:"ip"`
	Channel   string          `db:"channel" json:"channel"`
	Os        string          `db:"os" json:"os"`
	Message   string          `db:"message" json:"message"`
	Detail    json.RawMessage `db:"detail" json:"detail" swaggertype:"object"`
}

func (Journal) TableName() string { return "hst.journal" }
