package definitions

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"

	"github.com/avenga/couper/config"
	"github.com/avenga/couper/config/request"
	"github.com/avenga/couper/eval"
	"github.com/avenga/couper/handler/middleware"
	"github.com/avenga/couper/logging"
	"github.com/avenga/couper/server/writer"
	"github.com/avenga/couper/utils"
)

type Job struct {
	conf     *config.Job
	handler  http.Handler
	interval time.Duration
	settings *config.Settings
}

type Jobs []*Job

func (j Jobs) Run(ctx context.Context, log *logrus.Entry) {
	if len(j) == 0 {
		return
	}

	logEntry := log.WithContext(ctx)
	logEntry.Data["type"] = "job"

	for _, job := range j {
		go job.Run(ctx, logEntry)
	}
}

func NewJob(j *config.Job, h http.Handler, settings *config.Settings) (*Job, error) {
	interval, err := time.ParseDuration(j.Interval)
	if err != nil {
		return nil, err
	}

	if interval == 0 {
		return nil, fmt.Errorf("job: %s: interval must be a positive number", j.Name)
	}

	return &Job{
		conf:     j,
		handler:  h,
		interval: interval,
		settings: settings,
	}, nil
}

func (j *Job) Run(ctx context.Context, logEntry *logrus.Entry) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	req.Header.Set("User-Agent", "Couper / "+utils.VersionName+" job-"+j.conf.Name)

	uidFn := middleware.NewUIDFunc(j.settings.RequestIDBackendHeader)

	t := time.NewTicker(time.Millisecond * 50)
	defer t.Stop()

	firstRun := true

	clh := middleware.NewCustomLogsHandler([]hcl.Body{j.conf.Remain}, j.handler, j.conf.Name)

	for {
		select {
		case <-ctx.Done():
			logEntry.WithFields(logrus.Fields{
				"name": j.conf.Name,
			}).Errorf("stopping: %v", ctx.Err())
			return
		case <-t.C:
			uid := uidFn()

			outReq := req.Clone(context.WithValue(ctx, request.UID, uid))

			evalCtx := eval.ContextFromRequest(outReq).WithClientRequest(outReq) // setup syncMap, upstream custom logs
			delete(evalCtx.HCLContext().Variables, eval.ClientRequest)           // this is the noop req from above, not helpful

			outCtx := context.WithValue(evalCtx, request.LogEntry, logEntry)
			outCtx = context.WithValue(outCtx, request.LogCustomAccess, []hcl.Body{j.conf.Remain}) // local custom logs
			outReq = outReq.WithContext(outCtx)

			n := time.Now()
			w := writer.NewResponseWriter(&noopResponseWriter{}, "")
			clh.ServeHTTP(w, outReq)
			logEntry.
				WithFields(logrus.Fields{
					"name": j.conf.Name,
					"timings": logging.Fields{
						"total":    logging.RoundMS(time.Since(n)),
						"interval": logging.RoundMS(j.interval),
					},
					"uid": uid,
				}).WithContext(outCtx).
				WithTime(n).
				Info()

			if firstRun {
				t.Reset(j.interval)
				firstRun = false
			}
		}
	}
}

var _ http.ResponseWriter = &noopResponseWriter{}

type noopResponseWriter struct {
	header http.Header
}

func (n noopResponseWriter) Header() http.Header {
	if n.header == nil {
		n.header = make(http.Header)
	}
	return n.header
}

func (n noopResponseWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (n noopResponseWriter) WriteHeader(_ int) {}
