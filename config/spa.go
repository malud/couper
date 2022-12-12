package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/avenga/couper/config/meta"
	"github.com/avenga/couper/config/schema"
)

var (
	_ Body              = &Spa{}
	_ schema.BodySchema = &Spa{}
)

type SPAs []*Spa

// Spa represents the <Spa> object.
type Spa struct {
	AccessControl        []string       `hcl:"access_control,optional" docs:"Sets predefined [access control](../access-control) for {spa} block context."`
	BasePath             string         `hcl:"base_path,optional" docs:"Configures the path prefix for all requests."`
	BootStrapDataName    string         `hcl:"bootstrap_data_placeholder,optional" docs:"String which will be replaced with {bootstrap_data}." default:"__BOOTSTRAP_DATA__"`
	BootstrapData        hcl.Expression `hcl:"bootstrap_data,optional" docs:"JSON object which replaces the placeholder from {bootstrap_file} content."`
	BootstrapFile        string         `hcl:"bootstrap_file" docs:"Location of the bootstrap file."`
	CORS                 *CORS          `hcl:"cors,block" docs:"Configures [CORS](/configuration/block/cors) settings (zero or one)."`
	DisableAccessControl []string       `hcl:"disable_access_control,optional" docs:"Disables access controls by name."`
	Name                 string         `hcl:"name,label,optional"`
	Paths                []string       `hcl:"paths" docs:"List of SPA paths that need the bootstrap file."`
	Remain               hcl.Body       `hcl:",remain"`
}

// HCLBody implements the <Body> interface.
func (s Spa) HCLBody() *hclsyntax.Body {
	return s.Remain.(*hclsyntax.Body)
}

// Inline implements the <Inline> interface.
func (s Spa) Inline() interface{} {
	type Inline struct {
		meta.ResponseHeadersAttributes
		meta.LogFieldsAttribute
	}

	return &Inline{}
}

func (s Spa) Schema() *hcl.BodySchema {
	sh, _ := gohcl.ImpliedBodySchema(s)
	i, _ := gohcl.ImpliedBodySchema(s.Inline())
	return meta.MergeSchemas(sh, i, meta.ResponseHeadersAttributesSchema, meta.LogFieldsAttributeSchema)
}
