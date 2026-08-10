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

// Entry stores a journal line and commits it only once the broker holds it.
func (j *Journal) Entry(ctx context.Context, entry *model.Journal) error {
	if entry == nil {
		return nil
	}

	tx, err := j.db.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	entry.CreatedAt = time.Now().UnixNano()
	if len(entry.Detail) == 0 {
		entry.Detail = json.RawMessage("{}")
	}

	// an empty string is not an address, so the inet column takes null instead
	var ip *string
	if entry.Ip != "" {
		ip = &entry.Ip
	}

	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.journal (created_at, type, code, login, ip, channel, os, message, detail)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING journal_id`,
		entry.CreatedAt, entry.Type, entry.Code, entry.Login, ip,
		channelOr(entry.Channel), entry.Os, entry.Message, entry.Detail,
	).Scan(&entry.JournalId); err != nil {
		return err
	}

	// announced while the row is uncommitted, so a broker that refuses it takes the row down too
	if err := j.publishMsg(entry); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// wireEntry is a journal line as its owner sees it, matching what the API returns so a line reads
// the same whether it arrived by request or was announced.
type wireEntry struct {
	JournalId int64  `json:"journal_id"`
	CreatedAt int64  `json:"created_at"`
	Type      int32  `json:"type"`
	Code      int32  `json:"code"`
	Login     int64  `json:"login"`
	Channel   string `json:"channel"`
	Os        string `json:"os"`
	Ip        string `json:"ip"`
	Message   string `json:"message"`
}

// channelOr keeps the column honest: anything outside the known terminals is the platform itself.
func channelOr(c string) string {
	switch c {
	case "web", "desktop", "mobile", "api", "system":
		return c
	}

	return "system"
}

// publishMsg announces an entry and reports whether the broker actually holds it.
func (j *Journal) publishMsg(entry *model.Journal) error {
	journal, err := json.Marshal(&wireEntry{
		JournalId: entry.JournalId,
		CreatedAt: entry.CreatedAt,
		Type:      entry.Type,
		Code:      entry.Code,
		Login:     entry.Login,
		Channel:   channelOr(entry.Channel),
		Os:        entry.Os,
		Ip:        entry.Ip,
		Message:   entry.Message,
	})
	if err != nil {
		return err
	}

	msg := &natscore.Msg{
		Subject: model.SubjectJournal(entry.Login),
		Data:    journal,
		Header: natscore.Header{
			model.HeaderFormat: []string{"json"},
			model.HeaderEvent:  []string{model.EventJournal},
		},
	}

	if err := j.nats.NC.PublishMsg(msg); err != nil {
		return err
	}

	// a publish alone only reaches the write buffer; the flush round trips to the server
	return j.nats.NC.Flush()
}
