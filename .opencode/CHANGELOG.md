# Changelog

## 2026-04-11

- added a server-side session-start guard in `ClientManager` to block duplicate `startClient` launches for the same user
- updated `/session/connect` and startup restore flow to use that guard, reducing overlapping WhatsApp connections that can trigger `StreamReplaced`
- made `/session/status` nil-safe while clients are being replaced or cleaned up
- changed `StreamReplaced` handling to emit the event and schedule reconnect instead of permanently disabling recovery
- added regression tests for session-start guard behavior and verified with `go test ./...`
