package agentregistry

import (
	"context"
	"fmt"

	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/aitutormux"
	"golang.org/x/exp/constraints"
)

type integer interface {
	constraints.Integer
	comparable
}

type AgentClassDriverRegistry[T integer] struct {
	drivers  map[T]aitutormux.AgentClassDriver
	slugToId map[string]T
}

func (acdr *AgentClassDriverRegistry[T]) GetFromId(id T) (aitutormux.AgentClassDriver, bool) {
	driver, ok := acdr.drivers[id]
	return driver, ok
}

func (acdr *AgentClassDriverRegistry[T]) GetFromSlug(slug string) (aitutormux.AgentClassDriver, bool) {
	id, ok := acdr.slugToId[slug]
	if !ok {
		return nil, false
	}
	return acdr.GetFromId(id)
}

func (acdr *AgentClassDriverRegistry[T]) Register(id T, slug string, agentClass aitutormux.AgentClassDriver) error {
	if _, ok := acdr.drivers[id]; ok {
		return fmt.Errorf("duplicate agent class id, '%d'", id)
	}
	if _, ok := acdr.slugToId[slug]; ok {
		return fmt.Errorf("duplicate agent class slug, '%s'", slug)
	}

	acdr.slugToId[slug] = id
	acdr.drivers[id] = agentClass

	return nil
}

func New[T integer](ctx context.Context, slugToDriver map[string]aitutormux.AgentClassDriver, getIdFromSlug func(context.Context, string) (T, error)) (*AgentClassDriverRegistry[T], error) {
	acdr := AgentClassDriverRegistry[T]{
		drivers:  map[T]aitutormux.AgentClassDriver{},
		slugToId: map[string]T{},
	}

	for slug, driver := range slugToDriver {
		id, err := getIdFromSlug(ctx, slug)
		if err != nil {
			return nil, fmt.Errorf("failed to locate agent class's ID: %w", err)
		}
		if err := acdr.Register(id, slug, driver); err != nil {
			return nil, err
		}

	}
	return &acdr, nil
}
