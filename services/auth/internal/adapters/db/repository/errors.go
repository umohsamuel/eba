package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres SQLSTATE codes. See https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgForeignKeyViolation = "23503"
	pgUniqueViolation     = "23505"
)

func isUniqueViolation(err error) bool {
	return hasPgCode(err, pgUniqueViolation)
}

func isForeignKeyViolation(err error) bool {
	return hasPgCode(err, pgForeignKeyViolation)
}

func hasPgCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
