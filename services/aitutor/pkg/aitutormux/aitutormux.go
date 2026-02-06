package aitutormux

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/joshua-zingale/ucr-learning-services/services/aitutor/pkg/database"
)

type AuthService interface {
	Authenticate(r *http.Request) (*UserProfile, error)
}

type UserProfile struct {
	UserId     string
	UserGroups []string
}

type AiTutorConfig struct {
	Db   database.Queries
	Auth AuthService
}

func NewAiTutorMux(config *AiTutorConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /agents/{agentId}", func(w http.ResponseWriter, r *http.Request) {

		_, err := config.Auth.Authenticate(r)
		if err != nil {
			http.Error(w, "Not Authenticated", http.StatusUnauthorized)
			return
		}

		agent, err := getResourceByIntInPath(w, r, "agentId", config.Db.GetAgentFull)
		if err != nil {
			return
		}

		enc := json.NewEncoder(w)
		enc.Encode(struct {
			Id           int32  `json:"id"`
			Name         string `json:"name"`
			SystemPrompt string `json:"systemPrompt"`
		}{
			Id:           agent.AgentID,
			Name:         agent.Name,
			SystemPrompt: agent.SystemPrompt,
		})
	})

	return mux
}

var AlreadyResponded error = errors.New("response already written")

func getResourceByIntInPath[VAL any](w http.ResponseWriter, r *http.Request, pathParamName string, getter func(context.Context, int32) (VAL, error)) (VAL, error) {
	objId, err := strconv.ParseInt(r.PathValue(pathParamName), 10, 32)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Print(err.Error())
		var i VAL
		return i, AlreadyResponded
	}
	return getResource(w, r, int32(objId), getter)
}

func getResource[ID any, VAL any](w http.ResponseWriter, r *http.Request, objId ID, getter func(context.Context, ID) (VAL, error)) (VAL, error) {
	obj, err := getter(r.Context(), objId)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Print(err.Error())
		return obj, AlreadyResponded
	}
	return obj, nil
}
