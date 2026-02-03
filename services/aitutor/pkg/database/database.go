package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/aitutormux"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

type PostgresConfig struct {
	Username     string
	Password     string
	Host         string
	Port         string
	DatabaseName string
}

func NewPostgresDB(ctx context.Context, config *PostgresConfig) (aitutormux.Database, error) {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DatabaseName)
	pool, err := pgxpool.Connect(ctx, url)
	if err != nil {
		return nil, err
	}

	return &PostgresDB{
		Pool: pool,
	}, nil
}

func (pg *PostgresDB) GetAgent(ctx context.Context, agentId *aitutormux.AgentId) (*aitutormux.Agent, error) {
	row := pg.Pool.QueryRow(ctx, `
	SELECT name
	FROM agents
	WHERE agent_id = $1`, agentId.AgentId)

	var name string
	if err := row.Scan(&name); err != nil {
		return nil, err

	}

	return &aitutormux.Agent{
		AgentId: *agentId,
		Name:    name,
	}, nil
}
