package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentclass"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentmux"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentregistry"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/auth"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/database"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		"agentwombat",
		"",
		"localhost",
		"5432",
		"agentarena"))
	if err != nil {
		log.Fatalf("Failed to establish connection to database: %s", err.Error())
	}

	db := *database.New(pool)

	agentClassDriverRegistry, err := agentregistry.New(ctx,
		map[string]agentmux.AgentClassDriver{
			"standard": &agentclass.StandardAgentDriver{},
		},
		db.GetAgentClassIdFromSlug)
	if err != nil {
		log.Fatalf("Failed to initialize agent class drivers: %s", err.Error())
	}

	handler := agentmux.NewAiTutorMux(&agentmux.AgentArenaConfig{
		Db:                       db,
		Auth:                     auth.TaxisAuth{},
		AgentClassDriverRegistry: agentmux.AgentClassDriverRegistry(agentClassDriverRegistry),
	})

	http.ListenAndServe("localhost:7654", handler)
}
