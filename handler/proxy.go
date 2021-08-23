package handler

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"

	"github.com/avenga/couper/config"
	"github.com/avenga/couper/config/request"
	"github.com/avenga/couper/eval"
	"github.com/avenga/couper/eval/content"
	"github.com/avenga/couper/handler/transport"
	"github.com/avenga/couper/internal/seetie"
	"github.com/avenga/couper/server/writer"
)

// headerBlacklist lists all header keys which will be removed after
// context variable evaluation to ensure to not pass them upstream.
var headerBlacklist = []string{"Authorization", "Cookie"}

// Proxy wraps a httputil.ReverseProxy to apply additional configuration context
// and have control over the roundtrip configuration.
type Proxy struct {
	backend      http.RoundTripper
	context      hcl.Body
	logger       *logrus.Entry
	reverseProxy *httputil.ReverseProxy
}

func NewProxy(backend http.RoundTripper, ctx hcl.Body, logger *logrus.Entry) *Proxy {
	proxy := &Proxy{
		backend: backend,
		context: ctx,
		logger:  logger,
	}
	rp := &httputil.ReverseProxy{
		Director: proxy.director,
		ErrorHandler: func(rw http.ResponseWriter, _ *http.Request, err error) {
			if rec, ok := rw.(*transport.Recorder); ok {
				rec.SetError(err)
			}
		},
		ErrorLog:  newErrorLogWrapper(logger),
		Transport: backend,
	}

	proxy.reverseProxy = rp
	return proxy
}

func (p *Proxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := eval.ApplyRequestContext(req.Context(), p.context, req); err != nil {
		return nil, err
	}

	if err := p.applyWebsockets(req); err != nil {
		return nil, err
	}

	url, err := content.GetContextAttribute(p.context, req.Context(), "url")
	if err != nil {
		return nil, err
	}
	if url != "" {
		ctx := context.WithValue(req.Context(), request.URLAttribute, url)
		*req = *req.WithContext(ctx)
	}

	rw := req.Context().Value(request.ResponseWriter).(*writer.Response)
	rec := transport.NewRecorder(rw)

	p.reverseProxy.ServeHTTP(rec, req)
	beresp, err := rec.Response(req)
	if err != nil {
		return beresp, err
	}
	err = eval.ApplyResponseContext(req.Context(), p.context, beresp) // TODO: log only
	return beresp, err
}

func (p *Proxy) director(req *http.Request) {
	for _, key := range headerBlacklist {
		req.Header.Del(key)
	}
}

// ErrorWrapper logs httputil.ReverseProxy internals with our own logrus.Entry.
type ErrorWrapper struct{ l logrus.FieldLogger }

func (e *ErrorWrapper) Write(p []byte) (n int, err error) {
	e.l.Error(strings.Replace(string(p), "\n", "", 1))
	return len(p), nil
}

func newErrorLogWrapper(logger logrus.FieldLogger) *log.Logger {
	return log.New(&ErrorWrapper{logger}, "", log.Lshortfile)
}

func (p *Proxy) applyWebsockets(req *http.Request) error {
	bodyContent, _, diags := p.context.PartialContent(config.Proxy{Remain: p.context}.Schema(true))
	if diags.HasErrors() {
		return diags
	}

	if wss := bodyContent.Blocks.OfType("websockets"); len(wss) == 1 {
		ctx := req.Context()

		ctx = context.WithValue(ctx, request.AllowWebsockets, true)
		*req = *req.WithContext(ctx)

		// This method needs the 'request.AllowWebsockets' flag in the 'req.context'.
		if eval.IsUpgradeRequest(req) {
			bodyContent, _, diags = wss[0].Body.PartialContent(config.WebsocketsInlineSchema)
			if diags.HasErrors() {
				return diags
			}
			if err := eval.ApplyRequestContext(req.Context(), wss[0].Body, req); err != nil {
				return err
			}

			if attr, ok := bodyContent.Attributes["timeout"]; ok {
				val, diags := attr.Expr.Value(nil)
				if diags.HasErrors() {
					return diags
				}

				str := seetie.ValueToString(val)

				timeout, err := time.ParseDuration(str)
				if str != "" && err != nil {
					return err
				}

				ctx = context.WithValue(ctx, request.WebsocketsTimeout, timeout)
				*req = *req.WithContext(ctx)
			}
		}
	}

	return nil
}
