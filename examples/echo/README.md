# Echo Example

Simple echo over WebTransport at `https://localhost:4444/echo`.

## Run
```bash
# from repository root

# Start server
go run ./examples/echo/server

# In another terminal, start client
go run ./examples/echo/client
```

Notes:
- Uses a self-signed certificate and `InsecureSkipVerify` in code; configure proper TLS for production.
- QUIC options enable datagrams and 0-RTT.
