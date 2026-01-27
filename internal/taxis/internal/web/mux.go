package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/database"
)

const groupHeaderName string = "X-Groups"

const groupSeparator string = ","

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	db, err := database.GetGroupDBFromYaml(`
cs100:
 instructor:
  - bob@ucr.edu`)
	if err != nil {
		panic(err)
	}
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromRequest(r)
		groups, err := db.GetGroups(userId)
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
	w.Header().Set(groupHeaderName, strings.Join(groups, groupSeparator))
}

func getUserIdFromRequest(_ *http.Request) string {
	return "bob@ucr.edu"
}
