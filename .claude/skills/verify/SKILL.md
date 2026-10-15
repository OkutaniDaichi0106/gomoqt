---
name: verify
description: Verify a moqt change end to end through the public package, with a server and a client over real QUIC on localhost.
---

# Verifying moqt changes

moqt is a library: verify through its public package from a separate module,
not from tests inside `moqt/`.

1. Make a module outside the repo (a scratch dir) with
   `go mod init verify && go mod edit -replace github.com/qumo-dev/gomoqt=<repo path>`,
   then `go get github.com/qumo-dev/gomoqt@v0.0.0 && go mod tidy`.
2. Write two programs: a server (`moqt.Server` with `TrackMux`, and
   `AuthHandler` when auth is involved; a self-signed ECDSA cert built in code,
   `NextProtos: []string{moqt.NextProtoMOQ}`) and a client
   (`(&moqt.Dialer{TLSConfig: &tls.Config{InsecureSkipVerify: true}}).Dial(ctx, "moqt://<addr>/app", nil)`).
3. Build both with `go build -o`, start the server in the background on a free
   UDP port, run the client, then read the server log.

Gotchas:
- On Windows, pick a free UDP port each run (`UdpClient(0)` in PowerShell): a
  fixed port is often taken and the server only logs the bind error.
- Run `go vet` on the harness: it catches an `errors.As` target that is not a
  pointer error type (`*moqt.SubscribeError`, not `moqt.SubscribeError`).
- `AnnouncementReader.Announcements` yields each path once, when it becomes
  active; check an ending with `Announcement.IsActive` or `Done`.
- Make the server `log.Fatal` on `ListenAndServe`, so a failed listen ends it
  instead of leaving the client to time out.
