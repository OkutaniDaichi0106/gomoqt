# Native QUIC Example

Demonstrates using the QUIC transport (no WebTransport) at `moqt://localhost:4469/nativequic`.

## Run
```bash
# from repository root

# Start server (native QUIC)
go run ./examples/native_quic/server

# In another terminal, start client
go run ./examples/native_quic/client
```

Notes:
- Uses self-signed certificates and `InsecureSkipVerify` in code; secure appropriately for production.
- QUIC options enable datagrams and 0-RTT for low-latency demo.
