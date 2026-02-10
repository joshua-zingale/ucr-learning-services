package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/agentclass"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/agentregistry"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/aitutormux"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/auth"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/database"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		"agentwombat",
		"",
		"localhost",
		"5432",
		"aitutor_go"))
	if err != nil {
		log.Fatalf("Failed to establish connection to database: %s", err.Error())
	}

	db := *database.New(pool)

	agentClassDriverRegistry, err := agentregistry.New(ctx,
		map[string]aitutormux.AgentClassDriver{
			"standard": &agentclass.StandardAgentDriver{},
		},
		db.GetAgentClassIdFromSlug)
	if err != nil {
		log.Fatalf("Failed to initialize agent class drivers: %s", err.Error())
	}

	handler := aitutormux.NewAiTutorMux(&aitutormux.AiTutorConfig{
		Db:                       db,
		Auth:                     auth.TaxisAuth{},
		AgentClassDriverRegistry: aitutormux.AgentClassDriverRegistry(agentClassDriverRegistry),
	})

	http.ListenAndServe("localhost:7654", handler)
}
