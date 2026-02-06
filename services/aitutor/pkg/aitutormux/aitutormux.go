package aitutormux

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

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

		profile, err := config.Auth.Authenticate(r)
		if err != nil {
			http.Error(w, "Not Authenticated", http.StatusUnauthorized)
			return
		}

		agentId, err := strconv.ParseInt(r.PathValue("agentId"), 10, 32)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if hasPermission, err := config.Db.HasAgentPermission(r.Context(), database.HasAgentPermissionParams{
			UserID:   profile.UserId,
			GroupIds: profile.UserGroups,
			AgentID:  int32(agentId),
			Ability:  database.AgentPermissionTypeManage,
		}); err != nil || !hasPermission {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		agent, err := getResourceByIntInPath(w, r, "agentId", config.Db.GetAgentFull)
		if err != nil {
			return
		}

		if acceptsJson(r) {
			respondJson(w, agent)
			return
		}

		w.Write([]byte("Under construction!"))

	})

	return mux
}

func acceptsJson(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

func respondJson[T any](w http.ResponseWriter, v T) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(v)
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
