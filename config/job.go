package config

import (
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/avenga/couper/config/meta"
	"github.com/avenga/couper/config/schema"
)

var (
	_ schema.BodySchema = &Job{}
)

type Job struct {
	Interval string   `hcl:"interval" docs:"Execution interval." type:"duration"`
	Name     string   `hcl:"name,label"`
	Remain   hcl.Body `hcl:",remain"`
	Requests Requests `hcl:"request,block" docs:"Configures a [request](/configuration/block/request) (zero or more)."`

	// Internally used
	Endpoint         *Endpoint
	IntervalDuration time.Duration
}

// Inline implements the <Inline> interface.
func (j Job) Inline() interface{} {
	type Inline struct {
		meta.LogFieldsAttribute
	}

	return &Inline{}
}

func (j Job) Schema() *hcl.BodySchema {
	s, _ := gohcl.ImpliedBodySchema(j)
	i, _ := gohcl.ImpliedBodySchema(j.Inline())

	return meta.MergeSchemas(s, i, meta.LogFieldsAttributeSchema)
}
