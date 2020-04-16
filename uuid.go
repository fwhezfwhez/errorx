package errorx

// To avoid version conflict of uuid,
// uuid.go is copied from "github.com/satori/go.uuid"@b2ce2384e17bbe0c6d34077efa39dbab3e09123b

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Size of a UUID in bytes.
const Size = 16

// UUID representation compliant with specification
// described in RFC 4122.
type UUID [Size]byte

func (u *UUID) SetVersion(v byte) {
	u[6] = (u[6] & 0x0f) | (v << 4)
}

// SetVariant sets variant bits.
func (u *UUID) SetVariant(v byte) {
	switch v {
	case 0:
		u[8] = (u[8]&(0xff>>1) | (0x00 << 7))
	case 1:
		u[8] = (u[8]&(0xff>>2) | (0x02 << 6))
	case 2:
		u[8] = (u[8]&(0xff>>3) | (0x06 << 5))
	case 3:
		fallthrough
	default:
		u[8] = (u[8]&(0xff>>3) | (0x07 << 5))
	}
}

// Returns canonical string representation of UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}

type epochFunc func() time.Time
type hwAddrFunc func() (net.HardwareAddr, error)

var global = newRFC4122Generator()

// Generator provides interface for generating UUIDs.
type Generator interface {
	NewV4() (UUID, error)
}

func NewV4() (UUID, error) {
	return global.NewV4()
}

// Default generator implementation.
type rfc4122Generator struct {
	clockSequenceOnce sync.Once
	hardwareAddrOnce  sync.Once
	storageMutex      sync.Mutex

	rand io.Reader

	epochFunc     epochFunc
	hwAddrFunc    hwAddrFunc
	lastTime      uint64
	clockSequence uint16
	hardwareAddr  [6]byte
}

// NewV4 returns random generated UUID.
func (g *rfc4122Generator) NewV4() (UUID, error) {
	u := UUID{}
	if _, err := io.ReadFull(g.rand, u[:]); err != nil {
		return UUID{}, err
	}
	const V4 byte = 4
	const VariantRFC4122 byte = 1
	u.SetVersion(V4)
	u.SetVariant(VariantRFC4122)

	return u, nil
}

func newRFC4122Generator() Generator {
	return &rfc4122Generator{
		epochFunc:  time.Now,
		hwAddrFunc: defaultHWAddrFunc,
		rand:       rand.Reader,
	}
}

// Returns hardware address.
func defaultHWAddrFunc() (net.HardwareAddr, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return []byte{}, err
	}
	for _, iface := range ifaces {
		if len(iface.HardwareAddr) >= 6 {
			return iface.HardwareAddr, nil
		}
	}
	return []byte{}, fmt.Errorf("uuid: no HW address found")
}
