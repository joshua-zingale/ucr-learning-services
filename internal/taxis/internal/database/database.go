package database

import (
	"fmt"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/yaml"
)

const subGroupSeparator = "."

type GroupDB struct {
	userIdToGroups map[string][]string
}

func (gdb *GroupDB) GetGroups(userId string) ([]string, error) {
	if groups, ok := gdb.userIdToGroups[userId]; ok {
		return groups, nil
	}
	return []string{}, nil
}

func (gdb *GroupDB) GetAllUsersIds() []string {
	keys := make([]string, 0, len(gdb.userIdToGroups))
	for k := range gdb.userIdToGroups {
		keys = append(keys, k)
	}

	return keys
}

func GetGroupDBFromYaml(yamlSource string) (*GroupDB, error) {
	data, err := yaml.ParseYaml(yamlSource)
	if err != nil {
		return nil, err
	}
	if mapping, ok := data.(map[string]any); ok {
		userIdToGroups, err := invertYamlData(mapping)
		if err != nil {
			return nil, fmt.Errorf("invalid YAML: %e", err)
		}
		return &GroupDB{
			userIdToGroups: userIdToGroups,
		}, nil
	}
	return nil, fmt.Errorf("invalid type: found %T but must be a mapping", data)
}

func invertYamlData(yamlData map[string]any) (map[string][]string, error) {

	type groupingContext struct {
		ancestors []string
		mapping   map[string]any
	}

	userIdToGroups := make(map[string][]string)

	groupContexts := []groupingContext{{
		ancestors: []string{},
		mapping:   yamlData,
	}}

	for len(groupContexts) > 0 {
		groupContext := groupContexts[len(groupContexts)-1]
		groupContexts = groupContexts[:len(groupContexts)-1]

		for groupName, value := range groupContext.mapping {
			switch value := value.(type) {
			case map[string]any:
				subGrouping := value
				groupContexts = append(groupContexts, groupingContext{
					ancestors: append(groupContext.ancestors, groupName),
					mapping:   subGrouping,
				})
			case []string:
				userIds := value

				newGroup := strings.Join(append(groupContext.ancestors, groupName), subGroupSeparator)

				for _, userId := range userIds {
					if groups, ok := userIdToGroups[userId]; ok {
						userIdToGroups[userId] = append(groups, newGroup)
					} else {
						userIdToGroups[userId] = []string{newGroup}
					}
				}
			default:
				return nil, fmt.Errorf("invalid type: %T", value)
			}
		}
	}
	return userIdToGroups, nil
}
