# Protocol Specification [Version: Development (0xfeedbabe)]

This implementation is based on **moq-lite-draft-01** with the following differences:

- The `SUBSCRIBE_OK` message does not include a Publish Priority field
- Message Length is encoded as a big-endian uint16 instead of QUIC variable-length integer (maximum: 65,535 bytes)

## Reference

[Media over QUIC - Lite Draft 01](https://datatracker.ietf.org/doc/html/draft-ietf-moq-lite-01)

