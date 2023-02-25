---
page_title: "saml_metadata Resource - saml"
subcategory: ""
description: |-
  Download an SAML IDP metadata document and wait until the token signing key is available in the metadata document.
---

# saml_metadata (Resource)

Download an SAML IDP metadata document and wait until the token signing key is available in the metadata document.

This addresses two Azure AD SAML metadata endpoint behaviors:

1. The endpoint is eventually consistent. After the token signing key is set, it takes some time until the SAML metadata endpoint actually has the expected token signing key.
    * The `saml_metadata` resource will wait until the given token signing key is available.
2. It always returns a different document for each request. The returned document only differs by its signature.
    * The `saml_metadata` resource will only return a different document when anything but the signature changes.

## Schema

### Required

- `token_signing_key_thumbprint` (String) Token signing key thumbprint.
- `url` (String) SAML IDP Metadata document URL.

### Read-Only

- `document` (String) The SAML IDP Metadata document.
