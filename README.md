# About

[![build](https://github.com/rgl/terraform-provider-saml/actions/workflows/build.yml/badge.svg)](https://github.com/rgl/terraform-provider-saml/actions/workflows/build.yml)
[![terraform provider](https://img.shields.io/badge/terraform%20provider-rgl%2Fsaml-blue)](https://registry.terraform.io/providers/rgl/saml)

This caches a stable SAML document in the terraform state because the Azure AD SAML federation metadata endpoint always returns a different document for each request.

This is intended to be used in conjunction with the [`azuread_service_principal_token_signing_certificate` resource](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/service_principal_token_signing_certificate) to configure a [`aws_iam_saml_provider` resource](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_saml_provider).

Example usage:

```hcl
terraform {
  required_providers {
    # see https://github.com/rgl/terraform-provider-saml
    # see https://registry.terraform.io/providers/rgl/saml
    saml = {
      source  = "rgl/saml"
      version = "0.1.0"
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

# Build

```bash
make build
```

Set your `~/.terraformrc` file to point to this binary:

```bash
cat >~/.terraformrc <<EOF
provider_installation {
  dev_overrides {
    "registry.terraform.io/rgl/saml" = "$PWD"
  }
  direct {
  }
}
EOF
```

Execute the [example-saml-service-provider-azure terraform program](https://github.com/rgl/example-saml-service-provider-azure.git).

And play with:

```bash
wget -qO- "$(terraform output -raw saml_metadata_url)" >idp-metadata.xml
terraform show -json | jq -r '.values.root_module.resources[] | select(.address=="saml_metadata.example") | .values.document' >idp-metadata-state.xml
terraform destroy -target azuread_service_principal_token_signing_certificate.example
terraform destroy -target saml_metadata.example
```

