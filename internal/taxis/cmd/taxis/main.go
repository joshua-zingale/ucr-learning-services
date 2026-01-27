package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/database"
	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/web"
)

const commandName = "taxis"

var subCommands = map[string]func([]string){
	"serve": serve,
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
		routine(args)
	} else {
		fmt.Fprintf(os.Stderr, "Invalid sub command '%s'\n", cmd)
	}

}

func serve(args []string) {
	flag := flag.NewFlagSet(strings.Join([]string{commandName, "serve"}, " "), flag.ExitOnError)
	flag.PrintDefaults()
	var (
		groupsFilePath = flag.String("groups", "groups.yml", "the YAML file containing the group assignments.")
	)
	flag.Parse(args)

	yamlSource, err := os.ReadFile(*groupsFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open group file: %s\n", err.Error())
		os.Exit(1)
	}

	db, err := database.GetGroupDBFromYaml(string(yamlSource))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %s\n", err.Error())
		os.Exit(1)
	}

	mux := web.NewTaxisMux(&web.TaxisConfig{
		Database:        db,
		GroupHeaderName: "X-Groups",
	})

	http.ListenAndServe("0.0.0.0:3456", mux)
}
