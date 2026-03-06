package agentclass

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentmux"
	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/database"
)

type OllamaAgentDriver struct {
	Url string
}

type OllamaAgentConfig struct {
	SystemPrompt string `json:"systemPrompt"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaRole string

var (
	ollamaRoleUser      ollamaRole = "user"
	ollamaRoleAssistant ollamaRole = "assistant"
	ollamaRoleSystem    ollamaRole = "system"
)

type ollamaMessage struct {
	Role    ollamaRole `json:"role"`
	Content string     `json:"content"`
}

type ollamaResponse struct {
	Message ollamaMessage `json:"message"`
}

func (oad *OllamaAgentDriver) Generate(ctx context.Context, config []byte, messages []agentmux.ChatMessage) (string, error) {
	var ollamaConfig OllamaAgentConfig
	if err := json.Unmarshal(config, &ollamaConfig); err != nil {
		return "", fmt.Errorf("failed to decode config: %w", err)
	}

	ollamaMessages := make([]ollamaMessage, len(messages)+1)

	ollamaMessages[0] = ollamaMessage{
		Role:    ollamaRoleSystem,
		Content: ollamaConfig.SystemPrompt,
	}
	for i, msg := range messages {
		var role ollamaRole
		switch msg.MessageType {
		case database.MessageTypeAgent:
			role = ollamaRoleAssistant
		case database.MessageTypeUser:
			role = ollamaRoleUser
		}
		ollamaMessages[i+1] = ollamaMessage{
			Role:    role,
			Content: msg.Content,
		}
	}

	req := ollamaRequest{
		Model:    "llama3.2:3B",
		Messages: ollamaMessages,
		Stream:   false,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return "", err
	}

	ollamaReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/chat", oad.Url), &buf)
	if err != nil {
		return "", err
	}
	ollamaReq.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(ollamaReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", errors.New("could not get valid response from Ollama")
	}

	var resData ollamaResponse
	if err := json.NewDecoder(res.Body).Decode(&resData); err != nil {
		return "", err
	}

	return resData.Message.Content, nil
}
