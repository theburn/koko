package zmodem

const (
	ZCRCE = 0x68 // 'h', 104, frame ends, header packet follows
	ZCRCG = 0x69 // 'i', 105, frame continues nonstop
	ZCRCQ = 0x6a // 'j', 106, frame continues, ZACK expected
	ZCRCW = 0x6b // 'k', 107, frame ends, ZACK expected
)

