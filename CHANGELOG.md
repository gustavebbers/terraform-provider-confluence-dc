# Changelog

## Unreleased

## v0.3.0

- Added: new `confluencedc_global_permission` resource, for granting a
  group an instance-wide permission (e.g. standard use/login access via
  `use`/`application`, or system administrator rights via
  `administer`/`system`) rather than a permission scoped to one space.
  Uses `PUT /rest/api/permissions/group/{groupName}/grant` (and
  `.../revoke`); there is no legacy JSON-RPC fallback for this one, since
  global permission management was never exposed by Confluence's legacy
  remote API.

## v0.2.1

- Fixed: `confluencedc_group` creation via REST was sending `{"name": ...}`
  without a `type` field, which POST /rest/api/admin/group rejects with a
  400 ("missing type id property 'type'") since Confluence deserializes
  the body as its polymorphic Group model. The request body now includes
  `"type": "group"`.

## v0.2.0

- Changed: `confluencedc_group` and `confluencedc_space_permission` now
  create/delete groups and grant/revoke space permissions through REST
  (`/rest/api/admin/group`,
  `/rest/api/space/{spaceKey}/permissions/group/{groupName}/grant|revoke`)
  instead of Confluence's legacy JSON-RPC API. The provider falls back to
  JSON-RPC only if those REST endpoints aren't found (e.g. on a Data Center
  release old enough to predate them), so upgrading past a release that
  disables or removes the legacy Remote API no longer breaks these two
  resources.
- Improved: on instances where the JSON-RPC fallback is needed and also
  fails, a 404/405 from it now returns an error explaining that the
  Confluence "Remote API (XML-RPC & SOAP)" admin setting also gates the
  JSON-RPC endpoint, instead of a bare, hard-to-diagnose HTTP status.

<!--
Entries are added here manually as changes are made, and moved under a
new version heading when a release is cut. There is no changelog
generation tool (e.g. changie) wired up for this repository.
-->
