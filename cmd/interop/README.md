# Interop

Interop server and clients for testing MOQ Lite with WebTransport and QUIC.

## Run

### Using Mage (from repository root)
```bash
# Start the interop server (WebTransport + QUIC)
mage interop:server

# In another terminal, run the Go client
mage interop:client go

# Or run the TypeScript client
mage interop:client ts
```

### Using Go directly
```bash
# from repository root

# Start the interop server
go run ./cmd/interop/server

# In another terminal, run the Go client
go run ./cmd/interop/client
```

Notes:
- Uses self-signed certificates in the repo; configure proper TLS for production.
- QUIC config enables datagrams and 0-RTT for low-latency exercises.