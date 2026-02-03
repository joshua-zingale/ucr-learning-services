package main

import (
	"flag"
	"log"
	"os"

	"github.com/joshua-zingale/ucr-learning-services/internal/confenvflag/pkg/confenvflag"
)

func main() {
	flagSet := flag.NewFlagSet("Test", flag.ExitOnError)

	default_ := flagSet.String("default", "default", "s-1 blah")
	overwrittenByConf := flagSet.String("overwritten-by-conf", "default", "s2 blah")
	overwrittenByEnv := flagSet.String("overwritten-by-env", "default", "s2 blah")
	overwrittenByConfThenEnv := flagSet.String("overwritten-by-conf-then-env", "default", "s2 blah")
	overwrittenByConfThenEnvThenArg := flagSet.String("overwritten-by-conf-then-env-then-arg", "arg", "s2 blah")
	overwrittenByConfThenArg := flagSet.String("overwritten-by-conf-then-arg", "default", "s2 blah")

	if err := confenvflag.Parse(flagSet, "TEST_", os.Args[1:]); err != nil {
		log.Fatalf("%s", err.Error())
	}

	assertEq(*default_, "default")
	assertEq(*overwrittenByConf, "conf")
	assertEq(*overwrittenByEnv, "env")
	assertEq(*overwrittenByConfThenEnv, "env")
	assertEq(*overwrittenByConfThenEnvThenArg, "arg")
	assertEq(*overwrittenByConfThenArg, "arg")

}

func assertEq[T comparable](got T, want T) {
	if got != want {
		log.Fatalf("got %v, wanted %v", got, want)
	}
}
