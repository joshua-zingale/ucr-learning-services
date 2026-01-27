package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultGroupHeaderName string = "X-Groups"

const groupSeparator string = ","

type GroupDB interface {
	GetGroups(string) ([]string, error)
}

type TaxisConfig struct {

	// The database used to determine those groups to which a user belongs
	Database GroupDB

	// The name of the header to which the groups are written
	GroupHeaderName string
}

func setDefaults(config *TaxisConfig) error {
	if config.Database == nil {
		return cannotBeNilError("Database")
	}

	if len(config.GroupHeaderName) == 0 {
		config.GroupHeaderName = defaultGroupHeaderName
	}
	return nil
}

func cannotBeNilError(field string) error {
	return fmt.Errorf("%s cannot be nil", field)
}

func NewTaxisMux(config *TaxisConfig) *http.ServeMux {

	setDefaults(config)

	mux := http.NewServeMux()

	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromRequest(r)
		groups, err := config.Database.GetGroups(userId)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to assign organizational groups to '%s'", userId)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			log.Println(errorMessage)
			return
		}

		addGroupsToResponseHeader(w, config.GroupHeaderName, groups)
	})

	return mux
}

func addGroupsToResponseHeader(w http.ResponseWriter, headerName string, groups []string) {
	w.Header().Set(headerName, strings.Join(groups, groupSeparator))
}

func getUserIdFromRequest(_ *http.Request) string {
	return "bob@ucr.edu"
}
