package agentclass

import (
	"context"
	"fmt"

	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentmux"
)

type StandardAgentDriver struct{}

func (sad *StandardAgentDriver) Generate(ctx context.Context, config []byte, messages []agentmux.ChatMessage) (string, error) {
	var lastMessage string = "None"
	if len(messages) > 0 {
		lastMessage = messages[len(messages)-1].Content
	}
	return fmt.Sprintf("This is a default message from the standard agent driver. History length: %d, Previoius message: %s, config: %s", len(messages), lastMessage, config), nil
}
