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

func (pg *PostgresDB) GetAgent(ctx context.Context, agentId aitutormux.AgentId) (*aitutormux.Agent, error) {
	row := pg.Pool.QueryRow(ctx, `
	SELECT name
	FROM agents
	WHERE agent_id = $1`, agentId)

	var name string
	if err := row.Scan(&name); err != nil {
		return nil, err

	}

	return &aitutormux.Agent{
		AgentId: agentId,
		Name:    name,
	}, nil
}

func (pg *PostgresDB) GetAgentConfig(ctx context.Context, agentId aitutormux.AgentId) (*aitutormux.AgentConfig, error) {
	row := pg.Pool.QueryRow(ctx, `
	SELECT system_prompt
	FROM agent_configs
	WHERE agent_id = $1`, agentId)

	var systemPrompt string
	if err := row.Scan(&systemPrompt); err != nil {
		return nil, err

	}

	return &aitutormux.AgentConfig{
		AgentId:      agentId,
		SystemPrompt: systemPrompt,
	}, nil
}
