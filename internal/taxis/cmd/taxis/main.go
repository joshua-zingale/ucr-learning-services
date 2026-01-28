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

	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/database"
	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/filewatch"
	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/flagset"
	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/web"
)

const envVarPrefix = "TAXIS_"

var rootCommandName = os.Args[0]

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
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Please specify a sub command:")
		for subcommand := range subCommands {
			fmt.Fprintf(os.Stderr, "    - %s\n", subcommand)
		}
	}
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
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
		os.Exit(1)
	}

}

func serve(execution executionContext) {
	flag := flag.NewFlagSet(execution.CommandName(), flag.ExitOnError)
	flag.PrintDefaults()
	envFlag := flagset.NewEnvironmentDefaultFlagSet(flag, envVarPrefix)
	var (
		host                      = envFlag.String("host", "127.0.0.1", "the host at which the web server is broadcast.")
		port                      = envFlag.String("port", "14812", "the port on which the web server is broadcast.")
		rootPath                  = envFlag.String("root-path", "/taxis", "the root URI path for this web server.")
		groupsHeaderName          = envFlag.String("groups-header", "X-Groups", "the header to which the assigned groups will be written.")
		userIdHeaderName          = envFlag.String("user-id-header", "X-Email", "the header that will be searched for the userId.")
		watchGroupsFile           = envFlag.Bool("watch", false, "If specified, the GROUPS_FILE.yml file is watched, so that any update to the file triggers the database to reload.")
		printEnvironmentVariables = flag.Bool("print-supported-environment-variables", false, "If specified, prints all supported environment variables along with usages.")
	)
	flag.Parse(execution.args)

	if *printEnvironmentVariables {
		envFlag.PrintSupportedEnvironmentVariables()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Invalid usage: must be\n    %s [-flag [value]...] GROUPS_FILE.yml\n", execution.CommandName())
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
		RootPath:         *rootPath,
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
