package message

type GroupErrorCode uint64

type GroupDrop struct {
	GroupStartSequence uint64
	Count              uint64
	GroupErrorCode     GroupErrorCode
}
