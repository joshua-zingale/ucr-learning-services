package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/database"
)

var groupHeaderName string = "X-Groups"

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromRequest(r)
		groups, err := database.GetGroups(userId)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to assign organizational groups to '%s'", userId)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			log.Println(errorMessage)
			return
		}

		addGroupsToResponseHeader(w, groups)
	})

	return mux
}

func addGroupsToResponseHeader(w http.ResponseWriter, groups []string) {
	w.Header().Set(groupHeaderName, strings.Join(groups, ","))
}

func getUserIdFromRequest(_ *http.Request) string {
	return "bob@ucr.edu"
}
