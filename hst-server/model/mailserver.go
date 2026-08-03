package model

// MailServer is one SMTP account the platform sends from. It is not hst.mails, which is the
// internal mailbox between a trading account and support.
type MailServer struct {
	MailServerId int64  `db:"mail_server_id" json:"mail_server_id"`
	Enabled      bool   `db:"enabled" json:"enabled"`
	Name         string `db:"name" json:"name"`
	SenderEmail  string `db:"sender_email" json:"sender_email"`
	SenderName   string `db:"sender_name" json:"sender_name"`
	SmtpServer   string `db:"smtp_server" json:"smtp_server"`
	SmtpLogin    string `db:"smtp_login" json:"smtp_login"`

	// never serialised: a password leaves the process only on the way to the mail server
	SmtpPassword string `db:"smtp_password" json:"-"`

	IsDefault bool `db:"is_default" json:"is_default"`

	TotalSent   int64 `db:"total_sent" json:"total_sent"`
	TotalErrors int64 `db:"total_errors" json:"total_errors"`
	TimeMinMs   int64 `db:"time_min_ms" json:"time_min_ms"`
	TimeMaxMs   int64 `db:"time_max_ms" json:"time_max_ms"`
	TimeSumMs   int64 `db:"time_sum_ms" json:"-"`

	CreatedAt int64 `db:"created_at" json:"created_at"`
	UpdatedAt int64 `db:"updated_at" json:"updated_at"`
}

func (MailServer) TableName() string { return "hst.mail_servers" }

// TimeAvgMs is the average send time the manager list shows.
func (m MailServer) TimeAvgMs() int64 {
	if m.TotalSent == 0 {
		return 0
	}
	return m.TimeSumMs / m.TotalSent
}

// OutboxState is where an outgoing mail is in its life.
type OutboxState int32

const (
	OutboxState_queued OutboxState = 0
	OutboxState_sent   OutboxState = 1
	OutboxState_failed OutboxState = 2
)

// Enum value maps for OutboxState.
var (
	OutboxState_name = map[int32]string{
		0: "queued",
		1: "sent",
		2: "failed",
	}
	OutboxState_value = valuesOf(OutboxState_name)
)

// Outbox is one mail waiting to go out. A welcome mail carries the only copy of a generated
// password, so it is queued and retried rather than sent inline and lost.
type Outbox struct {
	OutboxId     int64       `db:"outbox_id" json:"outbox_id"`
	MailServerId int64       `db:"mail_server_id" json:"mail_server_id"`
	Recipient    string      `db:"recipient" json:"recipient"`
	Subject      string      `db:"subject" json:"subject"`
	Body         string      `db:"body" json:"body"`
	State        OutboxState `db:"state" json:"state"`
	Attempts     int32       `db:"attempts" json:"attempts"`
	LastError    string      `db:"last_error" json:"last_error"`
	CreatedAt    int64       `db:"created_at" json:"created_at"`
	SentAt       int64       `db:"sent_at" json:"sent_at"`
}

func (Outbox) TableName() string { return "hst.outbox" }

// OutboxMaxAttempts is where retrying stops and the row is marked failed.
const OutboxMaxAttempts int32 = 5
