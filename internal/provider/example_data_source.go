package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/redtigercod4/terraform-provider-pulp/openapi"
)

var _ datasource.DataSource = &StatusDataSource{}

func NewStatusDataSource() datasource.DataSource {
	return &StatusDataSource{}
}

type StatusDataSource struct {
	client *http.Client
}

type statusDataSourceModel struct {
	Versions              []statusDataSourceModelVersion `tfsdk:"versions"`
}

type statusDataSourceModelVersion struct {
	Component types.String `tfsdk:"component"`
	Version types.String `tfsdk:"version"`
	Package types.String `tfsdk:"package"`
	Module types.String `tfsdk:"module"`
	DomainCompatible types.Bool `tfsdk:"domain_compatible"`
}

func (d *StatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status"
}

func (d *StatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"versions": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"component": schema.StringAttribute{
							Computed: true,
						},
						"version": schema.StringAttribute{
							Computed: true,
						},
						"package": schema.StringAttribute{
							Computed: true,
						},
						"module": schema.StringAttribute{
							Computed: true,
						},
						"domain_compatible": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *StatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *StatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data statusDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := sdk.StatusAPIService{}
	res, _, err := svc.StatusReadExecute(svc.StatusRead(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading api status",
			err.Error(),
		)
	}

	data.Versions = Map(res.Versions, func(dto sdk.VersionResponse) statusDataSourceModelVersion {
		return statusDataSourceModelVersion{
			Component: types.StringValue(dto.Component),
			Version: types.StringValue(dto.Version),
			Package: types.StringValue(dto.Package),
			Module: types.StringValue(dto.Module),
			DomainCompatible: types.BoolValue(dto.DomainCompatible),
		}
	})

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}


func Map[T, U any](items []T, actor func(T) U) []U {
	result := make([]U, 0, len(items))
	for _, t := range items {
		u := actor(t)
		result = append(result, u)
	}
	return result
}