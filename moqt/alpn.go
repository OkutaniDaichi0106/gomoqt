package moqt

// NextProtoMOQ is the application layer protocol negotiation string used for
// MOQ over QUIC. It is used during TLS/ALPN negotiation to select the MOQ
// protocol for native QUIC sessions.
const NextProtoMOQ = "moq-00"

// NextProtoH3 is the ALPN token used to indicate HTTP/3 (used for WebTransport).
const NextProtoH3 = "h3"
