package aitutormux

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/restapi"
)

type AuthService interface {
	Authenticate(r *http.Request) (*UserProfile, error)
}

type UserProfile struct {
	UserId     string
	UserGroups []string
}

type AiTutorConfig struct {
	Db   Database
	Auth AuthService
}

func NewAiTutorMux(config *AiTutorConfig) http.Handler {
	mux := http.NewServeMux()

	restapi.HandleResource(mux, "GET", AGENT_RESOURCE, func(w http.ResponseWriter, r *http.Request, rc *restapi.ResourceContext) {
		_, err := config.Auth.Authenticate(r)
		if err != nil {
			http.Error(w, "Not Authenticated", http.StatusUnauthorized)
			return
		}

		agentId, err := rc.ResourceId.GetInt(restapi.GetResourcePathVariableName(AGENT_RESOURCE))
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			log.Print(err.Error())
			return
		}

		agent, err := config.Db.GetAgent(r.Context(), AgentId(agentId))
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			log.Print(err.Error())
			return
		}

		agentConfig, err := config.Db.GetAgentConfig(r.Context(), AgentId(agentId))
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			log.Print(err.Error())
			return
		}

		enc := json.NewEncoder(w)
		enc.Encode(struct {
			Id           int    `json:"id"`
			Name         string `json:"name"`
			SystemPrompt string `json:"systemPrompt"`
		}{
			Id:           agentId,
			Name:         agent.Name,
			SystemPrompt: agentConfig.SystemPrompt,
		})
	})

	return mux
}
