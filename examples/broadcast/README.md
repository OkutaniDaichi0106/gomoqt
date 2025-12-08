# Broadcast Example

Demonstrates publishing a broadcast track over WebTransport on `https://localhost:4469/broadcast`.

## Run

```bash
# from repository root

# Start server (WebTransport)
go run ./examples/broadcast/server

# In another terminal, start client
go run ./examples/broadcast/client
```

Notes:
- Uses a self-signed certificate and `InsecureSkipVerify` in the example code; configure proper TLS for production.
- QUIC options enable datagrams and 0-RTT for low-latency demo.
