// Package seed applies the starting rows a fresh install needs.
//
// The data lives in the seed directory at the repo root, one json file per
// subject. This package only reads those files and writes the rows. Every seed
// decides for itself whether it is needed, so running them again is a no op.
package seed

import (
	"context"

	"hstserver/config"
	"hstserver/pkg/crypto"
	"hstserver/pkg/db"
	"hstserver/pkg/logger"
)

// Dir holds the seed files, relative to the working directory.
const Dir = "seed"

type Seeder struct {
	DB     *db.PostgresDB
	Hasher *crypto.Hasher
	Log    *logger.Logger
	Cfg    *config.Config
}

func New(database *db.PostgresDB, hasher *crypto.Hasher, log *logger.Logger, cfg *config.Config) *Seeder {
	return &Seeder{DB: database, Hasher: hasher, Log: log, Cfg: cfg}
}

// Run applies every seed in order. To add one, drop its json in seed/, write
// the apply function beside this file, and add it to the list.
func (s *Seeder) Run(ctx context.Context) error {
	seeds := []struct {
		name  string
		apply func(context.Context) error
	}{
		{"managers", s.managers},
	}

	for _, seed := range seeds {
		if err := seed.apply(ctx); err != nil {
			return err
		}
	}

	return nil
}
