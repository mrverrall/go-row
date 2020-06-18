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

type pm5Client struct {
	Name           string
	DataCh         chan []byte
	client         ble.Client
	service        *ble.Service
	characteristic *ble.Characteristic
}

func NewClient() (*pm5Client, error) {

	pm5 := &pm5Client{}
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 60*time.Second))

	log.Printf("Searching for sensors...\n")
	cln, err := ble.Connect(ctx, pm5.filter)
	if err != nil {
		return nil, err
	}

	pm5.client = cln

	// listen for disconnections and cleanup if so

	log.Printf("%s: Discovering profiles\n", pm5.Name)

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
			}
		}
	}
	return pm5, nil
}

func (pm5 *pm5Client) Subscribe() error {

	pm5.DataCh = make(chan []byte)

	go func() {
		<-pm5.client.Disconnected()
		log.Printf("[ %s ] is disconnected \n", pm5.client.Addr())
		close(pm5.DataCh)
	}()

	log.Printf("%s: Subscribe to notification", pm5.Name)
	return pm5.client.Subscribe(pm5.characteristic, false, pm5.notifyHandler)
}

func (pm5 *pm5Client) filter(a ble.Advertisement) bool {
	fmt.Printf("device found: %v\n", a.LocalName())
	if strings.HasPrefix(a.LocalName(), "PM5") {
		fmt.Printf("PM5 connectable: %v\n", a.Connectable())

		return a.Connectable()
	}
	return false
}

func (pm5 *pm5Client) notifyHandler(req []byte) {

	// we want live data so no buffering
	select {
	case pm5.DataCh <- req:
		// nothing to do
	default:
		// fine to skip some data if there is nowhere to send itc
	}
}
