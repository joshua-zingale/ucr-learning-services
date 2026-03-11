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

type QApiParams[AuthData, Req, Res any] struct {
	Auth        AuthFunction[AuthData]
	Read        ReadRequestFunction[Req]
	Authz       AuthzFunction[AuthData, Req]
	Act         ActionFunction[AuthData, Req, Res]
	Render      RenderFunction[Res]
	RenderError RenderFunction[error]
}

// Creates a Q(uick)API endpoint. Any function in the params dict may be nil
// to be a NoOp.
func QApi[AuthData, Req, Res any](p QApiParams[AuthData, Req, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var authData AuthData
		var err error
		if p.Auth != nil {
			authData, err = p.Auth(r)
			if err != nil {
				p.RenderError(r.Context(), w, err)
				return
			}
		}

		var req Req
		if p.Read != nil {
			req, err = p.Read(r)
			if err != nil {
				p.RenderError(r.Context(), w, err)
				return
			}
		}

		if p.Authz != nil {
			if err := p.Authz(r.Context(), authData, req); err != nil {
				p.RenderError(r.Context(), w, err)
				return
			}
		}

		var res Res
		if p.Act != nil {

			res, err = p.Act(r.Context(), authData, req)
			if err != nil {
				p.RenderError(r.Context(), w, err)
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
