package main

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/mrverrall/go-row-cycle/cpm"
	"github.com/mrverrall/go-row-cycle/pm5"
)

func main() {
	for {
		setBT()

		cpm := cpm.NewServer()

		pm5, err := pm5.NewClient()
		if err != nil {
			log.Printf("Failed to get PM5 client")
			unsetBT()
			continue
		}

		if err := pm5.Subscribe(); err != nil {
			log.Printf("Failed to subscribe to PM5 Notifications")
			unsetBT()
			continue
		}

		for data := range pm5.DataCh {

			cycleData := convertPM5toCPM(data)
			select {
			case cpm.DataCh <- cycleData:
			default:
			}
		}

		unsetBT()
	}
}

func setBT() {
	d, err := linux.NewDevice()
	if err != nil {
		log.Fatalf("Can't get  BT device : %s", err)
	}
	ble.SetDefaultDevice(d)
}

func unsetBT() {
	ble.Stop()
	time.Sleep(time.Second * 5)
}

func convertPM5toCPM(d []byte) []byte {
	cyclePacket := []byte{0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

	// Bluetooth power meters count time in 1/1024th of a second while C2 reports in 100ths
	// map c2 time to uint32, covert to 1/1024s the take the 2 least significant bytes
	elapsedTime := make([]byte, 4)
	copy(elapsedTime[0:], d[:4])

	elapsedTime32 := binary.LittleEndian.Uint32(elapsedTime)
	elapsedTime32 = (elapsedTime32 * 1024) / 100
	binary.LittleEndian.PutUint32(elapsedTime, elapsedTime32)

	copy(cyclePacket[2:], d[3:5])
	copy(cyclePacket[4:], d[7:9])
	copy(cyclePacket[6:], elapsedTime[0:3])
	return cyclePacket
}
