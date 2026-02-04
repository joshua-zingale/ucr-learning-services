package aitutormux

import (
	"fmt"
	"net/http"
)

const _USER_GROUPS_HEADER_NAME = "X-Email"

const _USER_ID_HEADER_NAME = "X-Email"
const _USER_GROUP_SEPARATOR = ","

func Authenticate(r *http.Request) (*UserId, error) {
	if userIdString := r.Header.Get(_USER_ID_HEADER_NAME); userIdString != "" {
		userId := UserId(userIdString)
		return &userId, nil
	} else {
		return nil, fmt.Errorf("not authenticated")
	}
}
