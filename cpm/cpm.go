package cpm

import (
	"context"
	"fmt"
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

type cpService struct {
	Name    string
	DataCh  chan []byte
	service *ble.Service
}

func NewServer() *cpService {

	cps := &cpService{}

	cps.service = ble.NewService(cpSvcUUID)
	cps.service.AddCharacteristic(cps.simpleReadChr(cpfCharUUID, []byte{0x8, 0x0, 0x0, 0x0}))
	cps.service.AddCharacteristic(cps.simpleReadChr(cslCharUUID, []byte{0x0}))

	cps.service.AddCharacteristic(cps.newCpmChar())
	ble.AddService(cps.service)

	go ble.AdvertiseNameAndServices(context.Background(), "GoRowCycle", cpSvcUUID)

	return cps
}

func (cps *cpService) newCpmChar() *ble.Characteristic {

	cps.DataCh = make(chan []byte)
	c := ble.NewCharacteristic(cpmChrUUID)
	c.HandleNotify(ble.NotifyHandlerFunc(cps.notifyHandler))

	return c
}

func (cps cpService) simpleReadChr(u ble.UUID, d []byte) *ble.Characteristic {

	c := ble.NewCharacteristic(u)
	c.HandleRead(ble.ReadHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		rsp.Write(d)
	}))

	return c
}

func (cps *cpService) notifyHandler(req ble.Request, n ble.Notifier) {

	fmt.Println("Client Subscribed for Notifications...")

	defaultPacket := []byte{0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	timeout := time.Second * 5

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
			log.Printf("Downstream timeout sending default packet: %s", timeout)
			_, err := n.Write(defaultPacket)
			if err != nil {
				log.Printf("Client missing for notification: %s", err)
				return
			}
		}
	}
}
