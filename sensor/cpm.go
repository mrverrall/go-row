package sensor

import (
	"encoding/binary"

	"github.com/go-ble/ble"
	"github.com/mrverrall/go-row/pm5"
)

// NewCyclePower Create a new Cycle Power Service
func NewCyclePower(deviceName string) *Sensor {
	cpm := &Sensor{
		Name:   "go-row",
		UUID:   ble.UUID16(0x1818),
		DataCh: make(chan pm5.Status, 1),
		characteristics: []characteristic{
			{
				name:     "Measurment",
				function: "Notify",
				uuid:     ble.UUID16(0x2a63),
				payload:  []byte{0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
			{
				name:     "Features",
				function: "Read",
				uuid:     ble.UUID16(0x2a65),
				payload:  []byte{0x8, 0x0, 0x0, 0x0},
			},
			{
				name:     "Sensor Location",
				function: "Read",
				uuid:     ble.UUID16(0x2a5d),
				payload:  []byte{0x0},
			},
		},
		trasform: func(in pm5.Status, out []byte) []byte {
			// Elapsed Time 1/1024 s
			binary.LittleEndian.PutUint16(out[6:], uint16(in.LastStroke.Seconds()*1024))
			// Power
			binary.LittleEndian.PutUint16(out[2:], in.Power)
			// Revolution Count
			binary.LittleEndian.PutUint16(out[4:], uint16(in.StrokeCount*3))

			return out
		},
	}

	cpm.initalise()
	return cpm
}
