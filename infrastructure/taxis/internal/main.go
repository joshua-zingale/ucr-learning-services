package main

import (
	"fmt"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/yaml"
)

func main() {

	fmt.Println(yaml.ParseYaml("- tom\n- dick\n- harry"))
}
