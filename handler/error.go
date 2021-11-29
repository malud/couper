package handler

import (
	"context"
	"net/http"

	"github.com/hashicorp/hcl/v2"

	"github.com/avenga/couper/config/request"
	"github.com/avenga/couper/errors"
)

var _ http.Handler = &Error{}

type Error struct {
	kindsHandler map[string]http.Handler
	template     *errors.Template
}

func NewErrorHandler(kindsHandler map[string]http.Handler, errTpl *errors.Template) *Error {
	return &Error{
		kindsHandler: kindsHandler,
		template:     errTpl,
	}
}

func (e *Error) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	err, ok := req.Context().Value(request.Error).(*errors.Error)
	if !ok { // all errors within this context should have this type, otherwise an implementation error
		e.template.ServeError(errors.Server).ServeHTTP(rw, req)
		return
	}

	if e.kindsHandler == nil { // nothing defined, just serve err with template
		e.template.ServeError(err).ServeHTTP(rw, req)
		return
	}

	for _, kind := range err.Kinds() {
		ep, defined := e.kindsHandler[kind]
		if !defined {
			continue
		}

		// TODO: same for wildcard event match
		if eph, ek := ep.(interface{ BodyContext() hcl.Body }); ek {
			if b := req.Context().Value(request.LogCustomAccess); b != nil {
				bodies := b.([]hcl.Body)
				bodies = append(bodies, eph.BodyContext())
				*req = *req.WithContext(context.WithValue(req.Context(), request.LogCustomAccess, bodies))
			}
		}

		ep.ServeHTTP(rw, req)
		return
	}

	if ep, defined := e.kindsHandler[errors.Wildcard]; defined {
		ep.ServeHTTP(rw, req)
		return
	}

	// fallback with no matching error handler
	e.template.ServeError(err).ServeHTTP(rw, req)
}
