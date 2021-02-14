package rsc

import (
	"context"
	"log"
	"time"

	"github.com/go-ble/ble"
)

var (
	// Running Speed and Cadance (RSC) GATT Service
	rscSvcUUID = ble.UUID16(0x1814)
	// RSC features GATT Characteristic
	rscFeaturesCharUUID = ble.UUID16(0x2A54)
	// RSC Measurement GATT Characteristic
	rscMesurmentCharUUID = ble.UUID16(0x2A53)
)

// Service A Bluetooth LE RSC Service with a Data Channel for upstream notifications
type Service struct {
	Name    string
	DataCh  chan []byte
	service *ble.Service
}

// Feature Chracteristic compliant with org.bluetooth.characteristic.rsc_feature
type feature struct {
	instStrideLen   bool // Instantaneous Stride Length Measurement Supported
	totalDistance   bool // Total Distance Measurement Supported
	statusWalkRun   bool // Walking or Running Status Supported
	calibration     bool // Calibration Procedure Supported
	sensorLocations bool // Multiple Sensor Locations Supported
}

// payload returns a 8-bit []byte of supported features for transmission
func (f feature) payload() []byte {
	return []byte{}
}

// Measurment Characteristic compliant with org.bluetooth.characteristic.rsc_measurement
type measurment struct {

	// 16 bit
	// payload is probably a method that encodes these values as a []byte
	instSpeed     int // some sendible unit mm/s?
	instCadance   int
	instStrideLen int
	totalDistance int
}

// payload returns a 16-bit []byte of measurements for transmission
func (m measurment) payload() []byte {

	return []byte{}
}

// NewServer Create a new Bluetooth LE RSC Service and advertise indefinitly
func NewServer(deviceName string) *Service {

	rsc := &Service{}
	rsc.Name = deviceName

	rsc.service = ble.NewService(rscSvcUUID)
	rsc.service.AddCharacteristic(rsc.simpleReadChr(rscFeaturesCharUUID, []byte{0x0, 0x0}))

	rsc.service.AddCharacteristic(rsc.newRSCChar())
	ble.AddService(rsc.service)

	go ble.AdvertiseNameAndServices(context.Background(), deviceName, rscSvcUUID)

	return rsc
}

func (rsc *Service) newRSCChar() *ble.Characteristic {

	rsc.DataCh = make(chan []byte)
	c := ble.NewCharacteristic(rscMesurmentCharUUID)
	c.HandleNotify(ble.NotifyHandlerFunc(rsc.notifyHandler))

	return c
}

func (rsc Service) simpleReadChr(u ble.UUID, d []byte) *ble.Characteristic {

	c := ble.NewCharacteristic(u)
	c.HandleRead(ble.ReadHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		rsp.Write(d)
	}))

	return c
}

// TODO this can be much more generic i.e. used for all the notifiers
// implement payload as paramater
func (rsc *Service) notifyHandler(req ble.Request, n ble.Notifier) {

	log.Println("Client Subscribed for Notifications...")

	defaultPacket := []byte{0x0, 0x0, 0x0, 0x0}
	timeout := time.Second * 4

	for {
		select {
		case <-n.Context().Done():
			log.Println("Client un-subscribed for Notifications.")
			return
		case packet := <-rsc.DataCh:
			_, err := n.Write(packet)
			if err != nil {
				log.Printf("Client missing for notification: %s", err)
				return
			}
		case <-time.After(timeout):
			_, err := n.Write(defaultPacket)
			if err != nil {
				log.Printf("Client missing for notification: %s", err)
				return
			}
		}
	}
}
