package moqt

type downstream struct {
	sess                    ServerSession
	receiveSubscribeStreams []receiveSubscribeStream
}
