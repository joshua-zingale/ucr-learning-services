package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockGroupDB struct {
	mapping map[string][]string
}

func (m *MockGroupDB) GetGroups(userId string) ([]string, error) {
	if groups, ok := m.mapping[userId]; ok {
		return groups, nil
	}
	return []string{}, nil
}

func TestTaxisMux_Auth(t *testing.T) {

	db := &MockGroupDB{
		mapping: map[string][]string{
			"tom":  {"instructor", "administrator"},
			"dick": {"student"},
		},
	}

	tests := []struct {
		name             string
		requestHeaderKey string
		requestHeaderVal string
		expectedStatus   int
		expectedGroups   string
	}{
		{
			name:             "Success - Tom's groups",
			requestHeaderKey: "X-User-ID",
			requestHeaderVal: "tom",
			expectedStatus:   http.StatusOK,
			expectedGroups:   "instructor,administrator",
		},
		{
			name:             "Success - Dick's groups",
			requestHeaderKey: "X-User-ID",
			requestHeaderVal: "dick",
			expectedStatus:   http.StatusOK,
			expectedGroups:   "student",
		},
		{
			name:             "Error - Missing Header",
			requestHeaderKey: "Wrong-Header",
			requestHeaderVal: "tom",
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name:             "Success - User Not in DB (Empty Groups)",
			requestHeaderKey: "X-User-ID",
			requestHeaderVal: "harry",
			expectedStatus:   http.StatusOK,
			expectedGroups:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &TaxisConfig{
				Database:         db,
				UserIdHeaderName: "X-User-ID",
			}

			mux, err := NewTaxisMux(config)
			if err != nil {
				t.Fatalf("NewTaxisMux failed: %v", err)
			}

			req := httptest.NewRequest("GET", "/userinfo", nil)
			req.Header.Set(tt.requestHeaderKey, tt.requestHeaderVal)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			// Validate Status Code
			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedStatus)
			}

			// Validate Header Content (only on 200 OK)
			if tt.expectedStatus == http.StatusOK {
				groups := struct {
					Groups []string `json:"groups"`
				}{}
				if err := json.NewDecoder(rr.Body).Decode(&groups); err != nil {
					t.Errorf("Could not decode response as json data: %s", err.Error())
				}

				got := groups.Groups
				if strings.Join(got, ",") != tt.expectedGroups {
					t.Errorf("handler returned unexpected groups: got %v want %v",
						got, tt.expectedGroups)
				}
			}
		})
	}
}
