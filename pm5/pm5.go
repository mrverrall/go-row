package pm5

import (
	"context"
	"fmt"
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
	name           string
	DataCh         chan []byte
	client         ble.Client
	characteristic *ble.Characteristic
}

// NewClient Searches for and connects to the first PM5 it can
func NewClient() (*Client, error) {
	pm5 := &Client{}

	for pm5.client == nil {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 10*time.Second))
		cln, err := ble.Connect(ctx, pm5.filter)
		if err != nil {
			if err != context.DeadlineExceeded {
				return nil, fmt.Errorf("Unresolvable connection error: %s", err)
			}
			time.Sleep(time.Second * 10)
		}
		pm5.client = cln
	}

	log.Printf("%s: scanning for services", pm5.name)
	p, err := pm5.client.DiscoverProfile(false)
	if err != nil {
		ble.Client.CancelConnection(pm5.client)
		return nil, fmt.Errorf("%s: can't discover Service", err)
	}

	for _, s := range p.Services {
		for _, c := range s.Characteristics {
			if c.UUID.String() == pm5ChrUUID.String() {
				pm5.characteristic = c
				if err := pm5.subscribe(); err != nil {
					return nil, fmt.Errorf("failed to subscribe to PM5 notifications: %s", err)
				}
			}
		}
	}
	if pm5.characteristic == nil {
		return nil, fmt.Errorf("%s: failed to identify compatible service", pm5.name)
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

	log.Printf("%s: subscribe to notification", pm5.name)
	return pm5.client.Subscribe(pm5.characteristic, false, pm5.notifyHandler)
}

func (pm5 *Client) filter(a ble.Advertisement) bool {

	if strings.HasPrefix(a.LocalName(), "PM5") {
		log.Printf("device found: %s\n", a.LocalName())
		pm5.name = a.LocalName()
		return a.Connectable()
	}
	return false
}

func (pm5 *Client) notifyHandler(req []byte) {
	select {
	case pm5.DataCh <- req:
	default:
	}
}
