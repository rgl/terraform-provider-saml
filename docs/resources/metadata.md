---
page_title: "saml_metadata Resource - saml"
subcategory: ""
description: |-
  Caches an SAML IDP Metadata document.
---

# saml_metadata (Resource)

Caches an SAML IDP Metadata document.

## Schema

### Required

- `token_signing_key_thumbprint` (String) Token signing key thumbprint.
- `url` (String) SAML IDP Metadata document URL.

### Read-Only

- `document` (String) The SAML IDP Metadata document.
