package pm5

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-ble/ble"
)

var (
	pm5SvcUUID = ble.MustParse("CE060030-43E5-11E4-916C-0800200C9A66")
	pm5ChrUUID = ble.MustParse("CE060036-43E5-11E4-916C-0800200C9A66")
)

// Client a Bluetooth LE Client connection to a PM5 rowing computer with a data channel for noticications receiced from it's service
type Client struct {
	clientName     string
	DataCh         chan []byte
	client         ble.Client
	service        *ble.Service
	characteristic *ble.Characteristic
}

// NewClient Searches for and connects to the first PM5 it can
func NewClient() (*Client, error) {

	pm5 := &Client{}
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 60*time.Second))

	log.Printf("Searching for sensors...\n")
	cln, err := ble.Connect(ctx, pm5.filter)
	if err != nil {
		return nil, err
	}

	pm5.client = cln

	log.Printf("%s: Discovering profiles\n", pm5.clientName)
	p, err := pm5.client.DiscoverProfile(false)
	if err != nil {
		log.Printf("Can't discover profiles: %s\n", err)
		return nil, err
	}

	for _, s := range p.Services {
		for _, c := range s.Characteristics {
			if c.UUID.String() == pm5ChrUUID.String() {
				pm5.service = s
				pm5.characteristic = c
				if err := pm5.subscribe(); err != nil {
					log.Printf("Failed to subscribe to PM5 Notifications")
					return nil, err
				}
			}
		}
	}
	return pm5, nil
}

func (pm5 *Client) subscribe() error {

	pm5.DataCh = make(chan []byte)
	go func() {
		<-pm5.client.Disconnected()
		log.Printf("%s: disconnected \n", pm5.client.Addr())
		close(pm5.DataCh)
	}()

	log.Printf("%s: Subscribe to notification", pm5.clientName)
	return pm5.client.Subscribe(pm5.characteristic, false, pm5.notifyHandler)
}

func (pm5 *Client) filter(a ble.Advertisement) bool {

	if strings.HasPrefix(a.LocalName(), "PM5") {
		log.Printf("Device found: %s\n", a.LocalName())
		return a.Connectable()
	}
	return false
}

func (pm5 *Client) notifyHandler(req []byte) {
	// only realtime live data is valid so no buffering
	select {
	case pm5.DataCh <- req:
		// data sent, nothing else to do
	default:
		// fine to skip some data if there is nowhere to send it
	}
}
