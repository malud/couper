package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/docker/go-units"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"

	"github.com/avenga/couper/errors"
	"github.com/avenga/couper/eval"
	"github.com/avenga/couper/handler/producer"
)

var _ http.Handler = &Endpoint{}

const defaultReqBodyLimit = "64MiB"

type Endpoint struct {
	evalContext *hcl.EvalContext
	log         *logrus.Entry
	opts        *EndpointOptions
	proxies     producer.Roundtrips
	redirect    *producer.Redirect
	requests    producer.Roundtrips
	response    *producer.Response
}

type EndpointOptions struct {
	Context       hcl.Body
	ReqBufferOpts eval.BufferOption
	ReqBodyLimit  int64
	Error         *errors.Template
}

func NewEndpoint(opts *EndpointOptions, evalCtx *hcl.EvalContext, log *logrus.Entry,
	proxies producer.Proxies, requests producer.Requests) *Endpoint {
	opts.ReqBufferOpts |= eval.MustBuffer(opts.Context) // TODO: proper configuration on all hcl levels
	return &Endpoint{
		evalContext: evalCtx,
		log:         log,
		opts:        opts,
		proxies:     proxies,
		requests:    requests,
	}
}

func (e *Endpoint) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	subCtx, cancel := context.WithCancel(req.Context())
	defer cancel()

	if err := e.SetGetBody(req); err != nil {
		e.opts.Error.ServeError(err).ServeHTTP(rw, req)
		return
	}

	if ee := eval.ApplyRequestContext(e.evalContext, e.opts.Context, req); ee != nil {
		e.log.Error(ee)
	}

	proxyResults := make(producer.Results)
	requestResults := make(producer.Results)

	// go for it due to chan write on error
	go e.proxies.Produce(subCtx, req, e.evalContext, proxyResults)
	go e.requests.Produce(subCtx, req, e.evalContext, requestResults)

	beresps := make(map[string]*producer.Result)
	// TODO: read parallel, proxy first for now
	e.readResults(proxyResults, beresps)
	e.readResults(requestResults, beresps)

	var clientres *http.Response
	var err error

	// assume prio or err on conf load if set with response
	if e.redirect != nil {
		clientres = e.newRedirect()
	} else if e.response != nil {
		clientres = e.newResponse(beresps)
	} else {
		if len(beresps) > 1 {
			e.log.Error("endpoint configuration error")
			return
		}
		for _, result := range beresps {
			clientres = result.Beresp
			err = result.Err
			break
		}
	}

	if err != nil {
		e.log.Errorf("upstream error: %v", err)
		return
	}

	// always apply before write: redirect, response
	if err = eval.ApplyResponseContext(e.evalContext, e.opts.Context, req, clientres); err != nil {
		e.log.Error(err)
	}

	if err = clientres.Write(rw); err != nil {
		e.log.Errorf("endpoint write error: %v", err)
	}
}

// SetGetBody determines if we have to buffer a request body for further processing.
// First of all the user has a related reference within a related options context declaration.
// Additionally the request body is nil or a NoBody type and the http method has no body restrictions like 'TRACE'.
func (e *Endpoint) SetGetBody(req *http.Request) error {
	if req.Method == http.MethodTrace {
		return nil
	}

	if (e.opts.ReqBufferOpts & eval.BufferRequest) != eval.BufferRequest {
		return nil
	}

	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		buf := &bytes.Buffer{}
		lr := io.LimitReader(req.Body, e.opts.ReqBodyLimit+1)
		n, err := buf.ReadFrom(lr)
		if err != nil {
			return err
		}

		if n > e.opts.ReqBodyLimit {
			return errors.APIReqBodySizeExceeded
		}

		bodyBytes := buf.Bytes()
		req.GetBody = func() (io.ReadCloser, error) {
			return eval.NewReadCloser(bytes.NewBuffer(bodyBytes), req.Body), nil
		}
	}

	return nil
}

func (e *Endpoint) newResponse(beresps map[string]*producer.Result) *http.Response {
	// TODO: beresps.eval....
	clientres := &http.Response{
		StatusCode: e.response.Status,
		Header:     e.response.Header,
	}
	return clientres
}

func (e *Endpoint) newRedirect() *http.Response {
	// TODO use http.RedirectHandler
	status := http.StatusMovedPermanently
	if e.redirect.Status > 0 {
		status = e.redirect.Status
	}
	return &http.Response{
		Header: e.redirect.Header,
		//Body:   e.redirect.Body, // TODO: closeWrapper
		StatusCode: status,
	}
}

func (e *Endpoint) readResults(requestResults producer.Results, beresps map[string]*producer.Result) {
	i := 0
	for r := range requestResults { // collect resps
		if r == nil {
			panic("implement nil result handling")
		}

		name := "default"
		if r.Beresp != nil {
			// TODO: safe bereq access
			n, ok := r.Beresp.Request.Context().Value("requestName").(string)
			if ok && n != "" {
				name = n
			}
		}
		beresps[strconv.Itoa(i)+name] = r
		i++
	}
}

func ParseBodyLimit(limit string) (int64, error) {
	requestBodyLimit := defaultReqBodyLimit
	if limit != "" {
		requestBodyLimit = limit
	}
	return units.FromHumanSize(requestBodyLimit)
}
