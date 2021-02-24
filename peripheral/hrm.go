package peripheral

import (
	"github.com/go-ble/ble"
	"github.com/mrverrall/go-row/pm5"
)

// NewHRM Create a new Hear Rate Monitor
func NewHRM(deviceName string) *Sensor {
	rsc := &Sensor{
		Name:   "go-row",
		UUID:   ble.UUID16(0x180d),
		DataCh: make(chan pm5.Status, 1),
		characteristics: []characteristic{
			{
				name:     "Measurment",
				function: "Notify",
				uuid:     ble.UUID16(0x2A37),
				payload:  []byte{0x0, 0x0},
			},
		},
		trasform: func(in pm5.Status, out []byte) []byte {
			//BPM
			copy(out[1:], []byte{in.Heartrate}[:1])
			return out
		},
	}

	rsc.initalise()
	return rsc
}
