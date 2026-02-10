package agentclass

import (
	"context"

	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/aitutormux"
)

type StandardAgentDriver struct{}

func (sad *StandardAgentDriver) Generate(ctx context.Context, config []byte, messages []aitutormux.ChatMessage) (string, error) {
	return "This is a default message from the standard agent driver.", nil
}
