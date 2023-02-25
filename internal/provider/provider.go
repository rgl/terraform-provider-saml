package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ provider.Provider = &samlProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &samlProvider{
			version: version,
		}
	}
}

type samlProvider struct {
	version string
}

func (p *samlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "saml"
	resp.Version = p.version
}

func (p *samlProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *samlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *samlProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *samlProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMetadataResource,
	}
}
