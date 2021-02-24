package peripheral

import (
	"encoding/binary"

	"github.com/go-ble/ble"
	"github.com/mrverrall/go-row/pm5"
)

// NewRunningSpeed Create a new Running Speed and Cadance Service
func NewRunningSpeed(deviceName string) *Sensor {
	rsc := &Sensor{
		Name:   "go-row",
		UUID:   ble.UUID16(0x1814),
		DataCh: make(chan pm5.Status, 1),
		characteristics: []characteristic{
			{
				name:     "Measurment",
				function: "Notify",
				uuid:     ble.UUID16(0x2A53),
				payload:  []byte{0x0, 0x0, 0x0, 0x0},
			},
			{
				name:     "Features",
				function: "Read",
				uuid:     ble.UUID16(0x2A54),
				payload:  []byte{0x0, 0x0},
			},
		},
		trasform: func(in pm5.Status, out []byte) []byte {
			// Speed
			binary.LittleEndian.PutUint16(out[1:3], uint16(float32(in.Speed)*0.256))
			// Cadance 6x multiplier for more realistic running
			copy(out[3:], []byte{byte(in.Spm * 6)})
			return out
		},
	}

	rsc.initalise()
	return rsc
}
