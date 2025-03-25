package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"terraform-provider-gsolaceclustermgr/internal/missioncontrol"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// brokerResourceModel maps the resource schema data.
type brokerResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	DataCenterId           types.String `tfsdk:"datacenter_id"`
	Name                   types.String `tfsdk:"name"`
	ClusterName            types.String `tfsdk:"cluster_name"`
	MsgVpnName             types.String `tfsdk:"msg_vpn_name"`
	Created                types.String `tfsdk:"created"`
	LastUpdated            types.String `tfsdk:"last_updated"`
	Status                 types.String `tfsdk:"status"`
	ServiceClassId         types.String `tfsdk:"serviceclass_id"`
	CustomRouterName       types.String `tfsdk:"custom_router_name"`
	EventBrokerVersion     types.String `tfsdk:"event_broker_version"`
	MaxSpoolUsage          types.Int32  `tfsdk:"max_spool_usage"`
	MissionControlUserName types.String `tfsdk:"missioncontrol_username"`
	MissionControlPassword types.String `tfsdk:"missioncontrol_password"`
	HostNames              types.List   `tfsdk:"hostnames"`
	ServiceEndpointId      types.String `tfsdk:"service_endpoint_id"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &brokerResource{}
	_ resource.ResourceWithConfigure   = &brokerResource{}
	_ resource.ResourceWithImportState = &brokerResource{}
)

// NewBrokerResource is a helper function to simplify the provider implementation.
func NewBrokerResource() resource.Resource {
	return &brokerResource{}
}

// helper func to add bearer token auth header to requests
func (r *brokerResource) BearerReqEditorFn(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+r.cMProviderData.BearerToken)
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		tflog.Error(ctx, err.Error())
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Request: %s", dump))
	}
	return nil
}

// brokerResource is the resource implementation.
type brokerResource struct {
	cMProviderData CMProviderData
}

// Metadata returns the resource type name.
func (r *brokerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_broker"
}

// Configure adds the provider configured client to the resource.
func (r *brokerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "configure broker resource")
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

	r.cMProviderData = cMProviderData
}

// Schema defines the schema for the resource.
func (r *brokerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "define broker schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Event Broker Resource. Note that *name* is the only attribute you can update without forcing a replacement",
		Attributes: map[string]schema.Attribute{
			// creation params
			"name": schema.StringAttribute{
				MarkdownDescription: "Broker name",
				Required:            true,
			},
			"serviceclass_id": schema.StringAttribute{
				MarkdownDescription: "Serviceclass_id like DEVELOPER, ENTERPRISE_250_STANDALONE,... (see api docs)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datacenter_id": schema.StringAttribute{
				MarkdownDescription: "the datacenter, e.g. aks-germanywestcentral-1",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// optional attributes that be filled with defaults from API server
			"msg_vpn_name": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"cluster_name": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"custom_router_name": schema.StringAttribute{
				MarkdownDescription: "Custom Router Name prefix (the actual routername will be suffixed with primary (if generated) or primarycn",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 12),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]+$`),
						"must contain only lowercase letters and digits",
					),
				},
			},
			"event_broker_version": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			// figure out how to handle int32
			"max_spool_usage": schema.Int32Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The message spool size, in gigabytes (GB)",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.Int32{
					int32validator.Between(10, 6000),
				},
			},
			//
			// computed attributes
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"hostnames": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"service_endpoint_id": schema.StringAttribute{
				Computed: true,
			},
			"missioncontrol_username": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"missioncontrol_password": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

// Create a new resource.
func (r *brokerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Retrieving planned state")
	// Retrieve values from plannedState
	var plannedState brokerResourceModel
	diags := req.Plan.Get(ctx, &plannedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body = missioncontrol.CreateServiceJSONRequestBody{
		Name:               plannedState.Name.ValueString(),
		ServiceClassId:     missioncontrol.ServiceClassId(plannedState.ServiceClassId.ValueString()),
		DatacenterId:       plannedState.DataCenterId.ValueString(),
		MsgVpnName:         nullIfEmptyStringPtr(plannedState.MsgVpnName),
		ClusterName:        nullIfEmptyStringPtr(plannedState.ClusterName),
		EventBrokerVersion: nullIfEmptyStringPtr(plannedState.EventBrokerVersion),
		CustomRouterName:   nullIfEmptyStringPtr(plannedState.CustomRouterName),
		MaxSpoolUsage:      nullIfEmptyInt32Ptr(plannedState.MaxSpoolUsage),
	}
	tflog.Info(ctx, fmt.Sprintf("Request: %s %s %v %s using %s", "Foo", body.Name, body.ServiceClassId, body.DatacenterId, plannedState.ServiceClassId.ValueString()))

	// Use client to create new broker
	tflog.Info(ctx, fmt.Sprintf("Creating broker service using %v", body))

	createResp, err := r.cMProviderData.Client.CreateServiceWithResponse(ctx, body, r.BearerReqEditorFn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating broker service",
			"Could not create broker servicer, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Response Header:%s", createResp.HTTPResponse.Header))
	tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", createResp.Body))
	// As of 20250324 the errorResponse is an application/xml error object, it will not be mapped to json"
	if createResp.StatusCode() == 400 {
		var errMsg string
		if createResp.JSON400 == nil {
			errMsg = parseErrorDTO(createResp.Body)
		} else {
			errMsg = *(createResp.JSON400.Message)
			if createResp.JSON400.ValidationDetails != nil {
				errMsg = errMsg + fmt.Sprintf("\nValidation Error: %v", *(createResp.JSON400.ValidationDetails))
			}
		}
		resp.Diagnostics.AddError(
			"Error creating broker service",
			errMsg,
		)
		return
	}
	if createResp.StatusCode() == 401 {
		var errMsg string
		if createResp.JSON401 == nil {
			errMsg = parseErrorDTO(createResp.Body)
		} else {
			errMsg = *(createResp.JSON401.Message)
		}
		resp.Diagnostics.AddError(
			"Error creating broker service",
			errMsg,
		)
		return
	}

	if createResp.StatusCode() == 403 {
		var errMsg string
		if createResp.JSON403 == nil {
			errMsg = parseErrorDTO(createResp.Body)
		} else {
			errMsg = *(createResp.JSON403.Message)
		}
		resp.Diagnostics.AddError(
			"Error creating broker service",
			errMsg,
		)
		return
	}
	if createResp.StatusCode() == 503 {
		var errMsg string
		if createResp.JSON503 == nil {
			errMsg = parseErrorDTO(createResp.Body)
		} else {
			errMsg = *(createResp.JSON503.Message)
			if createResp.JSON503.ValidationDetails != nil {
				errMsg = errMsg + fmt.Sprintf("\nValidation Error: %v", *(createResp.JSON503.ValidationDetails))
			}
		}
		resp.Diagnostics.AddError(
			"Error creating broker service",
			errMsg,
		)
		return
	}

	if createResp.StatusCode() != 202 {
		resp.Diagnostics.AddError(
			"Error creating broker service",
			fmt.Sprintf("Unexpected response code: %v", createResp.StatusCode()),
		)
		return
	}

	resourceId := *(createResp.JSON202.Data.ResourceId)

	tflog.Info(ctx, fmt.Sprintf("Waiting for broker service using %s to finish creation", resourceId))

	// TODO: polling GET with full expansion is maybe expensive - we could poll for the operation and fetch the full state once instead

	timeout := time.Now().Add(r.cMProviderData.PollingTimeoutDuration)
	for created := false; !created; {
		// sleep, timeout
		if time.Now().After(timeout) {
			resp.Diagnostics.AddError(
				"Timeout",
				"timeout creating broker service",
			)
			return
		}
		time.Sleep(r.cMProviderData.PollingIntervalDuration)
		tflog.Info(ctx, fmt.Sprintf("Checking broker status for %s", resourceId))

		r.fullGet(ctx, resourceId, &plannedState, &resp.Diagnostics)

		tflog.Info(ctx, fmt.Sprintf("Broker status %s", plannedState.Status.ValueString()))

		if resp.Diagnostics.HasError() {
			return
		}

		if plannedState.Status.ValueString() == string(missioncontrol.ServiceCreationStateCOMPLETED) {
			created = true
		}

	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plannedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *brokerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "retrieve current state")
	// Get current currentState
	var currentState brokerResourceModel

	diags := req.State.Get(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed broker state
	r.fullGet(ctx, currentState.ID.ValueString(), &currentState, &resp.Diagnostics)
	if resp.Diagnostics.WarningsCount() > 0 {
		if resp.Diagnostics.Warnings()[0].Summary() == "404:VANISHED" {
			tflog.Info(ctx, "Removing vanished resource from state gracefully")
			resp.State.RemoveResource(ctx)
			return
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *brokerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Retrieving planned state")
	// Retrieve values from plannedState
	var plannedState brokerResourceModel
	diags := req.Plan.Get(ctx, &plannedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body = missioncontrol.UpdateServiceJSONRequestBody{
		Name: plannedState.Name.ValueStringPointer(),
	}
	brokerId := plannedState.ID.ValueString()

	// Use client to update broker
	tflog.Info(ctx, fmt.Sprintf("Updating broker service using %v", body))

	updateResp, err := r.cMProviderData.Client.UpdateServiceWithResponse(ctx, brokerId, body, r.BearerReqEditorFn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating broker service",
			"Could not update broker service, unexpected error: "+err.Error(),
		)
		return
	}

	// NOTE: in theory we will get a PENDING or INPROGRESS status, and should wait for the operatin to finish.
	// It is only a quick renaming however, so we do not bother...
	tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", updateResp.Body))
	if updateResp.StatusCode() != 200 {
		// do not catch 404 (vanished resources), that is an error
		resp.Diagnostics.AddError(
			"Error creating broker service",
			fmt.Sprintf("Unexpected response code: %v", updateResp.StatusCode()),
		)
		return
	}

	// Update will NOT deliver expanded infos (epand query param is not specified for this method)
	// Therfore we get the full info again
	// Get refreshed broker state
	r.fullGet(ctx, plannedState.ID.ValueString(), &plannedState, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Updated broker to %s %v %v", plannedState.Name.ValueString(), plannedState.Status.ValueString(), plannedState.LastUpdated.ValueString()))

	// Save updated data into Terraform state
	diags = resp.State.Set(ctx, &plannedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *brokerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var currentState brokerResourceModel
	diags := req.State.Get(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// then delete
	brokerId := currentState.ID.ValueString()
	delResp, err := r.cMProviderData.Client.DeleteServiceWithResponse(ctx, brokerId, r.BearerReqEditorFn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting broker service info",
			"Could not get broker service, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", delResp.Body))
	if delResp.StatusCode() != 202 {

		// handling a vanished resource (likely already detected in plan/read)
		if delResp.StatusCode() == 404 {
			// As of 20250127 the response is not as specified, so we cannot use getResp.JSON404
			if strings.Contains(string(delResp.Body), "Could not find event broker service with id") {
				tflog.Warn(ctx, "Could not find event broker service")
				// this is tolerable!
				return
			}
		}
		resp.Diagnostics.AddError(
			"Error Checking broker status",
			fmt.Sprintf("Unexpected response code: %v", delResp.StatusCode()),
		)
		tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", delResp.Body))
		return
	}
	operationId := *(delResp.JSON202.Data.Id)
	tflog.Debug(ctx, fmt.Sprintf("Delete-Operation %s on broker %s has been started.", operationId, brokerId))

	// TODO we might want to wait for the operation to finish?
}

func (r *brokerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute

	// check this
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// helper to fully retrieve brokerInfos
func (r *brokerResource) fullGet(ctx context.Context, id string, model *brokerResourceModel, diagnostics *diag.Diagnostics) {
	var diags diag.Diagnostics

	getParams := missioncontrol.GetServiceParams{
		Expand: &[]missioncontrol.GetServiceParamsExpand{"broker,serviceConnectionEndpoints"},
	}

	// Get refreshed broker state
	getResp, err := r.cMProviderData.Client.GetServiceWithResponse(ctx, id, &getParams, r.BearerReqEditorFn)
	if err != nil {
		diagnostics.AddError(
			"Error getting broker service",
			"Could not get broker service, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", getResp.Body))
	if getResp.StatusCode() != 200 {

		// handle vanished resources
		if getResp.StatusCode() == 404 {
			// As of 20250127 the response is not as specified, so we cannot use getResp.JSON404
			if strings.Contains(string(getResp.Body), "Could not find event broker service with id") {
				tflog.Warn(ctx, "Could not find event broker service")
				// this might be tolerable
				diagnostics.AddWarning(
					"404:VANISHED",
					"Could not find event broker service with id "+id)
				return
			}
		}
		if getResp.StatusCode() == 401 {
			var errMsg string
			if getResp.JSON401 == nil {
				errMsg = parseErrorDTO(getResp.Body)
			} else {
				errMsg = *(getResp.JSON401.Message)
			}
			diagnostics.AddError(
				"Error getting broker service",
				errMsg,
			)
			return
		}

		if getResp.StatusCode() == 403 {
			var errMsg string
			if getResp.JSON403 == nil {
				errMsg = parseErrorDTO(getResp.Body)
			} else {
				errMsg = *(getResp.JSON403.Message)
			}
			diagnostics.AddError(
				"Error getting broker service",
				errMsg,
			)
			return
		}

		diagnostics.AddError(
			"Error getting broker service",
			fmt.Sprintf("Unexpected response code: %v", getResp.StatusCode()),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", getResp.Body))
	// extract all infos when status is COMPLETED
	if *(getResp.JSON200.Data.CreationState) == missioncontrol.ServiceCreationStateCOMPLETED {
		model.ID = types.StringPointerValue(getResp.JSON200.Data.Id)
		if getResp.JSON200.Data.CreatedTime != nil {
			model.Created = types.StringValue(getResp.JSON200.Data.CreatedTime.Format(time.RFC850))
		} else {
			model.Created = types.StringValue("")
		}
		if getResp.JSON200.Data.UpdatedTime != nil {
			model.LastUpdated = types.StringValue(getResp.JSON200.Data.UpdatedTime.Format(time.RFC850))
		} else {
			model.LastUpdated = types.StringValue("")
		}
		model.ServiceClassId = types.StringPointerValue((*string)(getResp.JSON200.Data.ServiceClassId))
		model.DataCenterId = types.StringPointerValue(getResp.JSON200.Data.DatacenterId)
		model.EventBrokerVersion = types.StringValue(getResp.JSON200.Data.EventBrokerServiceVersion)
		model.Status = types.StringValue(string(*(getResp.JSON200.Data.CreationState)))
		model.Name = types.StringPointerValue(getResp.JSON200.Data.Name)
		model.ClusterName = types.StringPointerValue(getResp.JSON200.Data.Broker.Cluster.Name)

		model.CustomRouterName = types.StringValue(getRouterPrefix(*(getResp.JSON200.Data.Broker.Cluster.PrimaryRouterName)))
		model.MsgVpnName = types.StringPointerValue((*(getResp.JSON200.Data.Broker.MsgVpns))[0].MsgVpnName)
		model.MaxSpoolUsage = types.Int32PointerValue(getResp.JSON200.Data.Broker.MaxSpoolUsage)
		model.MissionControlUserName = types.StringPointerValue((*(getResp.JSON200.Data.Broker.MsgVpns))[0].MissionControlManagerLoginCredential.Username)
		model.MissionControlPassword = types.StringPointerValue((*(getResp.JSON200.Data.Broker.MsgVpns))[0].MissionControlManagerLoginCredential.Password)
		model.ServiceEndpointId = types.StringPointerValue((*getResp.JSON200.Data.ServiceConnectionEndpoints)[0].Id)
		hostNames := (*getResp.JSON200.Data.ServiceConnectionEndpoints)[0].HostNames

		model.HostNames, diags = types.ListValueFrom(ctx, types.StringType, hostNames)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, fmt.Sprintf("Read Broker state %s %s %s %v", model.ID, model.Name, model.Status.ValueString(), model.LastUpdated))
	}

}
