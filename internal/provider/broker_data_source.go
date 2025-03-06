package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"terraform-provider-gsolaceclustermgr/internal/missioncontrol"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type brokerDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	DataCenterId       types.String `tfsdk:"datacenter_id"`
	Name               types.String `tfsdk:"name"`
	ClusterName        types.String `tfsdk:"cluster_name"`
	MsgVpnName         types.String `tfsdk:"msg_vpn_name"`
	Created            types.String `tfsdk:"created"`
	LastUpdated        types.String `tfsdk:"last_updated"`
	Status             types.String `tfsdk:"status"`
	ServiceClassId     types.String `tfsdk:"serviceclass_id"`
	CustomRouterName   types.String `tfsdk:"custom_router_name"`
	EventBrokerVersion types.String `tfsdk:"event_broker_version"`
	MaxSpoolUsage      types.Int32  `tfsdk:"max_spool_usage"`
	ClientUsername     types.String `tfsdk:"client_username"`
	ClientSecret       types.String `tfsdk:"client_secret"`

	// TODO what about messagebroker hostname / alias, anything else?
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &brokerDataSource{}
	_ datasource.DataSourceWithConfigure = &brokerDataSource{}
)

// helper func to add bearer token auth header to requests
func (d *brokerDataSource) BearerReqEditorFn(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+d.cMProviderData.BearerToken)
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		tflog.Error(ctx, err.Error())
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Request: %s", dump))
	}
	return nil
}

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewBrokerDataSource() datasource.DataSource {
	return &brokerDataSource{}
}

// coffeesDataSource is the data source implementation.
type brokerDataSource struct {
	cMProviderData CMProviderData
}

// Metadata returns the data source type name.
func (d *brokerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_broker"
}

// Schema defines the schema for the data source.
func (d *brokerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"serviceclass_id": schema.StringAttribute{
				Computed: true,
			},
			"datacenter_id": schema.StringAttribute{
				Computed: true,
			},
			// optional attributes that be filled with defaults from API server
			"msg_vpn_name": schema.StringAttribute{
				Computed: true,
			},
			"cluster_name": schema.StringAttribute{
				Computed: true,
			},
			"custom_router_name": schema.StringAttribute{
				Computed: true,
			},
			"event_broker_version": schema.StringAttribute{
				Computed: true,
			},
			// figure out how to handle int32
			"max_spool_usage": schema.Int32Attribute{
				Computed: true,
			},

			"created": schema.StringAttribute{
				Computed: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"client_username": schema.StringAttribute{
				Computed: true,
			},
			"client_secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

// Read resource information.
// resource.ReadRequest
func (d *brokerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var currentState brokerDataSourceModel

	var queryID types.String

	tflog.Info(ctx, fmt.Sprintf("RequestCfg: %v", req.Config))

	diags := req.Config.GetAttribute(ctx, path.Root("id"), &queryID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Query for broker Id: %v", queryID))

	getParams := missioncontrol.GetServiceParams{
		Expand: &[]missioncontrol.GetServiceParamsExpand{"broker"},
	}

	// Get broker info
	getResp, err := d.cMProviderData.Client.GetServiceWithResponse(ctx, queryID.ValueString(), &getParams, d.BearerReqEditorFn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting broker service info",
			"Could not get broker service, unexpected error: "+err.Error(),
		)
		return
	}
	if getResp.StatusCode() != 200 {
		tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", getResp.Body))

		// handle vanished resources
		if getResp.StatusCode() == 404 {
			// As of 20250127 the response is not as specified, so we cannot use getResp.JSON404
			if strings.Contains(string(getResp.Body), "Could not find event broker service with id") {
				tflog.Warn(ctx, "Could not find event broker service")
				// refresh state
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError(
			"Error getting broker service info",
			fmt.Sprintf("Unexpected response code: %v", getResp.StatusCode()),
		)
		return
	}

	// map to response state
	tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", getResp.Body))
	currentState.ID = types.StringPointerValue(getResp.JSON200.Data.Id)
	currentState.ServiceClassId = types.StringPointerValue((*string)(getResp.JSON200.Data.ServiceClassId))
	currentState.DataCenterId = types.StringPointerValue(getResp.JSON200.Data.DatacenterId)
	currentState.EventBrokerVersion = types.StringValue(getResp.JSON200.Data.EventBrokerServiceVersion)
	if getResp.JSON200.Data.CreatedTime != nil {
		currentState.Created = types.StringValue(getResp.JSON200.Data.CreatedTime.Format(time.RFC850))
	}
	if getResp.JSON200.Data.UpdatedTime != nil {
		currentState.LastUpdated = types.StringValue(getResp.JSON200.Data.UpdatedTime.Format(time.RFC850))
	}
	currentState.Status = types.StringValue(string(*(getResp.JSON200.Data.CreationState)))
	currentState.Name = types.StringPointerValue(getResp.JSON200.Data.Name)
	currentState.ClusterName = types.StringPointerValue(getResp.JSON200.Data.Broker.Cluster.Name)
	routerPrefix, _ := strings.CutSuffix(*(getResp.JSON200.Data.Broker.Cluster.PrimaryRouterName), "primary")
	currentState.CustomRouterName = types.StringValue(routerPrefix)
	currentState.MsgVpnName = types.StringPointerValue((*(getResp.JSON200.Data.Broker.MsgVpns))[0].MsgVpnName)
	currentState.MaxSpoolUsage = types.Int32PointerValue(getResp.JSON200.Data.Broker.MaxSpoolUsage)
	currentState.ClientUsername = types.StringPointerValue((*(getResp.JSON200.Data.Broker.MsgVpns))[0].ServiceLoginCredential.Username)
	currentState.ClientSecret = types.StringPointerValue((*(getResp.JSON200.Data.Broker.MsgVpns))[0].ServiceLoginCredential.Password)

	tflog.Debug(ctx, fmt.Sprintf("Read Broker state %s %s %v", currentState.Name, currentState.Status.ValueString(), currentState.LastUpdated))

	// Set state
	diags2 := resp.State.Set(ctx, &currentState)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *brokerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Info(ctx, "configure broker datasource")
	if req.ProviderData == nil {
		return
	}

	cMProviderData, ok := req.ProviderData.(CMProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *missioncontrol.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.cMProviderData = cMProviderData
}
