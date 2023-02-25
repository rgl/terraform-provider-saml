---
page_title: Saml Provider
subcategory: ""
description: |-
  This provider complements the `azuread_service_principal_token_signing_certificate` resource by providing the `saml_metadata` resource to make the SAML metadata document available for further use.
---

# Saml Provider

This provider complements the [`azuread_service_principal_token_signing_certificate` resource](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/service_principal_token_signing_certificate) by providing the `saml_metadata` resource to make the SAML metadata document available for further use.

This addresses two Azure AD SAML metadata endpoint behaviors:

1. The endpoint is eventually consistent. After the token signing key is set, it takes some time until the SAML metadata endpoint actually has the expected token signing key.
    * The `saml_metadata` resource will wait until the given token signing key is available.
2. It always returns a different document for each request. The returned document only differs by its signature.
    * The `saml_metadata` resource will only return a different document when anything but the signature changes.

This is intended to be used in conjunction with the [`azuread_service_principal_token_signing_certificate` resource](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/service_principal_token_signing_certificate) to configure a [`aws_iam_saml_provider` resource](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_saml_provider).

Example usage:

```hcl
terraform {
  required_providers {
    # see https://github.com/rgl/terraform-provider-saml
    # see https://registry.terraform.io/providers/rgl/saml
    saml = {
      source  = "rgl/saml"
    }
    ...
  }
}

resource "saml_metadata" "example" {
  token_signing_key_thumbprint = azuread_service_principal_token_signing_certificate.example.thumbprint
  ...
}

resource "azuread_service_principal_token_signing_certificate" "example" {
  ...
}

resource "aws_iam_saml_provider" "example" {
  saml_metadata_document = saml_metadata.example.document
  ...
}
```

See the complete example at [example-saml-service-provider-azure terraform program](https://github.com/rgl/example-saml-service-provider-azure.git).
