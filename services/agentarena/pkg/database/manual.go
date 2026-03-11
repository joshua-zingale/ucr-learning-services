package database

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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

// Validates that config matches the json schema of agent's class before updating,
// returning an error if validation fails.
func (q *Queries) UpdateAgentConfig(ctx context.Context, agent GetAgentRow, config []byte, configSchema []byte) error {

	currentConfig := agent.Config.Bytes

	compiler := jsonschema.NewCompiler()
	compiler.AddResource("config.json", bytes.NewReader(configSchema))

	schema, err := compiler.Compile("config.json")
	if err != nil {
		return fmt.Errorf("invalid JSON schema for agent class %s: %w", agent.AgentClassID, err)
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
		AgentID: agent.AgentID,
		Config: pgtype.JSONB{
			Bytes:  patchedConfig,
			Status: pgtype.Present,
		},
	})
}
