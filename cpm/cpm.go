package cpm

import (
	"context"
	"log"
	"time"

	"github.com/go-ble/ble"
)

var (
	// Cycle Power Service and Characteristics
	cpSvcUUID   = ble.UUID16(0x1818)
	cpmChrUUID  = ble.UUID16(0x2a63)
	cpfCharUUID = ble.UUID16(0x2a65)
	cslCharUUID = ble.UUID16(0x2a5d)
)

// Service A Bluetooth LE Cycle Power Service with a Data Channel for upstream notifications (power and cadence)
type Service struct {
	Name    string
	DataCh  chan []byte
	service *ble.Service
}

// NewServer Create a new Bluetooth LE Cycle Power Service and advertise indefinitly
func NewServer(deviceName string) *Service {

	cps := &Service{}
	cps.Name = deviceName

	cps.service = ble.NewService(cpSvcUUID)
	cps.service.AddCharacteristic(cps.simpleReadChr(cpfCharUUID, []byte{0x8, 0x0, 0x0, 0x0}))
	cps.service.AddCharacteristic(cps.simpleReadChr(cslCharUUID, []byte{0x0}))

	cps.service.AddCharacteristic(cps.newCpmChar())
	ble.AddService(cps.service)

	go ble.AdvertiseNameAndServices(context.Background(), deviceName, cpSvcUUID)

	return cps
}

func (cps *Service) newCpmChar() *ble.Characteristic {

	cps.DataCh = make(chan []byte)
	c := ble.NewCharacteristic(cpmChrUUID)
	c.HandleNotify(ble.NotifyHandlerFunc(cps.notifyHandler))

	return c
}

func (cps Service) simpleReadChr(u ble.UUID, d []byte) *ble.Characteristic {

	c := ble.NewCharacteristic(u)
	c.HandleRead(ble.ReadHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		rsp.Write(d)
	}))

	return c
}

func (cps *Service) notifyHandler(req ble.Request, n ble.Notifier) {

	log.Println("Client Subscribed for Notifications...")

	defaultPacket := []byte{0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	timeout := time.Second * 4

	for {
		select {
		case <-n.Context().Done():
			log.Println("Client un-subscribed for Notifications.")
			return
		case packet := <-cps.DataCh:
			_, err := n.Write(packet)
			if err != nil {
				log.Printf("Client missing for notification: %s", err)
				return
			}
		case <-time.After(timeout):
			log.Printf("Downstream timeout (%s), sending default packet.", timeout)
			_, err := n.Write(defaultPacket)
			if err != nil {
				log.Printf("Client missing for notification: %s", err)
				return
			}
		}
	}
}
