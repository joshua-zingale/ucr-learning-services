package main

import (
	"fmt"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/yaml"
)

func main() {
	fmt.Println(yaml.ParseYaml(`
bob:
 - 1
fred:
 - 3
 - 4
bob:
  seiger: 
   - 1
   - fdg
  bobinson:
roger:
 - hello`))
	// 	fmt.Println(yaml.ParseYaml(`

	// bob:
	//  tom:
	//  - h
	//  dick:
	//  - e
	//  - f

	//  harry:
	//   charles:
	//     - 1
	//     - 2
	//  creg:
	//    - 3
	//    - 4

	// rob:
	// - a
	// - b`))
}
