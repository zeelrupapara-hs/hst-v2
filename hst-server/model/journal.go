package model

import "encoding/json"

// Journal is one entry of the server journal. Type and code are the MT5 log
// taxonomy, already defined as logger.Type and logger.Code in pkg/logger.
type Journal struct {
	JournalId int64           `db:"journal_id" json:"journal_id"`
	CreatedAt int64           `db:"created_at" json:"created_at" format:"int64"`
	Type      int32           `db:"type" json:"type"`
	Code      int32           `db:"code" json:"code"`
	Login     int64           `db:"login" json:"login"`
	Ip        string          `db:"ip" json:"ip"`
	Message   string          `db:"message" json:"message"`
	Detail    json.RawMessage `db:"detail" json:"detail" swaggertype:"object"`
}

func (Journal) TableName() string { return "hst.journal" }
