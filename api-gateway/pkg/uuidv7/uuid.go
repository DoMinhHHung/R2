package uuidv7

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

var counter atomic.Uint32

func New() string {
	ms := uint64(time.Now().UnixMilli())
	var b [16]byte

	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = 0x70 | byte(ms>>4)&0x0f

	seq := counter.Add(1) & 0x3fff
	b[7] = byte(seq >> 6)
	b[8] = 0x80 | byte(seq&0x3f)
	b[9] = byte(seq)

	rand.Read(b[10:])

	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:]),
	)
}
