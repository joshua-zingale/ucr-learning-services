package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/aitutormux"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/auth"
	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/database"
)

func main() {
	ctx := context.Background()
	db, err := database.NewPostgresDB(ctx, &database.PostgresConfig{
		Username:     "agentwombat",
		Password:     "",
		Host:         "localhost",
		Port:         "5432",
		DatabaseName: "aitutor_go",
	})
	if err != nil {
		log.Fatalf("Failed to establish connection to database: %s", err.Error())
	}

	handler := aitutormux.NewAiTutorMux(&aitutormux.AiTutorConfig{
		Db:   db,
		Auth: auth.TaxisAuth{},
	})

	http.ListenAndServe("localhost:7654", handler)
}
