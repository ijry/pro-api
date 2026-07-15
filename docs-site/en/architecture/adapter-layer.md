---
title: Adapter Layer
outline: deep
---

# Adapter Layer

> :construction: This page is not translated yet. Please refer to the [Chinese version](/zh/architecture/adapter-layer) for now. Full English translation is planned for **M2**.

English content TBD in M2. The Chinese page covers:

- Section headings (H2 / H3) defined in the spec
- Code examples (bash / Python / Go / YAML)
- Tables and diagrams

## Grok Providers

- `grok-build`: xAI official OpenAI-compatible API. Store the xAI API key in the channel/account API key field.
- `grok-web`: Grok Web reverse adapter. Store the bare SSO token in the channel/account API key field; the adapter sends it as `sso` and `sso-rw` cookies.

Translating this page is tracked in the [M2 documentation roadmap](https://github.com/ijry/pro-api/issues).
