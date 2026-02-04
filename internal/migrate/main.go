package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const _MIGRATION_STATE_FILENAME = ".migrationstate"

const _THE_BEGINNING = ""
const _MOST_RECENT = "~~~~~~~~~~"

const _UP_MIGRATION_INDICATOR = ".up."
const _DOWN_MIGRATION_INDICATOR = ".down."

var _EXECUTABLE_NAME = os.Args[0]

func main() {
	flag := flag.NewFlagSet(_EXECUTABLE_NAME, flag.ExitOnError)
	var (
		migrateCommand = flag.String("command", "", "the command to be run on each migration; '@#' will be swaped for the name of the migration file being run.")
		to             = flag.String("to", "", "the desired migration state: BOTTOM is a reserved name to undo all migrations and TOP to get to the latest.")
		isDryRun       = flag.Bool("dry-run", false, "if set, does not execute migration commands, but prints them out one per line")
		printState     = flag.Bool("print-state", false, "if set, prints the current state, i.e. the last 'up' migration run, then exits.")
	)

	flag.Parse(os.Args[1:])
	args := flag.Args()

	if *to == "" {
		fmt.Fprintln(os.Stderr, "must specify -to")
		os.Exit(1)
	}
	if *migrateCommand == "" {
		fmt.Fprintln(os.Stderr, "must specify -command")
		os.Exit(1)
	}

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-[-]flag [value]...] PATH_TO_MIGRATIONS\n", _EXECUTABLE_NAME)
		os.Exit(1)
	}

	pathToMigrations := args[0]
	pathToMigrationState := filepath.Join(pathToMigrations, _MIGRATION_STATE_FILENAME)

	var currentState string
	if _, err := os.Stat(pathToMigrationState); errors.Is(err, os.ErrNotExist) {
		currentState = ""
	} else {
		contents, err := os.ReadFile(pathToMigrationState)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not %s\n", err.Error())
			os.Exit(1)
		}
		currentState = string(contents)
	}

	if *printState {
		fmt.Println(currentState)
		os.Exit(0)
	}

	upMigrationIds := setFromList(map_(
		throw(getFilesContaining(pathToMigrations, ".up.")),
		func(name string) string {
			return trimSuffixBeginningIn(name, _UP_MIGRATION_INDICATOR)
		}))

	downMigrationIds := setFromList(map_(throw(getFilesContaining(pathToMigrations, ".down.")),
		func(name string) string {
			return trimSuffixBeginningIn(name, _DOWN_MIGRATION_INDICATOR)
		}))

	if !upMigrationIds.Eq(downMigrationIds) {
		upMigrationIds.Minus(downMigrationIds).ForEach(func(s string) {
			fmt.Fprintf(os.Stderr, "missing down migration for id '%s'\n", s)
		})
		downMigrationIds.Minus(upMigrationIds).ForEach(func(s string) {
			fmt.Fprintf(os.Stderr, "missing up migration for id '%s'\n", s)
		})
		os.Exit(1)
	}

	var targetState string
	switch *to {
	case "TOP":
		targetState = _MOST_RECENT
	case "BOTTOM":
		targetState = _THE_BEGINNING
	default:
		targetState = *to
		if !upMigrationIds.Contains(targetState) {
			fmt.Fprintf(os.Stderr, "Unknown destination state, '%s'\n", targetState)
			os.Exit(1)
		}
	}

	goingUp := currentState < targetState

	var migrationFileIndicator string
	var invertIfGoingDown func(bool) bool
	if goingUp {
		migrationFileIndicator = _UP_MIGRATION_INDICATOR
		invertIfGoingDown = func(b bool) bool { return b }
	} else {
		migrationFileIndicator = _DOWN_MIGRATION_INDICATOR
		invertIfGoingDown = func(b bool) bool { return !b }
	}

	migrationNameFromPath := func(s string) string {
		return trimSuffixBeginningIn(filepath.Base(s), migrationFileIndicator)
	}

	type MigrationFile struct {
		migrationName string
		filePath      string
	}

	migrationFiles := map_(
		throw(getFilesContaining(pathToMigrations, migrationFileIndicator)),
		func(name string) string {
			return filepath.Join(pathToMigrations, name)
		})

	sort.Slice(migrationFiles, func(i, j int) bool {
		return invertIfGoingDown(migrationFiles[i] < migrationFiles[j])
	})

	for i, migrationFile := range migrationFiles {
		var nextMigrationName string
		if goingUp {
			nextMigrationName = migrationNameFromPath(migrationFile)
		} else if i >= len(migrationFiles)-1 {
			nextMigrationName = _THE_BEGINNING
		} else {
			nextMigrationName = migrationNameFromPath(migrationFiles[i+1])
		}

		if invertIfGoingDown(nextMigrationName < currentState) || nextMigrationName == currentState {
			continue
		} else if invertIfGoingDown(nextMigrationName > targetState) && nextMigrationName != targetState {
			break
		}

		cmdString := strings.Replace(*migrateCommand, "@#", migrationFile, -1)
		if *isDryRun {
			fmt.Println(cmdString)
		} else {
			cmd := getCommand(cmdString)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed at %s: %s", migrationFile, err.Error())
				os.Exit(1)
			}
		}

		currentState = nextMigrationName
	}

	if *isDryRun {
		os.Exit(0)
	}

	if err := os.WriteFile(pathToMigrationState, []byte(currentState), os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Could not %s", err.Error())
		os.Exit(1)
	}

}

func trimSuffixBeginningIn(s string, suffixStart string) string {
	return strings.Split(s, suffixStart)[0]
}

func filter[T any](slice []T, filter func(T) bool) []T {
	var filtered []T
	for _, e := range slice {
		if filter(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func map_[T any, U any](slice []T, map_ func(T) U) []U {
	mapped := make([]U, len(slice))
	for i, e := range slice {
		mapped[i] = map_(e)
	}
	return mapped
}

func getFilesContaining(dir string, substr string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	return filter(
		map_(
			filter(
				entries,
				func(e os.DirEntry) bool {
					return !e.IsDir()
				}),
			func(e os.DirEntry) string {
				return e.Name()
			}),
		func(name string) bool {
			return strings.Contains(name, substr)
		}), nil
}

type set[T comparable] map[T]struct{}

func (s set[T]) Add(e T) {
	s[e] = struct{}{}
}

func (s set[T]) Contains(e T) bool {
	_, ok := s[e]
	return ok
}

func (s set[T]) Minus(subtrahend set[T]) set[T] {
	diff := make(set[T])
	for e := range s {
		if _, ok := subtrahend[e]; !ok {
			diff.Add(e)
		}
	}
	return diff
}

func (s set[T]) Eq(other set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for e := range other {
		if _, ok := s[e]; !ok {
			return false
		}
	}
	return true
}

func (s set[T]) ForEach(f func(T)) {
	for e := range s {
		f(e)
	}
}

func throw[T any](other T, err error) T {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
	return other
}

func setFromList[T comparable](l []T) set[T] {
	s := make(set[T])
	for _, e := range l {
		s.Add(e)
	}
	return s
}

func getCommand(cmd string) *exec.Cmd {
	var c *exec.Cmd
	operatingSystem := runtime.GOOS

	switch operatingSystem {
	case "windows":
		c = exec.Command("cmd", "/C", cmd)
	default:
		c = exec.Command("sh", "-c", cmd)
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c
}
