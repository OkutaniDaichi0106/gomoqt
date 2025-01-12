package moqtransfork

type SubscribeGap struct {
	StartGroupSequence GroupSequence
	Count              uint64
	GroupErrorCode     GroupErrorCode
}
