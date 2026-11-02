package repository

import (
	"database/sql"

	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/sqlc"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
)

type AdaptersRepository struct {
	Users user.UserRepository
}

type Config struct {
	DB *sql.DB
}

func New(deps Config) *AdaptersRepository {
	queries := sqlc.New(deps.DB)

	return &AdaptersRepository{
		Users: NewUserRepositoryAdapter(deps.DB, queries),
	}
}
