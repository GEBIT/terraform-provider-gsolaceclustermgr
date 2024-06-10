package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"terraform-provider-clustermanager/internal/missioncontrol"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// brokerResourceModel maps the resource schema data.
type brokerResourceModel struct {
	ID             types.String `tfsdk:"id"`
	LastUpdated    types.String `tfsdk:"last_updated"`
	Status         types.String `tfsdk:"status"`
	ServiceClassId types.String `tfsdk:"serviceclass_id"`
	Name           types.String `tfsdk:"name"`
	DataCenterId   types.String `tfsdk:"datacenter_id"`
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
	req.Header.Set("Authorization", "Bearer "+r.clientHolder.BearerToken)
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
	clientHolder MCClientHolder
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

	clientHolder, ok := req.ProviderData.(MCClientHolder)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *missioncontrol.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.clientHolder = clientHolder
}

// Schema defines the schema for the resource.
func (r *brokerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "define broker schema")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			// creation params
			"serviceclass_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"datacenter_id": schema.StringAttribute{
				Optional: true,
			},
			// computed attributes
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
		Name:           plannedState.Name.ValueString(),
		ServiceClassId: missioncontrol.ServiceClassId(plannedState.ServiceClassId.ValueString()),
		DatacenterId:   plannedState.DataCenterId.ValueString(),
	}
	tflog.Info(ctx, fmt.Sprintf("Request: %s %s %v %s using %s", "Foo", body.Name, body.ServiceClassId, body.DatacenterId, plannedState.ServiceClassId.ValueString()))

	// Use client to create new broker
	tflog.Info(ctx, fmt.Sprintf("Creating broker service using %v", body))

	createResp, err := r.clientHolder.Client.CreateServiceWithResponse(ctx, body, r.BearerReqEditorFn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating broker service",
			"Could not create broker servicer, unexpected error: "+err.Error(),
		)
		return
	}
	if createResp.StatusCode() != 202 {
		resp.Diagnostics.AddError(
			"Error creating broker service",
			fmt.Sprintf("Unexpected response code: %v", createResp.StatusCode()),
		)
		tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", createResp.Body))
		return
	}

	resourceId := *(createResp.JSON202.Data.ResourceId)
	tflog.Info(ctx, fmt.Sprintf("Waiting for broker service using %s to finish creation", resourceId))

	getParams := missioncontrol.GetServiceParams{}

	timeout := time.Now().Add(30 * time.Minute)
	for created := false; !created; {
		// sleep, timeout
		if time.Now().After(timeout) {
			resp.Diagnostics.AddError(
				"Timeout",
				"timeout creating broker service",
			)
			return
		}
		time.Sleep(30 * time.Second)
		tflog.Info(ctx, fmt.Sprintf("Checking broker status for %s", resourceId))
		getResp, err := r.clientHolder.Client.GetServiceWithResponse(ctx, resourceId, &getParams, r.BearerReqEditorFn)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting broker service",
				"Could not get broker service, unexpected error: "+err.Error(),
			)
			return
		}
		if getResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(
				"Error Checking broker status",
				fmt.Sprintf("unexpected response code: %v from ", getResp.StatusCode()),
			)
			tflog.Debug(ctx, fmt.Sprintf("CreateResponse Body:%s", getResp.Body))
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("CreationState %v", getResp.JSON200.Data.CreationState))
		if *(getResp.JSON200.Data.CreationState) == missioncontrol.ServiceCreationStateCOMPLETED {
			created = true
			// Map response body to schema and populate Computed attribute values
			plannedState.ID = types.StringValue(resourceId)
			// fake...
			plannedState.Status = types.StringValue("COMPLETED")
			plannedState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
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

	getParams := missioncontrol.GetServiceParams{}

	diags := req.State.Get(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed broker state
	getResp, err := r.clientHolder.Client.GetServiceWithResponse(ctx, currentState.ID.ValueString(), &getParams, r.BearerReqEditorFn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting broker service info",
			"Could not get broker service, unexpected error: "+err.Error(),
		)
		return
	}
	if getResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error Checking broker status",
			fmt.Sprintf("Unexpected response code: %v", getResp.StatusCode()),
		)
		tflog.Debug(ctx, fmt.Sprintf("Response Body:%s", getResp.Body))
		return
	}

	// Overwrite items with refreshed state
	currentState.LastUpdated = types.StringValue(getResp.JSON200.Data.UpdatedTime.Format(time.RFC850))
	currentState.Status = types.StringValue(string(*(getResp.JSON200.Data.CreationState)))

	// Set refreshed state
	diags = resp.State.Set(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *brokerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *brokerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state brokerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete broker

}

func (r *brokerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute

	// check this
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
