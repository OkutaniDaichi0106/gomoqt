package moqtmessage

/*
 * Status code of the "SUBSCRIBE_DONE" message
 */
type SubscribeDoneStatusCode int

const (
	SUBSCRIBE_DONE_UNSUBSCRIBED       SubscribeDoneStatusCode = 0x00
	SUBSCRIBE_DONE_INTERNAL_ERROR     SubscribeDoneStatusCode = 0x01
	SUBSCRIBE_DONE_UNAUTHORIZED       SubscribeDoneStatusCode = 0x02
	SUBSCRIBE_DONE_TRACK_ENDED        SubscribeDoneStatusCode = 0x03
	SUBSCRIBE_DONE_SUBSCRIPTION_ENDED SubscribeDoneStatusCode = 0x04
	SUBSCRIBE_DONE_GOING_AWAY         SubscribeDoneStatusCode = 0x05
	SUBSCRIBE_DONE_EXPIRED            SubscribeDoneStatusCode = 0x06
)
