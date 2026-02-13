package database

import (
	"context"
	"database/sql"
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
