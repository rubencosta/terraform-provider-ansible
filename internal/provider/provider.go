package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &terraformProviderAnsible{
			version: version,
		}
	}
}

type terraformProviderAnsible struct {
	version string
}

func (p *terraformProviderAnsible) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ansible"
	resp.Version = p.version
}

func (p *terraformProviderAnsible) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *terraformProviderAnsible) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *terraformProviderAnsible) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPlaybookResource,
	}
}

func (p *terraformProviderAnsible) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}
