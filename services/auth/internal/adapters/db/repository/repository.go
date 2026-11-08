package repository

import (
	"database/sql"

	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/sqlc"
	"github.com/umohsamuel/eba/services/auth/internal/domain/refreshtoken"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
	"github.com/umohsamuel/eba/services/auth/internal/domain/usertokens"
)

type AdaptersRepository struct {
	Users     user.UserRepository
	Refresh   refreshtoken.RefreshTokenRepository
	UserToken usertokens.UserTokensRepository
}

type Config struct {
	DB *sql.DB
}

func New(deps Config) *AdaptersRepository {
	queries := sqlc.New(deps.DB)

	return &AdaptersRepository{
		Users:     NewUserRepositoryAdapter(deps.DB, queries),
		Refresh:   NewRefreshRepositoryAdapter(deps.DB, queries),
		UserToken: NewUserTokensRepositoryAdapter(deps.DB, queries),
	}
}
