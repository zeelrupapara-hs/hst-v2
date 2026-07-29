package seed

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"
)

// managersFile describes the first administrator. It carries no secret, so it
// is committed; the password comes from SEED_MANAGER_PASSWORD.
var managersFile = filepath.Join(Dir, "managers.json")

// managerSeed is the shape of that file.
type managerSeed struct {
	Group string `json:"group"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// managers creates the first administrator so a fresh install has a way in.
// It runs only while hst.managers is empty, so it is a no op on every later
// start and can never mint a second admin.
func (s *Seeder) managers(ctx context.Context) error {
	var managers int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.managers`).Scan(&managers); err != nil {
		return err
	}
	if managers > 0 {
		return nil
	}

	password := s.Cfg.Auth.SeedManagerPassword
	if password == "" {
		s.Log.Log(logger.TypeSys, logger.CodeWarn,
			"no manager exists and no seed password was given",
			"env", "SEED_MANAGER_PASSWORD")
		return nil
	}

	raw, err := os.ReadFile(managersFile)
	if errors.Is(err, os.ErrNotExist) {
		s.Log.Log(logger.TypeSys, logger.CodeWarn,
			"no manager exists and the seed file is missing", "file", managersFile)
		return nil
	}
	if err != nil {
		return err
	}

	var m managerSeed
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("invalid %s: %w", managersFile, err)
	}
	if m.Group == "" || m.Name == "" {
		return fmt.Errorf("%s needs group and name", managersFile)
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

	// reset_pass forces the first login through a password change, so the value
	// in the environment stops being a working credential once it is used
	rights := model.UsersRights_enabled |
		model.UsersRights_password |
		model.UsersRights_reset_pass

	var login int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.users
		   ("group", rights, name, email, password_main, registration, last_pass_change, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$6,$6)
		 RETURNING login`,
		m.Group, int64(rights), m.Name, m.Email, hash, now).Scan(&login); err != nil {
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
		login, m.Name, []string{m.Group}, now); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.Log.Log(logger.TypeSys, logger.CodeLogin, "seed administrator created",
		"login", login, "group", m.Group)

	return nil
}
