// Package seed applies the starting rows a fresh install needs.
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

func (s *Seeder) Seed(ctx context.Context) error {
	if err := s.SeedManager(ctx); err != nil {
		return err
	}

	return nil
}
