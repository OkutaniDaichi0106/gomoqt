# Relay Example

Demonstrates relaying tracks via WebTransport on `https://moqt.example.com:9000/hang` (default in example code).

## Run
```bash
# from repository root

# Start relay server
# (listens on moqt.example.com:9000 with self-signed cert in example code)
go run ./examples/relay/server

# In another terminal, start relay client
go run ./examples/relay/client
```

Notes:
- The sample uses a self-signed certificate, `InsecureSkipVerify`, and a placeholder host `moqt.example.com:9000`; adjust host/TLS for your environment (e.g., use localhost and a valid cert).
- QUIC options enable datagrams and 0-RTT.
