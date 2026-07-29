// Package seed applies the starting rows a fresh install needs.
//
// Each subject gets a file here: the values as a plain Go struct, next to the
// function that writes them. Every seed decides for itself whether it is
// needed, so running them again is a no op.
package seed

import (
	"context"

	"hstserver/config"
	"hstserver/pkg/crypto"
	"hstserver/pkg/db"
	"hstserver/pkg/logger"
)

type Seeder struct {
	DB     *db.PostgresDB
	Hasher *crypto.Hasher
	Log    *logger.Logger
	Cfg    *config.Config
}

func New(database *db.PostgresDB, hasher *crypto.Hasher, log *logger.Logger, cfg *config.Config) *Seeder {
	return &Seeder{DB: database, Hasher: hasher, Log: log, Cfg: cfg}
}

// Seed applies every seed in order. To add one: write its file in this folder
// with the values and a Seed function, then call it here.
func (s *Seeder) Seed(ctx context.Context) error {
	if err := s.SeedManager(ctx); err != nil {
		return err
	}

	return nil
}
