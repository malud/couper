package hooks

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"

	"github.com/avenga/couper/config/request"
	"github.com/avenga/couper/eval"
)

var _ logrus.Hook = &CustomLogs{}

const customLogField = "custom"

type CustomLogs struct{}

func (c *CustomLogs) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (c *CustomLogs) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		if t, exists := entry.Data["type"]; exists {
			switch t {
			case "couper_access":
				fire(entry, request.LogCustomAccess)
			case "couper_backend":
				fire(entry, request.LogCustomUpstream)
			}
		}
	}

	return nil
}

func fire(entry *logrus.Entry, bodyKey request.ContextKey) {
	var evalCtx *eval.Context

	customEvalCtxCh, ok := entry.Context.Value(request.LogCustomEvalResult).(chan *eval.Context)
	if ok {
		select {
		case evalCtx = <-customEvalCtxCh:
		default: // pass through, e.g. on early errors we will not receive something useful
		}
	}

	if evalCtx == nil {
		evalCtx, ok = entry.Context.Value(request.ContextType).(*eval.Context)
		if !ok {
			return
		}
	}

	bodies := entry.Context.Value(bodyKey)
	if bodies == nil {
		return
	}

	hclBodies, ok := bodies.([]hcl.Body)
	if !ok {
		return
	}

	if fields := eval.ApplyCustomLogs(evalCtx.HCLContext(), hclBodies, entry); len(fields) > 0 {
		entry.Data[customLogField] = fields
	}
}
