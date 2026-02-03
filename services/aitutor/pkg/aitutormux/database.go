package aitutormux

import "context"

type Database interface {
	GetAgent(context.Context, *AgentId) (*Agent, error)
	// GetAgentConversationsForUser(context.Context, AgentId, UserId)
	// GetMessages(context.Context, ConversationId) []Message
	// PostMessage(context.Context, ConversationId, Message)
}

type AgentId struct {
	AgentId int `json:"agentId"`
}

type UserId struct {
	UserId string `json:"userId"`
}

type ConversationId struct {
	AgentId
	UserId
	ConversationId int `json:"conversationId"`
}

type MessageID struct {
	AgentId
	UserId
	ConversationId
	MessageID int `json:"messageID"`
}

type Agent struct {
	AgentId
	Name string `json:"name"`
}

type Conversation struct{}

type Message struct{}
