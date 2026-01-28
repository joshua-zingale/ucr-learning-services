package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/database"
	"github.com/joshua-zingale/ucr-learning-services/internal/taxis/internal/web"
)

// Helper to find items in l1 that are not in l2
func listSetMinus[T comparable](l1 []T, l2 []T) []T {
	var l3 []T
outer:
	for _, e1 := range l1 {
		for _, e2 := range l2 {
			if e1 == e2 {
				continue outer
			}
		}
		l3 = append(l3, e1)
	}
	return l3
}

func TestTaxisMux_Integration(t *testing.T) {
	yamlData := `
administrator:
 - tom
cs100:
  instructor:
    - tom
  student:
    - dick
`

	db, err := database.GetGroupDBFromYaml(yamlData)
	if err != nil {
		t.Fatalf("Database initialization failed: %v", err)
	}

	config := &web.TaxisConfig{
		Database:         db,
		UserIdHeaderName: "X-User-ID",
		RootPath:         "/taxis",
	}

	mux, err := web.NewTaxisMux(config)
	if err != nil {
		t.Fatalf("Mux initialization failed: %v", err)
	}

	tests := []struct {
		name           string
		requestHeader  string
		userIdValue    string
		expectedStatus int
		expectedGroups []string
	}{
		{
			name:           "Retrieve_Multiple_Groups_For_Tom",
			requestHeader:  "X-User-ID",
			userIdValue:    "tom",
			expectedStatus: http.StatusOK,
			expectedGroups: []string{"cs100.instructor", "administrator"},
		},
		{
			name:           "Retrieve_Single_Group_For_Dick",
			requestHeader:  "X-User-ID",
			userIdValue:    "dick",
			expectedStatus: http.StatusOK,
			expectedGroups: []string{"cs100.student"},
		},
		{
			name:           "Empty_Groups_For_Unknown_User",
			requestHeader:  "X-User-ID",
			userIdValue:    "harry",
			expectedStatus: http.StatusOK,
			expectedGroups: []string{},
		},
		{
			name:           "Fail_On_Missing_User_ID_Header",
			requestHeader:  "Wrong-Header-Name",
			userIdValue:    "tom",
			expectedStatus: http.StatusBadRequest,
			expectedGroups: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/taxis/userinfo", nil)
			req.Header.Set(tt.requestHeader, tt.userIdValue)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Status mismatch: got %d, want %d", rr.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				groups := struct {
					Groups []string `json:"groups"`
				}{}
				if err := json.NewDecoder(rr.Body).Decode(&groups); err != nil {
					t.Errorf("Could not decode response as json data: %s", err.Error())
				}

				foundGroups := groups.Groups

				if missing := listSetMinus(tt.expectedGroups, foundGroups); len(missing) > 0 {
					t.Errorf("Missing roles: %v", missing)
				}

				if extra := listSetMinus(foundGroups, tt.expectedGroups); len(extra) > 0 {
					t.Errorf("Superfluous roles: %v", extra)
				}
			}
		})
	}
}
