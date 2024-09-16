package moqtmessage

/*
 * Status code of the "SUBSCRIBE_DONE" message
 */
type SubscribeDoneStatusCode int

const (
	SUBSCRIBE_DONE_STATUS_UNSUBSCRIBED       = 0x00
	SUBSCRIBE_DONE_STATUS_INTERNAL_ERROR     = 0x01
	SUBSCRIBE_DONE_STATUS_UNAUTHORIZED       = 0x02
	SUBSCRIBE_DONE_STATUS_TRACK_ENDED        = 0x03
	SUBSCRIBE_DONE_STATUS_SUBSCRIPTION_ENDED = 0x04
	SUBSCRIBE_DONE_STATUS_GOING_AWAY         = 0x05
	SUBSCRIBE_DONE_STATUS_EXPIRED            = 0x06
)
