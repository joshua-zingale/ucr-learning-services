package aitutormux

import (
	"context"
	"encoding/json"
	"fmt"
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

		agentId, err := strconv.ParseInt(r.PathValue("agentId"), 10, 32)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			log.Print(err.Error())
			return
		}

		agent, err := config.Db.GetAgent(r.Context(), int32(agentId))
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			log.Print(err.Error())
			return
		}

		agentConfig, err := config.Db.GetAgentConfig(r.Context(), int32(agentId))
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			log.Print(err.Error())
			return
		}

		enc := json.NewEncoder(w)
		enc.Encode(struct {
			Id           int64  `json:"id"`
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

func getResourceByIntInPath[VAL any](w http.ResponseWriter, r *http.Request, pathParamName string, getter func(context.Context, int) (*VAL, error)) (*VAL, error) {
	objId, err := strconv.Atoi(r.PathValue(pathParamName))
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Print(err.Error())
		return nil, fmt.Errorf("Not Found")
	}

	obj, err := getter(r.Context(), objId)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Print(err.Error())
		return nil, fmt.Errorf("Not Found")
	}
	return obj, nil
}

func getResource[ID any, VAL any](w http.ResponseWriter, r *http.Request, objId ID, getter func(context.Context, ID) (*VAL, error)) (*VAL, error) {
	obj, err := getter(r.Context(), objId)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Print(err.Error())
		return nil, fmt.Errorf("Not Found")
	}
	return obj, nil
}
