# Changelog

## Unreleased

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
