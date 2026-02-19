package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/jackc/pgtype"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type GetConversationRow struct {
	Messages       []GetConversationMessagesRow `json:"messages"`
	Name           string                       `json:"name"`
	ConversationId int32                        `json:"conversationId"`
	ActiveAgentID  sql.NullInt32                `json:"activeAgentId"`
}

func (q *Queries) GetConversation(ctx context.Context, conversationID int32) (GetConversationRow, error) {
	var row GetConversationRow
	metadata, err := q.GetConversationMetadata(ctx, conversationID)
	if err != nil {
		return row, err
	}
	messages, err := q.GetConversationMessages(ctx, conversationID)
	if err != nil {
		return row, err
	}
	row.ActiveAgentID = metadata.ActiveAgentID
	row.ConversationId = conversationID
	row.Messages = messages
	row.Name = metadata.Name
	return row, nil
}

// Validates that config matches the json schema of agent's class,
// returning an error if not.
func (q *Queries) SetAgentConfig(ctx context.Context, agentID int32, config []byte) error {

	agent, err := q.GetAgent(ctx, agentID)
	if err != nil {
		return err
	}

	currentConfig := []byte(agent.Config)

	compiler := jsonschema.NewCompiler()
	compiler.AddResource("config.json", strings.NewReader(agent.ConfigSchema))

	schema, err := compiler.Compile("config.json")
	if err != nil {
		return fmt.Errorf("invalid JSON schema for agent class %d: %w", agent.AgentClassID, err)
	}

	var requestJson any
	if err = json.Unmarshal(config, &requestJson); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := schema.Validate(requestJson); err != nil {
		return err
	}

	patchedConfig, err := jsonpatch.MergePatch(currentConfig, config)
	if err != nil {
		return err
	}

	return q.setAgentConfigUnchecked(ctx, setAgentConfigUncheckedParams{
		AgentID: agentID,
		Config: pgtype.JSONB{
			Bytes:  patchedConfig,
			Status: pgtype.Present,
		},
	})
}
