package moqt

type Relayer interface {
}

func NewRelayer(manager *RelayManager, sess ServerSession) Relayer {
	return nil
}
