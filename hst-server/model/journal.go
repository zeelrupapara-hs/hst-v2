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

// Journal is one entry of the server journal. Type and code are the log
// taxonomy, already defined as logger.Type and logger.Code in pkg/logger.
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
