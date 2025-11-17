package moqt

// SubscribeID uniquely identifies a subscription within the session.
// It is used to correlate subscription-related messages (e.g., groups, updates)
// between client and server.
type SubscribeID uint64
