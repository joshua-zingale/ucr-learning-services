package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type GroupDB interface {
	GetGroups(string) ([]string, error)
}

type TaxisConfig struct {

	// The database used to determine those groups to which a user belongs
	Database GroupDB

	// The name of the header from which the userId is read
	UserIdHeaderName string

	// The root URI path. e.g. "/taxis" or "/"
	RootPath string
}

func setDefaultsAndValidate(config *TaxisConfig) error {
	if config.Database == nil {
		return cannotBeNilError("Database")
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

	mux.HandleFunc(fmt.Sprintf("%s/userinfo", config.RootPath), func(w http.ResponseWriter, r *http.Request) {
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
		addGroupsToResponse(w, groups)
	})

	return mux, nil
}

func addGroupsToResponse(w http.ResponseWriter, groups []string) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	encoder.Encode(struct {
		Groups []string `json:"groups"`
	}{
		Groups: groups,
	})
}

func getUserIdFromRequest(r *http.Request, headerName string) (string, error) {
	userId := r.Header.Get(headerName)
	if len(userId) == 0 {
		return "", fmt.Errorf("missing header value: %s", headerName)
	}
	return userId, nil
}
