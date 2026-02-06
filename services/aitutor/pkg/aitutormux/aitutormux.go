package aitutormux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/database"
)

type AuthService interface {
	Authenticate(r *http.Request) (UserProfile, error)
}

type UserProfile struct {
	UserId     string
	UserGroups []string
}

type AiTutorConfig struct {
	Db   database.Queries
	Auth AuthService
}

type aiTutorHandler struct {
	*AiTutorConfig
}

type authFunction[AuthData any] func(w http.ResponseWriter, r *http.Request) (AuthData, bool)
type fetchFunction[AuthData any, Data any] func(w http.ResponseWriter, r *http.Request, data AuthData) (Data, bool)
type renderFunction[Data any] func(w http.ResponseWriter, r *http.Request, data Data)

func buildRoute[AuthData, Data any](auth authFunction[AuthData], fetch fetchFunction[AuthData, Data], render renderFunction[Data]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authData, ok := auth(w, r)
		if !ok {
			return
		}
		data, ok := fetch(w, r, authData)
		if !ok {
			return
		}
		render(w, r, data)
	}
}

func NewAiTutorMux(config *AiTutorConfig) http.Handler {
	if config == nil {
		panic("config cannot be nil")
	}

	mux := http.NewServeMux()

	h := aiTutorHandler{
		AiTutorConfig: config,
	}

	mux.HandleFunc("GET /api/agents/{agentId}", buildRoute(h.authenticate, h.fetchAgent, renderJson))
	mux.HandleFunc("GET /api/conversations", buildRoute(h.authenticate, h.fetchConversations, renderJson))

	return mux
}

func (ath *aiTutorHandler) fetchConversations(w http.ResponseWriter, r *http.Request, profile UserProfile) ([]database.GetConversationsRow, bool) {
	conversations, err := ath.Db.GetConversations(r.Context(), profile.UserId)
	if err != nil {
		http.Error(w, "Internal Error: failed to fetch data", http.StatusInternalServerError)
		return nil, false
	}
	return conversations, true
}

func (ath *aiTutorHandler) fetchAgent(w http.ResponseWriter, r *http.Request, profile UserProfile) (database.GetAgentFullRow, bool) {

	var agent database.GetAgentFullRow
	fmt.Println("here")
	agentId, err := strconv.ParseInt(r.PathValue("agentId"), 10, 32)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return agent, false
	}

	if hasPermission, err := ath.Db.HasAgentPermission(r.Context(), database.HasAgentPermissionParams{
		UserID:   profile.UserId,
		GroupIds: profile.UserGroups,
		AgentID:  int32(agentId),
		Ability:  database.AgentAbilityTypeManage,
	}); err != nil || !hasPermission {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return agent, false
	}

	agent, err = ath.Db.GetAgentFull(r.Context(), int32(agentId))
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return agent, false
	}

	return agent, true
}

func (ath *aiTutorHandler) authenticate(w http.ResponseWriter, r *http.Request) (UserProfile, bool) {
	var profile UserProfile
	var err error
	profile, err = ath.Auth.Authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return profile, false
	}

	return profile, true
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
