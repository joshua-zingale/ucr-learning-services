// Package qapi provides a functional, generic wrapper for creating standardized
// HTTP handlers. It decomposes the typical HTTP request lifecycle into discrete,
// testable stages: Authentication, Request Parsing, Authorization, Business Logic,
// and Rendering.
package qapi

import (
	"context"
	"net/http"
	"reflect"
)

type AuthFunction[AuthData any] func(*http.Request) (AuthData, error)
type ReadRequestFunction[Data any] func(*http.Request) (Data, error)
type AuthzFunction[AuthData, Req any] func(context.Context, AuthData, Req) error
type ActionFunction[AuthData, Req, Data any] func(context.Context, AuthData, Req) (Data, error)
type RenderFunction[Data any] func(context.Context, http.ResponseWriter, Data)

// QApiParams defines the lifecycle hooks for an HTTP endpoint.
// Any field may be left nil to skip that stage of processing.
//
//   - Auth:   Extracts credentials or session data from the raw request.
//   - Read:   Parses the request body or parameters into a structured Req type.
//   - Authz:  Validates if the authenticated user has permission to perform the action.
//   - Act:    The core business logic; takes the request and returns a response.
//   - Render: Handles writing the successful response to the ResponseWriter.
//   - RenderError: Centralized handler for any error occurring in the pipeline.
type QApiParams[AuthData, Req, Res any] struct {
	Auth        AuthFunction[AuthData]
	Read        ReadRequestFunction[Req]
	Authz       AuthzFunction[AuthData, Req]
	Act         ActionFunction[AuthData, Req, Res]
	Render      RenderFunction[Res]
	RenderError RenderFunction[error]
}

// Q(uick)Api constructs an http.HandlerFunc from a set of functional parameters.
//
// The execution flow follows a strict linear pipeline:
//
//	Auth -> Read -> Authz -> Act -> Render
//
// If any stage returns an error, execution halts and RenderError is invoked.
// If Render is nil, the handler returns http.StatusNoContent (204) on success.
// If RenderError is nil, the handler returns http.StatusInternalServerError (500) on error.
//
// Type safety is enforced via generics: the output of Auth and Read are
// automatically passed as typed arguments to Authz and Act.
func QApi[AuthData, Req, Res any](p QApiParams[AuthData, Req, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		renderError := p.RenderError
		if renderError == nil {
			renderError = func(ctx context.Context, w http.ResponseWriter, err error) {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
			}
		}
		var authData AuthData
		var err error
		if p.Auth != nil {
			authData, err = p.Auth(r)
			if err != nil {
				renderError(r.Context(), w, err)
				return
			}
		}

		var req Req
		if p.Read != nil {
			req, err = p.Read(r)
			if err != nil {
				renderError(r.Context(), w, err)
				return
			}
		}

		if p.Authz != nil {
			if err := p.Authz(r.Context(), authData, req); err != nil {
				renderError(r.Context(), w, err)
				return
			}
		}

		var res Res
		if p.Act != nil {

			res, err = p.Act(r.Context(), authData, req)
			if err != nil {
				renderError(r.Context(), w, err)
				return
			}
			res = fixNilSlice(res)
		}

		if p.Render != nil {
			p.Render(r.Context(), w, res)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func fixNilSlice[T any](val T) T {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Slice && v.IsNil() {
		return reflect.MakeSlice(v.Type(), 0, 0).Interface().(T)
	}
	return val
}
