package postgres

import (
	"context"
	"fmt"
	"strings"

	"sync-algo/internal/config"
	"sync-algo/internal/models"
	"sync-algo/internal/storage"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(cfg *config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db := stdlib.OpenDB(*pool.Config().ConnConfig)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) CreateClient(ctx context.Context, clientInfo *models.Client) (*models.Client, error) {
	const op = "storage.postgres.CreateClient"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // Re-throw panic after Rollback
		}
	}()

	query := `
        INSERT INTO clients (name, version, image, cpu, memory, priority, need_restart, spawned_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, name, version, image, cpu, memory, priority, need_restart, spawned_at, created_at, updated_at
    `

	row := tx.QueryRow(ctx, query,
		clientInfo.ClientName,
		clientInfo.Version,
		clientInfo.Image,
		clientInfo.CPU,
		clientInfo.Memory,
		clientInfo.Priority,
		clientInfo.NeedRestart,
		clientInfo.SpawnedAt,
		clientInfo.CreatedAt,
		clientInfo.UpdatedAt,
	)

	var client models.Client
	err = row.Scan(&client.ID, &client.ClientName, &client.Version, &client.Image, &client.CPU, &client.Memory, &client.Priority, &client.NeedRestart, &client.SpawnedAt, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		defer tx.Rollback(ctx)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO algorithm_statuses (client_id) VALUES ($1)", client.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &client, nil
}

func (s *Storage) UpdateClient(ctx context.Context, fields []string, values []interface{}) (*models.Client, error) {
	const op = "storage.postgres.UpdateClient"

	f := strings.Join(fields[:], ", ")

	q := fmt.Sprintf("UPDATE clients SET %s WHERE id = $%d RETURNING id, name, version, image, cpu, memory, priority, need_restart, spawned_at, created_at, updated_at", f, len(values))

	row := s.pool.QueryRow(ctx, q, values...)

	var client models.Client
	err := row.Scan(&client.ID, &client.ClientName, &client.Version, &client.Image, &client.CPU, &client.Memory, &client.Priority, &client.NeedRestart, &client.SpawnedAt, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &client, nil
}

func (s *Storage) RemoveClient(ctx context.Context, id int) error {
	const op = "storage.postgres.RemoveClient"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if p := recover(); p != nil {
			defer tx.Rollback(ctx)
			panic(p) // Re-throw panic after Rollback
		}
	}()

	// Удаление из algorithm_statuses
	_, err = tx.Exec(ctx, `DELETE FROM algorithm_statuses WHERE client_id = $1`, id)
	if err != nil {
		defer tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	// Удаление из clients
	ct, err := tx.Exec(ctx, `DELETE FROM clients WHERE id = $1`, id)
	if err != nil {
		defer tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	if ct.RowsAffected() == 0 {
		defer tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateStatuses(ctx context.Context, fields []string, values []interface{}) (*models.AlgoStatuses, error) {
	const op = "storage.postgres.UpdateStatuses"

	if len(fields) == 0 {
		return nil, fmt.Errorf("%s: %w", op)
	}

	f := strings.Join(fields, ", ")

	q := fmt.Sprintf("UPDATE algorithm_statuses SET %s WHERE id = $%d RETURNING client_id, vwap, twap, hft", f, len(values))

	row := s.pool.QueryRow(ctx, q, values...)

	var algoStatuses models.AlgoStatuses
	err := row.Scan(&algoStatuses.ClientID, &algoStatuses.VWAP, &algoStatuses.TWAP, &algoStatuses.HFT)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &algoStatuses, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
