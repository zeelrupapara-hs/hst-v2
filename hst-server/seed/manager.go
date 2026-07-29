package seed

import (
	"context"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"
)

// firstManager is the administrator a fresh install starts with.
var firstManager = struct {
	Group string
	Name  string
	Email string
}{
	Group: `managers\admin`,
	Name:  "First Admin",
	Email: "admin@hybridsolutions.com",
}

// SeedManager creates the first administrator on an empty database.
func (s *Seeder) SeedManager(ctx context.Context) error {
	var managers int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.managers`).Scan(&managers); err != nil {
		return err
	}
	if managers > 0 {
		return nil
	}

	password := s.Cfg.Auth.FirstManagerPassword
	if password == "" {
		s.Log.Log(logger.TypeSys, logger.CodeWarn,
			"no manager exists and no seed password was given",
			"env", "FIRST_MANAGER_PASSWORD")
		return nil
	}

	hash, err := s.Hasher.HashPassword(password)
	if err != nil {
		return err
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	// reset_pass forces a password change, so the env value stops working after first use.
	rights := model.UsersRights_enabled |
		model.UsersRights_password |
		model.UsersRights_reset_pass

	var login int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.users
		   ("group", rights, name, email, password_main, registration, last_pass_change, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$6,$6)
		 RETURNING login`,
		firstManager.Group, int64(rights), firstManager.Name, firstManager.Email, hash, now).Scan(&login); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.accounts (login, currency_digits, margin_leverage, updated_at)
		 VALUES ($1, 2, 100, $2)`, login, now); err != nil {
		return err
	}

	// admin and manager are the terminal gates; the rest are granted in the panel.
	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.managers (login, name, groups, right_admin, right_manager,
		                           right_cfg_managers, updated_at)
		 VALUES ($1, $2, $3, 1, 1, 1, $4)`,
		login, firstManager.Name, []string{firstManager.Group}, now); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.Log.Log(logger.TypeSys, logger.CodeLogin, "seed administrator created",
		"login", login, "group", firstManager.Group)

	return nil
}
