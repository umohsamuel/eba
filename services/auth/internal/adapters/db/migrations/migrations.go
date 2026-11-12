package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
)

//go:embed *.sql
var embedded embed.FS

type Result struct {
	Version    int64  `json:"version"`
	Path       string `json:"path"`
	Direction  string `json:"direction"`
	DurationMS int64  `json:"duration_ms"`
	Empty      bool   `json:"empty"`
}

type Status struct {
	Version   int64      `json:"version"`
	Path      string     `json:"path"`
	State     string     `json:"state"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
}

type Migrator struct {
	provider *goose.Provider
}

func New(db *sql.DB) (*Migrator, error) {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("create migration locker: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, embedded, goose.WithSessionLocker(locker))
	if err != nil {
		return nil, fmt.Errorf("create migration provider: %w", err)
	}

	return &Migrator{provider: provider}, nil
}

func (m *Migrator) Status(ctx context.Context) ([]Status, error) {
	statuses, err := m.provider.Status(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	out := make([]Status, 0, len(statuses))
	for _, s := range statuses {
		st := Status{Version: s.Source.Version, Path: s.Source.Path, State: string(s.State)}
		if !s.AppliedAt.IsZero() {
			appliedAt := s.AppliedAt
			st.AppliedAt = &appliedAt
		}
		out = append(out, st)
	}

	return out, nil
}

func (m *Migrator) Version(ctx context.Context) (int64, error) {
	version, err := m.provider.GetDBVersion(ctx)
	if err != nil {
		return 0, mapError(err)
	}

	return version, nil
}

func (m *Migrator) Up(ctx context.Context) ([]Result, error) {
	results, err := m.provider.Up(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	return toResults(results...), nil
}

func (m *Migrator) UpTo(ctx context.Context, version int64) ([]Result, error) {
	results, err := m.provider.UpTo(ctx, version)
	if err != nil {
		return nil, mapError(err)
	}

	return toResults(results...), nil
}

func (m *Migrator) DownTo(ctx context.Context, version int64) ([]Result, error) {
	results, err := m.provider.DownTo(ctx, version)
	if err != nil {
		return nil, mapError(err)
	}

	return toResults(results...), nil
}

func toResults(results ...*goose.MigrationResult) []Result {
	out := make([]Result, 0, len(results))
	for _, r := range results {
		if r == nil {
			continue
		}
		out = append(out, Result{
			Version:    r.Source.Version,
			Path:       r.Source.Path,
			Direction:  r.Direction,
			DurationMS: r.Duration.Milliseconds(),
			Empty:      r.Empty,
		})
	}

	return out
}

func mapError(err error) error {
	switch {
	case errors.Is(err, goose.ErrNoNextVersion), errors.Is(err, goose.ErrNoCurrentVersion):
		return domainerrors.ErrNoMigrationToApply.Wrap(err)
	case errors.Is(err, goose.ErrVersionNotFound):
		return domainerrors.ErrMigrationVersionNotFound.Wrap(err)
	default:
		return domainerrors.ErrInternal.Wrap(err)
	}
}
