package provider

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/crewjam/saml/samlsp"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getMetadata(ctx context.Context, url, tokenSigningKeyThumbprint, currentMetadata string) (string, error) {
	// Download the metadata document.
	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return "", fmt.Errorf("response failed with status code %d and body: %s", res.StatusCode, body)
	}
	// Find the signing key that has the tokenSigningKeyThumbprint thumbprint.
	metadata, err := samlsp.ParseMetadata(body)
	if err != nil {
		return "", err
	}
	foundTokenSigningKey := false
	expectedThumbprint, err := hex.DecodeString(tokenSigningKeyThumbprint)
	if err != nil {
		return "", fmt.Errorf("failed to decode tokenSigningKeyThumbprint: %w", err)
	}
	for _, d := range metadata.IDPSSODescriptors {
		for _, kd := range d.KeyDescriptors {
			if kd.Use == "signing" {
				for _, c := range kd.KeyInfo.X509Data.X509Certificates {
					b, err := base64.StdEncoding.DecodeString(c.Data)
					if err != nil {
						continue
					}
					thumbprint := sha1.Sum(b)
					if bytes.Equal(thumbprint[:], expectedThumbprint) {
						foundTokenSigningKey = true
					}
				}
			}
		}
	}
	if !foundTokenSigningKey {
		return "", fmt.Errorf("could not find the token signing key")
	}
	// Return the current metadata when they only differ by their signature.
	// NB this is required because each time we request a metadata document from
	//    azure ad it always generates a new signature, which is quite annoying
	//    for terraform diff.
	if currentMetadata != "" {
		parsedCurrentMetadata, err := samlsp.ParseMetadata([]byte(currentMetadata))
		if err != nil {
			return "", fmt.Errorf("failed to parse the current metadata: %w", err)
		}
		parsedCurrentMetadata.ID = metadata.ID
		parsedCurrentMetadata.Signature = metadata.Signature
		currentMetadataXml, err := xml.Marshal(parsedCurrentMetadata)
		if err != nil {
			return "", fmt.Errorf("failed to marshal the current metadata: %w", err)
		}
		metadataXml, err := xml.Marshal(metadata)
		if err != nil {
			return "", fmt.Errorf("failed to marshal the metadata: %w", err)
		}
		if bytes.Equal(currentMetadataXml, metadataXml) {
			return currentMetadata, nil
		}
	}
	return string(body), nil
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &metadataResource{}
	_ resource.ResourceWithConfigure   = &metadataResource{}
	_ resource.ResourceWithImportState = &metadataResource{}
)

func NewMetadataResource() resource.Resource {
	return &metadataResource{}
}

type metadataResource struct {
}

type metadataResourceModel struct {
	URL                       types.String `tfsdk:"url"`
	TokenSigningKeyThumbprint types.String `tfsdk:"token_signing_key_thumbprint"`
	Document                  types.String `tfsdk:"document"`
}

func (r *metadataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metadata"
}

func (r *metadataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Caches an SAML IDP Metadata document.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "SAML IDP Metadata document URL.",
				Required:    true,
			},
			"token_signing_key_thumbprint": schema.StringAttribute{
				Description: "Token signing key thumbprint.",
				Required:    true,
			},
			"document": schema.StringAttribute{
				Description: "The SAML IDP Metadata document.",
				Computed:    true,
			},
		},
	}
}

func (r *metadataResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

func (r *metadataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan metadataResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed state.
	metadata, err := getMetadata(ctx, plan.URL.ValueString(), plan.TokenSigningKeyThumbprint.ValueString(), "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating metadata",
			"Could not create metadata, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Document = types.StringValue(metadata)

	// Set values into plan.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *metadataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state metadataResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed state.
	metadata, err := getMetadata(ctx, state.URL.ValueString(), state.TokenSigningKeyThumbprint.ValueString(), state.Document.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading metadata",
			"Could not read metadata, unexpected error: "+err.Error(),
		)
		return
	}
	state.Document = types.StringValue(metadata)

	// Set refreshed state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *metadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan metadataResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed state.
	metadata, err := getMetadata(ctx, plan.URL.ValueString(), plan.TokenSigningKeyThumbprint.ValueString(), plan.Document.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating metadata",
			"Could not update metadata, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Document = types.StringValue(metadata)

	// Set refreshed state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *metadataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *metadataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
