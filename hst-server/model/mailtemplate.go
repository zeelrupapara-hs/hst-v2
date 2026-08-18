package model

// MailTemplate is one manager's saved compose template. Saving under a taken name overwrites,
// so a manager's templates behave like named slots rather than a growing archive.
type MailTemplate struct {
	TemplateId   int64  `db:"template_id" json:"template_id"`
	ManagerLogin int64  `db:"manager_login" json:"manager_login"`
	Name         string `db:"name" json:"name"`
	Subject      string `db:"subject" json:"subject"`
	Body         string `db:"body" json:"body"`
	CreatedAt    int64  `db:"created_at" json:"created_at"`
	UpdatedAt    int64  `db:"updated_at" json:"updated_at"`
}

func (MailTemplate) TableName() string { return "hst.mail_templates" }
