package gomoq

// Status Codes
/*
 * Status code of the track
 */
type TrackStatusCode byte

func IN_PROGRESS() TrackStatusCode {
	return TrackStatusCode(0x00)
}
func NOT_EXIST() TrackStatusCode {
	return TrackStatusCode(0x01)
}
func NOT_BEGUN_YET() TrackStatusCode {
	return TrackStatusCode(0x02)
}
func FINISHED() TrackStatusCode {
	return TrackStatusCode(0x03)
}
func RELAY() TrackStatusCode {
	return TrackStatusCode(0x04)
}

/*
 * Status code of the "SUBSCRIBE_DONE" message
 */
type SubscribeDoneStatusCode int

func (SubscribeDoneMessage) UNSUBSCRIBED() SubscribeDoneStatusCode {
	return SubscribeDoneStatusCode(0x0)
}
func (SubscribeDoneMessage) INTERNAL_ERROR() SubscribeDoneStatusCode {
	return SubscribeDoneStatusCode(0x1)
}
func (SubscribeDoneMessage) UNAUTHORIZED() SubscribeDoneStatusCode {
	return SubscribeDoneStatusCode(0x2)
}
func (SubscribeDoneMessage) TRACK_ENDED() SubscribeDoneStatusCode {
	return SubscribeDoneStatusCode(0x3)
}
func (SubscribeDoneMessage) SUBSCRIPTION_ENDED() SubscribeDoneStatusCode {
	return SubscribeDoneStatusCode(0x4)
}
func (SubscribeDoneMessage) GOING_AWAY() SubscribeDoneStatusCode {
	return SubscribeDoneStatusCode(0x5)
}
func (SubscribeDoneMessage) EXPIRED() SubscribeDoneStatusCode {
	return SubscribeDoneStatusCode(0x6)
}
