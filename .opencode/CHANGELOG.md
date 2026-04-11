# Changelog

## 2026-04-11

- added a server-side session-start guard in `ClientManager` to block duplicate `startClient` launches for the same user
- updated `/session/connect` and startup restore flow to use that guard, reducing overlapping WhatsApp connections that can trigger `StreamReplaced`
- made `/session/status` nil-safe while clients are being replaced or cleaned up
- changed `StreamReplaced` handling to emit the event and schedule reconnect instead of permanently disabling recovery
- added regression tests for session-start guard behavior and verified with `go test ./...`
- added persistent `autostart` session intent so cold boot restore no longer depends on transient `users.connected`
- changed startup restore to load only paired users with `autostart=1` and non-empty `jid`, avoiding accidental resurrection of explicitly logged-out sessions
- synced `autostart` writes across connect, pair success, connected, disconnect, and logged-out flows
- added migration and regression tests for `autostart` backfill/startup user selection and re-verified with `go test ./...`
