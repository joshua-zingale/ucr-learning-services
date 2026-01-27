package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/database"
	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/filewatch"
	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/web"
)

const rootCommandName = "taxis"

var subCommands = map[string]func(executionContext){
	"serve": serve,
}

type executionContext struct {
	name string
	args []string
}

func (ec *executionContext) CommandName() string {
	return strings.Join([]string{rootCommandName, ec.name}, " ")
}

func main() {

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify a sub command:")

		for subcommand, _ := range subCommands {
			fmt.Fprintf(os.Stderr, "    - %s\n", subcommand)
		}
		os.Exit(1)
	}

	cmd, args := args[0], args[1:]

	if routine, ok := subCommands[cmd]; ok {
		routine(executionContext{
			name: cmd,
			args: args,
		})
	} else {
		fmt.Fprintf(os.Stderr, "Invalid sub command '%s'\n", cmd)
	}

}

func serve(execution executionContext) {
	flag := flag.NewFlagSet(execution.CommandName(), flag.ExitOnError)
	flag.PrintDefaults()
	var (
		host             = flag.String("host", "127.0.0.1", "the host at which the web server is broadcast.")
		port             = flag.String("port", "14812", "the host at which the web server is broadcast.")
		groupsHeaderName = flag.String("groups-header", "X-Groups", "the header to which the assigned groups will be written.")
		userIdHeaderName = flag.String("user-id-header", "X-Email", "the header that will be searched for the userId.")
		watchGroupsFile  = flag.Bool("watch", false, "If specified, the GROUPS_FILE.yml file is watched, so that any update to the file triggers the database to reload.")
	)
	flag.Parse(execution.args)

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Invalid usage: must be\n    %s GROUPS_FILE.yml", execution.CommandName())
		os.Exit(1)
	}
	groupsFilePath := args[0]

	yamlSource, err := os.ReadFile(groupsFilePath)
	if err != nil {
		log.Fatalf("Error: Could not %s\n", err.Error())
	}

	db, err := database.GetGroupDBFromYaml(string(yamlSource))
	if err != nil {
		log.Fatalf("Error initializing database: %s\n", err.Error())
	}

	mux, err := web.NewTaxisMux(&web.TaxisConfig{
		Database:         db,
		GroupsHeaderName: *groupsHeaderName,
		UserIdHeaderName: *userIdHeaderName,
	})
	if err != nil {
		log.Fatalf("Invalid config: %s", err.Error())
	}

	address := fmt.Sprintf("%s:%s", *host, *port)

	ctx := context.Background()

	if *watchGroupsFile {
		go filewatch.RunOnUpdate(ctx, groupsFilePath, 500*time.Millisecond, func() {

			yamlSource, err := os.ReadFile(groupsFilePath)
			if err != nil {
				log.Printf("Error reloading %s: %s", groupsFilePath, err.Error())
			}
			if err := db.LoadGroupsFromYaml(string(yamlSource)); err != nil {
				log.Printf("Error reloading %s: %s", groupsFilePath, err.Error())
			} else {
				log.Printf("Reloaded %s", groupsFilePath)
			}
		})
	}

	log.Printf("Starting server at %s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("%s", err.Error())
	}

	log.Println("Server Stopped")
}
