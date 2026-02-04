package aitutormux

import "context"

type Database interface {
	GetAgent(context.Context, AgentId) (*Agent, error)
	GetAgentConfig(context.Context, AgentId) (*AgentConfig, error)
	// GetAgentConversationsForUser(context.Context, AgentId, UserId)
	// GetMessages(context.Context, ConversationId) []Message
	// PostMessage(context.Context, ConversationId, Message)
}

type AgentId int
type UserId string
type ConversationId int
type MessageId int

type Agent struct {
	AgentId AgentId
	Name    string
}

type AgentConfig struct {
	AgentId      AgentId
	SystemPrompt string
	// BeginChatMessage string  `json:"beginChatMessage"`
}

type Conversation struct{}

type Message struct{}
