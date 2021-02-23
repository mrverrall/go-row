package sensor

import (
	"fmt"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/mrverrall/go-row/pm5"
)

// Sensor a ble sensor fit for both CPM and RSC and more...
type Sensor struct {
	Name           string
	DataCh         chan pm5.Status
	timeout        time.Duration
	defaultPayload []byte
	trasform
	characteristics
	ble.UUID
	*ble.Service
}

// Characteristics array which will be added to the service when Iniialised
type characteristics []characteristic

// Characteristic properties
type characteristic struct {
	name     string
	function string
	payload  []byte
	uuid     ble.UUID
}

type trasform func(pm5.Status, []byte) []byte

// Initalise the ble service and advertise

func (s *Sensor) initalise() (*Sensor, error) {
	s.Service = ble.NewService(s.UUID)
	s.timeout = time.Second * 4

	for _, c := range s.characteristics {
		switch c.function {
		case "Notify":
			s.defaultPayload = c.payload
			s.AddCharacteristic(s.newNotifyChar(c.uuid))
		case "Read":
			s.AddCharacteristic(s.newReadChr(c.uuid, c.payload))
		default:
			err := fmt.Errorf("Unknown Characteristic Function: %v", c.function)
			return nil, err
		}
	}
	ble.AddService(s.Service)

	// settle
	time.Sleep(time.Second)
	return s, nil
}

func (s Sensor) newNotifyChar(u ble.UUID) *ble.Characteristic {
	c := ble.NewCharacteristic(u)
	c.HandleNotify(ble.NotifyHandlerFunc(s.notifyHandler))

	return c
}

func (s Sensor) newReadChr(u ble.UUID, d []byte) *ble.Characteristic {
	c := ble.NewCharacteristic(u)
	c.HandleRead(ble.ReadHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		rsp.Write(d)
	}))

	return c
}

func (s *Sensor) notifyHandler(req ble.Request, n ble.Notifier) {
	log.Println("Client Subscribed for Notifications...")

	for {
		select {
		case <-n.Context().Done():
			log.Println("Client un-subscribed for Notifications.")
			return
		case pm5status := <-s.DataCh:
			dpl := s.defaultPayload
			out := s.trasform(pm5status, dpl)
			_, err := n.Write(out)
			if err != nil {
				log.Printf("Client missing for notification: %s", err)
				return
			}
		case <-time.After(s.timeout):
			log.Printf("Waiting for data, default packet sent: %v END", s.defaultPayload)
			_, err := n.Write(s.defaultPayload)

			if err != nil {
				log.Printf("Client missing for notification: %s", err)
				return
			}
		}
	}
}
