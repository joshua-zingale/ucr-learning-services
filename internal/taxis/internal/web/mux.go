package web

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/constants"
)

type GroupDB interface {
	GetGroups(string) ([]string, error)
}

type TaxisConfig struct {
	Authenticator Authenticator

	// The database used to determine those groups to which a user belongs
	Database GroupDB

	// The name of the header to which the groups are written
	GroupsHeaderName string

	// The name of the header to which the userId is written
	UserIdHeaderName string

	// The root URI path. e.g. "/taxis"
	RootPath string
}

type Authenticator interface {
	Authenticate(r *http.Request) (string, error)
}

type Taxis struct {
	config *TaxisConfig
}

func (t *Taxis) handleAuth(w http.ResponseWriter, r *http.Request) {
	userId, err := t.config.Authenticator.Authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println(err.Error())
		return
	}
	groups, err := t.config.Database.GetGroups(userId)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to assign organizational groups to '%s'", userId)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		log.Println(errorMessage)
		return
	}
	w.Header().Set(t.config.UserIdHeaderName, userId)
	w.Header().Set(t.config.GroupsHeaderName, strings.Join(groups, constants.GroupSeparatorInHeader))
}

func NewTaxisMux(config *TaxisConfig) (*http.ServeMux, error) {

	if err := setDefaultsAndValidate(config); err != nil {
		return nil, err
	}

	taxis := Taxis{config: config}

	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("%s/auth", config.RootPath), taxis.handleAuth)

	return mux, nil
}

type ProxyAuthenticator struct {
	AuthenticationURL *url.URL
	UserIdHeaderName  string
	client            http.Client
}

func (pa *ProxyAuthenticator) Authenticate(r *http.Request) (string, error) {
	authReq, _ := http.NewRequestWithContext(r.Context(), "GET", pa.AuthenticationURL.String(), nil)
	authReq.Header = r.Header
	authResp, err := pa.client.Do(authReq)

	if err != nil {
		return "", err
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("authentication failed with status: %d", authResp.StatusCode)
	}

	userId := authResp.Header.Get(pa.UserIdHeaderName)

	if userId == "" {
		return "", fmt.Errorf("missing identity header: %s", pa.UserIdHeaderName)
	}
	return userId, nil
}

func setDefaultsAndValidate(config *TaxisConfig) error {
	if config.Database == nil {
		return cannotBeNilError("Database")
	}

	if len(config.GroupsHeaderName) == 0 {
		return cannotBeNilError("GroupHeaderName")
	}

	config.RootPath = strings.TrimSuffix(config.RootPath, "/")

	return nil
}

func cannotBeNilError(field string) error {
	return fmt.Errorf("%s cannot be nil", field)
}
