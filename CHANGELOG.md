# Changelog

## Unreleased

- Improved: JSON-RPC 404/405 responses now return an error explaining that
  the Confluence "Remote API (XML-RPC & SOAP)" admin setting also gates the
  JSON-RPC endpoint `confluencedc_group` and `confluencedc_space_permission`
  depend on, instead of a bare, hard-to-diagnose HTTP status.

<!--
Entries are added here manually as changes are made, and moved under a
new version heading when a release is cut. There is no changelog
generation tool (e.g. changie) wired up for this repository.
-->
