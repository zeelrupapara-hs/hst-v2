package oauth2

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"
)

// BootstrapFile is read once, only when no manager exists yet.
const BootstrapFile = "config/bootstrap.json"

// placeholderPassword is what ships in the example file. Booting with it on a
// real host would publish an admin account to the internet.
const placeholderPassword = "change-me-before-first-boot"

// bootstrapConfig is the first administrator.
type bootstrapConfig struct {
	Group    string `json:"group"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Bootstrap creates the first admin when hst.managers is empty, so a fresh
// install has a way in. It is a no op on every later start.
func (o *OAuth2) Bootstrap(ctx context.Context) error {
	var managers int
	if err := o.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.managers`).Scan(&managers); err != nil {
		return err
	}
	if managers > 0 {
		return nil
	}

	raw, err := os.ReadFile(BootstrapFile)
	if errors.Is(err, os.ErrNotExist) {
		o.Log.Log(logger.TypeSys, logger.CodeWarn,
			"no manager exists and no bootstrap file was found",
			"file", BootstrapFile)
		return nil
	}
	if err != nil {
		return err
	}

	var cfg bootstrapConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("invalid %s: %w", BootstrapFile, err)
	}

	if cfg.Password == "" || cfg.Password == placeholderPassword {
		return fmt.Errorf("%s still holds the placeholder password", BootstrapFile)
	}
	if cfg.Group == "" || cfg.Name == "" {
		return fmt.Errorf("%s needs group and name", BootstrapFile)
	}

	hash, err := o.Hasher.HashPassword(cfg.Password)
	if err != nil {
		return err
	}

	tx, err := o.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	// reset_pass forces the first login through a password change, so the
	// value sitting in the file stops being a working credential
	rights := model.UsersRights_enabled |
		model.UsersRights_password |
		model.UsersRights_reset_pass

	var login int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.users
		   ("group", rights, name, email, password_main, registration, last_pass_change, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$6,$6)
		 RETURNING login`,
		cfg.Group, int64(rights), cfg.Name, cfg.Email, hash, now).Scan(&login); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.accounts (login, currency_digits, margin_leverage, updated_at)
		 VALUES ($1, 2, 100, $2)`, login, now); err != nil {
		return err
	}

	// admin and manager are the two terminal gates; the rest are granted
	// through the panel once someone is logged in
	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.managers (login, name, groups, right_admin, right_manager,
		                           right_cfg_managers, updated_at)
		 VALUES ($1, $2, $3, 1, 1, 1, $4)`,
		login, cfg.Name, []string{cfg.Group}, now); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	o.Log.Log(logger.TypeSys, logger.CodeLogin, "bootstrap administrator created",
		"login", login, "group", cfg.Group)

	return nil
}
