// Package journal records what happened on the server: who did it, to which
// record, and when.
//
// The file log in logs/YYYYMMDD.log keeps the same entries for an operator
// reading a terminal. This table is the one the back office queries and
// filters, so it is written deliberately rather than scraped back out of text.
package journal

import (
	"context"
	"math"
	"time"

	"hstserver/model"
	"hstserver/pkg/db"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"

	"github.com/goccy/go-json"
	natscore "github.com/nats-io/nats.go"
)

// Journal writes entries and announces them.
type Journal struct {
	db   *db.PostgresDB
	nats *nats.Nats
	log  *logger.Logger
}

// New wires the writer.
func New(database *db.PostgresDB, nc *nats.Nats, log *logger.Logger) *Journal {
	return &Journal{db: database, nats: nc, log: log}
}

// Entry is one line of the journal, before it has been stored.
//
// Type and Code are the taxonomy pkg/logger already uses, so an entry reads the
// same whether it came from the file or the table.
type Entry struct {
	Type logger.Type
	Code logger.Code
	// Login is who did it, not who it was done to.
	Login int64
	// Ip is the address the request came from, as the header reader resolved it.
	Ip string
	// Message is one line a human reads.
	Message string
	// Detail is whatever the caller wants to keep: the record that changed, the
	// fields that moved, an error. Stored as jsonb, so it stays queryable.
	Detail any
}

// Write stores the entry and then announces it.
//
// The row lands first and the announcement follows, never the other way round:
// an entry that was broadcast but not stored would appear in a live view and
// then vanish the next time the page was loaded, which is worse than not
// showing it at all.
//
// The announcement failing does not fail the write. The entry is safe in the
// table by then, and a live view catching up a moment later is a smaller
// problem than a caller believing its record was not journalled.
func (j *Journal) Write(ctx context.Context, e Entry) error {
	detail := json.RawMessage("{}")
	if e.Detail != nil {
		raw, err := json.Marshal(e.Detail)
		if err != nil {
			// a detail that cannot be encoded is the caller's bug, and losing
			// the whole entry over it would hide what actually happened
			j.log.Log(logger.TypeSys, logger.CodeWarn, "could not encode a journal detail",
				"message", e.Message, "error", err.Error())
		} else {
			detail = raw
		}
	}

	entry := model.Journal{
		CreatedAt: time.Now().UnixNano(),
		Type:      narrow(int(e.Type)),
		Code:      narrow(int(e.Code)),
		Login:     e.Login,
		Ip:        e.Ip,
		Message:   e.Message,
		Detail:    detail,
	}

	// host() drops the /32 that inet carries, and an empty string is not an
	// address, so it is stored as null rather than rejected
	var ip *string
	if e.Ip != "" {
		ip = &e.Ip
	}

	if err := j.db.DB.QueryRow(ctx,
		`INSERT INTO hst.journal (created_at, type, code, login, ip, message, detail)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING journal_id`,
		entry.CreatedAt, entry.Type, entry.Code, entry.Login, ip, entry.Message, entry.Detail,
	).Scan(&entry.JournalId); err != nil {
		return err
	}

	j.announce(entry)

	return nil
}

// narrow takes a taxonomy value down to the column's width. Both enumerations
// fit in a byte, so the bound is not there for today's values; it is there so a
// future one cannot wrap silently into something that reads as a valid code.
func narrow(v int) int32 {
	if v < 0 || v > math.MaxInt32 {
		return 0
	}
	return int32(v)
}

// announce publishes a stored entry to the managers watching the journal.
func (j *Journal) announce(entry model.Journal) {
	raw, err := json.Marshal(entry)
	if err != nil {
		j.log.Log(logger.TypeSys, logger.CodeWarn, "could not encode a journal entry",
			"journal_id", entry.JournalId, "error", err.Error())
		return
	}

	msg := &natscore.Msg{
		Subject: model.SubjectSystemJournal,
		Data:    raw,
		Header: natscore.Header{
			model.HeaderFormat: []string{"json"},
			model.HeaderEvent:  []string{model.EventJournalCreated},
		},
	}

	if err := j.nats.NC.PublishMsg(msg); err != nil {
		j.log.Log(logger.TypeNet, logger.CodeWarn, "could not announce a journal entry",
			"journal_id", entry.JournalId, "error", err.Error())
	}
}
