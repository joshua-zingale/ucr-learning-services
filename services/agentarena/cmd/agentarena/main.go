package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentclass"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentmux"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentregistry"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/auth"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/database"

	"github.com/joshua-zingale/ucr-learning-services/internal/confenvflag/pkg/confenvflag"
)

const _ENV_VAR_PREFIX = "AGENTARENA_"

func main() {
	flag := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		host = flag.String("host", "127.0.0.1", "the host at which the web server is broadcast.")
		port = flag.String("port", "46307", "the port on which the web server is broadcast.")

		pg_username = flag.String("pg-username", "", "the username of postgres user.")
		pg_password = flag.String("pg-password", "", "the password for postgres user.")
		pg_host     = flag.String("pg-host", "127.0.0.1", "the host at which postgres in broadcast.")
		pg_port     = flag.String("pg-port", "5432", "the port on which postgres in broadcast.")
		pg_database = flag.String("pg-database", "", "the name of the postgres databse.")
	)

	if err := confenvflag.Parse(flag, _ENV_VAR_PREFIX, os.Args[1:]); err != nil {
		log.Fatal(err.Error())
	}

	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		*pg_username,
		*pg_password,
		*pg_host,
		*pg_port,
		*pg_database))
	if err != nil {
		log.Fatalf("Failed to establish connection to database: %s", err.Error())
	}

	db := *database.New(pool)

	agentClassDriverRegistry, err := agentregistry.New(ctx,
		map[string]agentmux.AgentClassDriver{
			"ollama": &agentclass.OllamaAgentDriver{
				Url: "http://localhost:11434",
			},
		})
	if err != nil {
		log.Fatalf("Failed to initialize agent class drivers: %s", err.Error())
	}

	handler := agentmux.NewAiTutorMux(&agentmux.AgentArenaConfig{
		Db:                       db,
		Auth:                     auth.TaxisAuth{},
		AgentClassDriverRegistry: agentmux.AgentClassDriverRegistry(agentClassDriverRegistry),
	})

	url := fmt.Sprintf("%s:%s", *host, *port)

	log.Printf("Now listening at %s", url)
	http.ListenAndServe(url, handler)
}
