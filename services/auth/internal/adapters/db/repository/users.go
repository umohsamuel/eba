package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/sqlc"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
)

type UserRepositoryAdapter struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewUserRepositoryAdapter(db *sql.DB, queries *sqlc.Queries) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{
		db:      db,
		queries: queries,
	}
}

var _ user.UserRepository = (*UserRepositoryAdapter)(nil)

func (r *UserRepositoryAdapter) List(ctx context.Context, filter user.UserFilter) ([]*user.User, error) {
	dbUsers, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	users := make([]*user.User, 0, len(dbUsers))

	for _, dbUser := range dbUsers {
		users = append(users, &user.User{
			ID:    dbUser.ID,
			Email: dbUser.Email,
		})
	}

	return users, nil
}

func (r *UserRepositoryAdapter) Create(ctx context.Context, u *user.User) (*user.User, error) {
	params := sqlc.CreateUserParams{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	}

	row, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainerrors.ErrEmailAlreadyExists.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	u.ID = row.ID

	return u, nil
}

func (r *UserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &user.User{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
	}, nil
}

func (r *UserRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	dbUser, err := r.queries.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &user.User{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
	}, nil
}

func (r *UserRepositoryAdapter) Update(ctx context.Context, _user *user.User) (*user.User, error) {

	existingUser, err := r.queries.GetUser(ctx, _user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	if _user.Email != "" {
		existingUser.Email = _user.Email
	}

	if _user.PasswordHash != "" {
		existingUser.PasswordHash = _user.PasswordHash
	}

	updatedUser, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:           existingUser.ID,
		Email:        existingUser.Email,
		PasswordHash: existingUser.PasswordHash,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound.Wrap(err)
		}
		if isUniqueViolation(err) {
			return nil, domainerrors.ErrEmailAlreadyExists.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}

	return &user.User{
		ID:           updatedUser.ID,
		Email:        updatedUser.Email,
		PasswordHash: updatedUser.PasswordHash,
	}, nil
}

func (r *UserRepositoryAdapter) Delete(ctx context.Context, _id uuid.UUID) (*uuid.UUID, error) {
	id, err := r.queries.DeleteUser(ctx, _id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound.Wrap(err)
		}
		return nil, domainerrors.ErrInternal.Wrap(err)
	}
	return &id, nil
}
