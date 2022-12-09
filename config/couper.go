package config

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/avenga/couper/config/configload/file"
	"github.com/avenga/couper/config/schema"
)

// DefaultFilename defines the default filename for a couper config file.
const DefaultFilename = "couper.hcl"

var _ schema.BodySchema = Couper{}

// Couper represents the <Couper> config object.
type Couper struct {
	Context     context.Context
	Environment string
	Files       file.Files
	Defaults    *Defaults    `hcl:"defaults,block"`
	Definitions *Definitions `hcl:"definitions,block"`
	Servers     Servers      `hcl:"server,block"`
	Settings    *Settings    `hcl:"settings,block"`
}

func (c Couper) Schema() *hcl.BodySchema {
	couperSchema, _ := gohcl.ImpliedBodySchema(c)
	return couperSchema
}

func init() {
	schema.Registry.AddRecursive(Couper{})
	println("_")
}
