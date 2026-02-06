package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
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

	handler := aitutormux.NewAiTutorMux(&aitutormux.AiTutorConfig{
		Db:   *database.New(pool),
		Auth: auth.TaxisAuth{},
	})

	http.ListenAndServe("localhost:7654", handler)
}
