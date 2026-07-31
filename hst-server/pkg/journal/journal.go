// Package journal records what happened on the server: who did it, to which
// record, and when.
//
// The file log in logs/YYYYMMDD.log keeps the same entries for an operator
// reading a terminal. This table is the one the back office queries and
// filters, so it is written deliberately rather than scraped back out of text.
package journal

import (
	"context"
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

func (j *Journal) Entry(ctx context.Context, entry *model.Journal) error {
	if entry == nil {
		return nil
	}

	entry.CreatedAt = time.Now().UnixNano()
	if len(entry.Detail) == 0 {
		entry.Detail = json.RawMessage("{}")
	}

	// an empty string is not an address, so it is stored as null rather than
	// refused by the inet column
	var ip *string
	if entry.Ip != "" {
		ip = &entry.Ip
	}

	if err := j.db.DB.QueryRow(ctx,
		`INSERT INTO hst.journal (created_at, type, code, login, ip, message, detail)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING journal_id`,
		entry.CreatedAt, entry.Type, entry.Code, entry.Login, ip, entry.Message, entry.Detail,
	).Scan(&entry.JournalId); err != nil {
		return err
	}

	j.publishMsg(entry)

	return nil
}

// publishMsg announces a stored entry to the managers watching the journal.
func (j *Journal) publishMsg(entry *model.Journal) {
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
