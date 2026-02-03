package aitutormux

import (
	"encoding/json"
	"fmt"
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

		agentId, err := rc.ResourceId.GetInt("agent")
		if err != nil {
			http.Error(w, fmt.Sprintf("%s", err.Error()), http.StatusBadRequest)
			return
		}

		agent, err := config.Db.GetAgent(r.Context(), &AgentId{
			AgentId: agentId,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Not Found: %s", err.Error()), http.StatusNotFound)
		}

		enc := json.NewEncoder(w)
		enc.Encode(agent)
	})

	return mux
}
