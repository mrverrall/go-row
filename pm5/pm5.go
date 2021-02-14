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
	svcUUID = ble.MustParse("CE060030-43E5-11E4-916C-0800200C9A66")
	// multiChr C2 multiplexed information characteristic
	multiChr = ble.MustParse("CE060080-43E5-11E4-916C-0800200C9A66")
)

// Client a Bluetooth LE Client connection to a PM5 rowing computer with a data channel for noticications receiced from it's service
type Client struct {
	name   string
	DataCh chan []byte
	client ble.Client
	//characteristics []*ble.Characteristic
}

type characteristic struct {
	UUID    ble.UUID
	Handler ble.NotificationHandler
}

// NewClient Searches for and connects to the first PM5 it can
func NewClient() (*Client, error) {
	pm5 := &Client{}
	pm5.DataCh = make(chan []byte)

	for pm5.client == nil {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 10*time.Second))
		cln, err := ble.Connect(ctx, pm5.filter)
		if err != nil {
			if err.Error() != "can't scan: context deadline exceeded" {
				return nil, fmt.Errorf("Unresolvable connection error: %s", err)
			}
			time.Sleep(time.Second * 10)
		}
		pm5.client = cln
	}
	err := pm5.SubscribeToCharacteristics([]ble.UUID{multiChr})
	if err != nil {
		return nil, fmt.Errorf("Subscription error: %s", err)
	}
	return pm5, nil
}

// SubscribeToCharacteristics Subscribe to notifications for charcteristics
func (pm5 *Client) SubscribeToCharacteristics(chrs []ble.UUID) error {

	log.Printf("%s: scanning for services", pm5.name)
	p, err := pm5.client.DiscoverProfile(false)
	if err != nil {
		ble.Client.CancelConnection(pm5.client)
		return fmt.Errorf("%s: can't discover Service", err)
	}

	for _, s := range p.Services {
		for _, c := range s.Characteristics {
			for _, chr := range chrs {
				if c.UUID.String() == chr.String() {
					if err := pm5.subscribe(c); err != nil {
						return fmt.Errorf("failed to subscribe: %s", err)
					}
				}
			}
		}
	}
	return nil
}

func (pm5 *Client) subscribe(c *ble.Characteristic) error {

	go func() {
		<-pm5.client.Disconnected()
		log.Printf("%s: disconnected \n", pm5.client.Addr())
		close(pm5.DataCh)
	}()

	log.Printf("%s: subscribing to notifications for %s", pm5.name, c.UUID.String())
	return pm5.client.Subscribe(c, false, pm5.notifyHandler)
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
