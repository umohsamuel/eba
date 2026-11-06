package user

import (
	"context"

	"github.com/google/uuid"
)

type UserFilter struct {
	Limit  int
	Offset int
}

type UserRepository interface {
	List(ctx context.Context, filter UserFilter) ([]*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) (*uuid.UUID, error)
}
