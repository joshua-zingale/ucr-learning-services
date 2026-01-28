package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/constants"
)

type GroupDB interface {
	GetGroups(string) ([]string, error)
}

type TaxisConfig struct {

	// The database used to determine those groups to which a user belongs
	Database GroupDB

	// The name of the header to which the groups are written
	GroupsHeaderName string

	// The name of the header from which the userId is read
	UserIdHeaderName string

	// The root URI path. e.g. "/taxis"
	RootPath string
}

func setDefaultsAndValidate(config *TaxisConfig) error {
	if config.Database == nil {
		return cannotBeNilError("Database")
	}

	if len(config.GroupsHeaderName) == 0 {
		return cannotBeNilError("GroupHeaderName")
	}

	if len(config.UserIdHeaderName) == 0 {
		return cannotBeNilError("UserIdHeaderName")
	}

	config.RootPath = strings.TrimSuffix(config.RootPath, "/")

	return nil
}

func cannotBeNilError(field string) error {
	return fmt.Errorf("%s cannot be nil", field)
}

func NewTaxisMux(config *TaxisConfig) (*http.ServeMux, error) {

	if err := setDefaultsAndValidate(config); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("%s/auth", config.RootPath), func(w http.ResponseWriter, r *http.Request) {
		userId, err := getUserIdFromRequest(r, config.UserIdHeaderName)
		if err != nil {
			errorMessage := fmt.Sprintf("%s", err.Error())
			http.Error(w, errorMessage, http.StatusBadRequest)
			log.Println(errorMessage)
		}
		groups, err := config.Database.GetGroups(userId)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to assign organizational groups to '%s'", userId)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			log.Println(errorMessage)
			return
		}
		w.Header().Set(config.UserIdHeaderName, userId)
		addGroupsToResponseHeader(w, config.GroupsHeaderName, groups)
	})

	return mux, nil
}

func addGroupsToResponseHeader(w http.ResponseWriter, headerName string, groups []string) {
	w.Header().Set(headerName, strings.Join(groups, constants.GroupSeparatorInHeader))
}

func getUserIdFromRequest(r *http.Request, headerName string) (string, error) {
	userId := r.Header.Get(headerName)
	if len(userId) == 0 {
		return "", fmt.Errorf("missing header value: %s", headerName)
	}
	return userId, nil
}
