package moqtransport

type Role byte

const (
	PUB     Role = 0x00
	SUB     Role = 0x01
	PUB_SUB Role = 0x02
)

type Endpoint byte

const (
	CLIENT Endpoint = 0x00
	SERVER Endpoint = 0x01
)
