package aitutormux

import (
	"fmt"
	"strconv"

	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/restapi"
)

func NewAgentId(rid restapi.ResourceId) (*AgentId, error) {
	agentIdString, ok := rid["agent"]
	if !ok {
		return nil, fmt.Errorf("missing agent id")
	}

	agentId, err := strconv.Atoi(agentIdString)
	if err != nil {
		return nil, fmt.Errorf("invalid agent id: %s", err)
	}
	return &AgentId{
		AgentId: agentId,
	}, nil
}
