package moqt

import (
	"fmt"
)

var _ GroupError = (*SubscribeGap)(nil)

type SubscribeGap struct {
	start GroupSequence
	count uint64
	code  GroupErrorCode
}

func (sg SubscribeGap) Error() string {
	return fmt.Sprintf(
		"failed to deliver groups: missing groups in the range [%d, %d) (error code: %d)",
		sg.start, sg.start+GroupSequence(sg.count), sg.code,
	)
}

func (sg SubscribeGap) GroupErrorCode() GroupErrorCode {
	return sg.code
}

func (sg SubscribeGap) String() string {
	return fmt.Sprintf("SubscribeGap: { MinGapSequence: %d, MaxGapSequence: %d, GroupErrorCode: %d }", sg.start, sg.count, sg.code)
}
