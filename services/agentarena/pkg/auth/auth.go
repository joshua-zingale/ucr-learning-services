package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/joshua-zingale/ucr-learning-services/services/agentarena/pkg/agentmux"
)

const _USER_ID_HEADER_NAME = "X-Email"
const _USER_GROUPS_HEADER_NAME = "X-Groups"
const _USER_GROUP_SEPARATOR = ","

type TaxisAuth struct{}

func (ta TaxisAuth) Authenticate(r *http.Request) (agentmux.UserProfile, error) {

	userId := r.Header.Get(_USER_ID_HEADER_NAME)
	if userId == "" {
		var up agentmux.UserProfile
		return up, fmt.Errorf("missing '%s' header", _USER_ID_HEADER_NAME)
	}

	return agentmux.UserProfile{
		UserId:     userId,
		UserGroups: strings.Split(r.Header.Get(_USER_GROUPS_HEADER_NAME), _USER_GROUP_SEPARATOR),
	}, nil
}
