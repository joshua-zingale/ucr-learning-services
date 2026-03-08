package agentregistry

import (
	"context"
	"fmt"

	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentmux"
)

type AgentClassDriverRegistry struct {
	drivers map[string]agentmux.AgentClassDriver
}

func (acdr *AgentClassDriverRegistry) GetFromId(id string) (agentmux.AgentClassDriver, bool) {
	driver, ok := acdr.drivers[id]
	return driver, ok
}

func (acdr *AgentClassDriverRegistry) Register(id string, agentClass agentmux.AgentClassDriver) error {
	if _, ok := acdr.drivers[id]; ok {
		return fmt.Errorf("duplicate agent class id, '%s'", id)
	}

	acdr.drivers[id] = agentClass

	return nil
}

func New(ctx context.Context, idToDriver map[string]agentmux.AgentClassDriver) (*AgentClassDriverRegistry, error) {
	acdr := AgentClassDriverRegistry{
		drivers: map[string]agentmux.AgentClassDriver{},
	}
	for id, driver := range idToDriver {
		if err := acdr.Register(id, driver); err != nil {
			return nil, err
		}
	}
	return &acdr, nil
}
