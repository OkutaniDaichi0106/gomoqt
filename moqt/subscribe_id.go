package moqt

import "fmt"

type SubscribeID uint64

func (id SubscribeID) String() string {
	return fmt.Sprintf("SubscribeID: %d", id)
}
