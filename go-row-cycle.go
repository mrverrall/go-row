package main

import (
	"encoding/binary"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/mrverrall/go-row-cycle/cpm"
	"github.com/mrverrall/go-row-cycle/pm5"
	"github.com/mrverrall/go-row-cycle/rsc"
)

var (
	deviceName    = "go-row"
	doubleCadence = flag.Bool("dc", false, "double spm for cadance")
)

func main() {

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	flag.Parse()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("signal received from os: %s", sig)
		unsetBT()
		done <- true
	}()
	go btWorker(done)
	<-done
}

func btWorker(done chan bool) {
	for {
		unsetBT()
		d, err := linux.NewDeviceWithName(deviceName)
		if err != nil {
			log.Printf("can't get  BT device: %s", err)
			done <- true
		}
		ble.SetDefaultDevice(d)

		log.Printf("searching for PM5...")
		rower, err := pm5.NewClient()

		if err != nil {
			log.Printf("PM5 error: %s", err)
			continue
		}

		rsc := rsc.NewServer(deviceName)
		cpm := cpm.NewServer(deviceName)

		runPayload := []byte{}
		cyclePayload := []byte{}
		transmit := false

		for data := range rower.DataCh {

			switch data[0] {
			case 50:
				runPayload = convertPM5toRSC(data)
			case 54:

				cyclePayload = convertPM5toCPM(data)
				transmit = true
			}

			if transmit {
				select {
				case cpm.DataCh <- cyclePayload:
				default:
				}
				select {
				case rsc.DataCh <- runPayload:
				default:
				}
				transmit = false
			}
		}
	}
}

func unsetBT() {
	ble.Stop()
	time.Sleep(time.Second * 5)
}

func convertPM5toRSC(d []byte) []byte {
	runPacket := []byte{0x0, 0x0, 0x0, 0x0}

	// speed
	pm5Speed := binary.LittleEndian.Uint16(d[4:6])
	binary.LittleEndian.PutUint16(runPacket[1:3], uint16(float32(pm5Speed)*0.256))

	// cadance
	pm5SPM := d[6] * 6
	copy(runPacket[3:], []byte{byte(pm5SPM)})

	return runPacket
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

	// Elapsed time
	copy(cyclePacket[6:], elapsedTime[0:3])

	// Power
	copy(cyclePacket[2:], d[3:5])

	// Stroke Count
	if *doubleCadence {
		spm := binary.LittleEndian.Uint16(d[7:9])
		binary.LittleEndian.PutUint16(cyclePacket[4:], spm*2)
	} else {
		copy(cyclePacket[4:], d[7:9])
	}
	return cyclePacket
}
