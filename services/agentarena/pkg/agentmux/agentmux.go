package agentmux

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/database"
)

type AuthService interface {
	Authenticate(r *http.Request) (UserProfile, error)
}

type UserProfile struct {
	UserId     string
	UserGroups []string
}

type ChatMessage struct {
	Sender  database.MessageType
	Content string
}

// Defines functionality for an AgentClass
type AgentClassDriver interface {
	Generate(ctx context.Context, config []byte, messages []ChatMessage) (string, error)
}

type AgentClassDriverRegistry interface {
	GetFromId(id int32) (AgentClassDriver, bool)
	GetFromSlug(slug string) (AgentClassDriver, bool)
}

type AgentArenaConfig struct {
	Db                       database.Queries
	Auth                     AuthService
	AgentClassDriverRegistry AgentClassDriverRegistry
}

type agentArenaHandler struct {
	*AgentArenaConfig
}

type contextKey int

type idType int

const (
	idKey contextKey = iota
)

type setContextFunction func(w http.ResponseWriter, r *http.Request) bool
type authFunction[AuthData any] func(w http.ResponseWriter, r *http.Request) (AuthData, bool)
type authzFunction[AuthData any] func(w http.ResponseWriter, r *http.Request, authData AuthData) bool
type fetchFunction[AuthData any, Data any] func(w http.ResponseWriter, r *http.Request, data AuthData) (Data, bool)
type renderFunction[Data any] func(w http.ResponseWriter, r *http.Request, data Data)

func NewAiTutorMux(config *AgentArenaConfig) http.Handler {
	if config == nil {
		panic("config cannot be nil")
	}

	mux := http.NewServeMux()

	h := agentArenaHandler{
		AgentArenaConfig: config,
	}

	mux.HandleFunc("GET /api/agents/{agentId}", buildRoute(h.setId("agentId"), h.authenticate, h.authorizeManageAgent, h.fetchAgent, renderJson))
	mux.HandleFunc("GET /api/conversations", buildRoute(nil, h.authenticate, nil, h.fetchConversations, renderJson))
	mux.HandleFunc("GET /api/conversations/{conversationId}/messages", buildRoute(h.setId("conversationId"), h.authenticate, h.authorizeStartedConversation, h.fetchConversationMessages, renderJson))

	return mux
}

func (ath *agentArenaHandler) fetchConversations(w http.ResponseWriter, r *http.Request, profile UserProfile) ([]database.GetConversationsRow, bool) {
	conversations, err := ath.Db.GetConversations(r.Context(), profile.UserId)
	if err != nil {
		ath.internalError(w, r)
		return nil, false
	}
	return conversations, true
}

func (ath *agentArenaHandler) fetchAgent(w http.ResponseWriter, r *http.Request, profile UserProfile) (database.GetAgentFullRow, bool) {

	agentId := r.Context().Value(idKey).(idType)

	agent, err := ath.Db.GetAgentFull(r.Context(), int32(agentId))
	if err != nil {
		ath.notFoundError(w, r)
		return agent, false
	}

	return agent, true
}

func (ath *agentArenaHandler) fetchConversationMessages(w http.ResponseWriter, r *http.Request, _ UserProfile) ([]database.GetConversationMessagesRow, bool) {
	conversationId := r.Context().Value(idKey).(idType)

	messages, err := ath.Db.GetConversationMessages(r.Context(), int32(conversationId))
	if err != nil {
		ath.internalError(w, r)
		return messages, false
	}

	return messages, true
}

func (ath *agentArenaHandler) authenticate(w http.ResponseWriter, r *http.Request) (UserProfile, bool) {
	var profile UserProfile
	var err error
	profile, err = ath.Auth.Authenticate(r)
	if err != nil {
		ath.unauthorizedError(w, r)
		return profile, false
	}

	return profile, true
}

func (ath *agentArenaHandler) authorizeManageAgent(w http.ResponseWriter, r *http.Request, profile UserProfile) bool {
	agentId := r.Context().Value(idKey).(idType)

	if hasPermission, err := ath.Db.HasAgentPermission(r.Context(), database.HasAgentPermissionParams{
		UserID:   profile.UserId,
		GroupIds: profile.UserGroups,
		AgentID:  int32(agentId),
		Ability:  database.AgentAbilityTypeManage,
	}); err != nil || !hasPermission {
		ath.forbiddenError(w, r)
		return false
	}

	return true
}

func (ath *agentArenaHandler) authorizeStartedConversation(w http.ResponseWriter, r *http.Request, profile UserProfile) bool {
	conversationId := r.Context().Value(idKey).(idType)

	if hasPermission, err := ath.Db.StartedConversation(r.Context(), database.StartedConversationParams{
		ConversationID: int32(conversationId),
		UserID:         profile.UserId,
	}); err != nil || !hasPermission {
		ath.forbiddenError(w, r)
		return false
	}

	return true
}

func (ath *agentArenaHandler) setId(pathParamName string) setContextFunction {
	return func(w http.ResponseWriter, r *http.Request) bool {
		id, err := strconv.ParseInt(r.PathValue(pathParamName), 10, 64)
		if err != nil {
			ath.notFoundError(w, r)
			return false
		}
		(*r) = *r.WithContext(context.WithValue(r.Context(), idKey, idType(id)))
		return true
	}

}

func (ath *agentArenaHandler) unauthorizedError(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (ath *agentArenaHandler) forbiddenError(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

func (ath *agentArenaHandler) notFoundError(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

func (ath *agentArenaHandler) internalError(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func renderJson[T any](w http.ResponseWriter, r *http.Request, data T) {
	if acceptsJson(r) {
		respondJson(w, data)
		return
	}
	http.Error(w, "Not Acceptable: JSON", http.StatusNotAcceptable)
}

func acceptsJson(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") || strings.Contains(r.Header.Get("Accept"), "*/*")
}

func respondJson[T any](w http.ResponseWriter, v T) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(v)
}

func buildRoute[AuthData, Data any](contextFunction setContextFunction, auth authFunction[AuthData], authz authzFunction[AuthData], fetch fetchFunction[AuthData, Data], render renderFunction[Data]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contextFunction != nil && !contextFunction(w, r) {
			return
		}

		var authData AuthData
		var ok bool

		if auth != nil {
			if authData, ok = auth(w, r); !ok {
				return
			}
		}

		if authz != nil && !authz(w, r, authData) {
			return
		}

		var data Data
		if fetch != nil {
			if data, ok = fetch(w, r, authData); !ok {
				return
			}
		}
		render(w, r, data)
	}
}
