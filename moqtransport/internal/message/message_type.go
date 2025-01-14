package message

type MessageType uint64

const (
	subscribe_update MessageType = 0x02
	subscribe        MessageType = 0x03
	subscribe_ok     MessageType = 0x04
	subscribe_error  MessageType = 0x05

	announce        MessageType = 0x06
	announce_ok     MessageType = 0x07
	announce_error  MessageType = 0x08
	unannounce      MessageType = 0x09
	unsubscribe     MessageType = 0x0A
	subscribe_done  MessageType = 0x0B
	announce_cancel MessageType = 0x0C

	track_status_request MessageType = 0x0D
	track_status         MessageType = 0x0E

	go_away MessageType = 0x10

	subscribe_announce       MessageType = 0x11
	subscribe_announce_ok    MessageType = 0x12
	subscribe_announce_error MessageType = 0x13
	unsubscribe_announce     MessageType = 0x14

	max_subscribe_id MessageType = 0x15

	fetch        MessageType = 0x16
	fetch_cancel MessageType = 0x17
	fetch_ok     MessageType = 0x18
	fetch_error  MessageType = 0x19

	subscribe_blocked MessageType = 0x1A

	client_setup MessageType = 0x40
	server_setup MessageType = 0x41
)
