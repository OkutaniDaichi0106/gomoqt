package protocol

type Version uint64

const (
	Develop Version = 0xffffff00
	Draft01 Version = 0xff0bad01
	Draft02 Version = 0xff0bad02
	Draft03 Version = 0xff0bad03
)
