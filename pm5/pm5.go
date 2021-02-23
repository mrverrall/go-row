package pm5

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-ble/ble"
)

// Client a Bluetooth LE Client connection to a PM5 rowing computer with a data channel for noticications receiced from it's service
type Client struct {
	name     string
	StatusCh chan Status
	mutex    sync.Mutex
	ble.Client
	Status
}

// Status of a Client, e.g. stroke count, current speed etc.
type Status struct {
	StrokeCount uint16        // Total strokes for session
	LastStroke  time.Duration // Time marker for last stroke event
	Power       uint16        // Power in watts
	Speed       uint16        // Speed in 0.001m/s
	RowState    byte          // Are we still rowing?
	Spm         byte          // Strokes per miniute
	Heartrate   byte          // Heartrate

}

type subscription struct {
	Name     string
	uuid     ble.UUID
	notifier ble.NotificationHandler
}

// NewClient Searches for and connects to the first PM5 it can
func NewClient() (*Client, error) {
	var err error

	pm5 := &Client{
		StatusCh: make(chan Status, 1),
		mutex:    sync.Mutex{},
	}

	err = pm5.setBleClient()
	if err != nil {
		return nil, err
	}

	err = pm5.subscribeToCharacteristics()
	if err != nil {
		return nil, err
	}

	// watch for disconnections
	go func() {
		<-pm5.Client.Disconnected()
		log.Printf("%s: disconnected \n", pm5.Client.Addr())
		close(pm5.StatusCh)
	}()

	return pm5, nil
}

func (pm5 *Client) setBleClient() error {
	for pm5.Client == nil {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 10*time.Second))
		cln, err := ble.Connect(ctx, pm5.filterAdvert)
		if err != nil {
			if err.Error() != "can't scan: context deadline exceeded" {
				return fmt.Errorf("Unresolvable connection error: %s", err)
			}
			time.Sleep(time.Second * 10)
		}
		pm5.Client = cln
	}
	return nil
}

func (pm5 *Client) subscriptions() []subscription {
	return []subscription{
		{
			Name:     "C2 rowing general status characteristic",
			uuid:     ble.MustParse("CE060031-43E5-11E4-916C-0800200C9A66"),
			notifier: pm5.notifyHandler31,
		},
		{
			Name:     "C2 rowing additional status 1 characteristic",
			uuid:     ble.MustParse("CE060032-43E5-11E4-916C-0800200C9A66"),
			notifier: pm5.notifyHandler32,
		},
		{
			Name:     "C2 rowing additional stroke data characteristic",
			uuid:     ble.MustParse("CE060036-43E5-11E4-916C-0800200C9A66"),
			notifier: pm5.notifyHandler36,
		},
	}
}

func (pm5 *Client) subscribeToCharacteristics() error {
	log.Printf("%s: scanning for services", pm5.name)
	p, err := pm5.Client.DiscoverProfile(false)
	if err != nil {
		ble.Client.CancelConnection(pm5.Client)
		return fmt.Errorf("%s: can't discover Service", err)
	}

	for _, s := range p.Services {
		for _, c := range s.Characteristics {
			for _, sub := range pm5.subscriptions() {
				if c.UUID.Equal(sub.uuid) {
					if err := pm5.subscribe(c, sub); err != nil {
						return fmt.Errorf("failed to subscribe: %s", err)
					}
				}
			}
		}
	}
	return nil
}

func (pm5 *Client) subscribe(c *ble.Characteristic, s subscription) error {
	log.Printf("%s: subscribing to notifications for %s", pm5.name, c.UUID.String())
	return pm5.Client.Subscribe(c, false, s.notifier)
}

func (pm5 *Client) filterAdvert(a ble.Advertisement) bool {
	if strings.HasPrefix(a.LocalName(), "PM5") {
		log.Printf("device found: %s\n", a.LocalName())
		return a.Connectable()
	}
	return false
}

func (pm5 *Client) notifyHandler31(data []byte) {
	pm5.RowState = data[9]
	pm5.notify()
}

func (pm5 *Client) notifyHandler32(data []byte) {
	pm5.Speed = binary.LittleEndian.Uint16(data[3:5])
	pm5.Heartrate = data[6]
	pm5.Spm = data[5]
}

func (pm5 *Client) notifyHandler36(data []byte) {
	pm5.LastStroke = pm5Time2Duration(data)
	pm5.Power = binary.LittleEndian.Uint16(data[3:5])
	pm5.StrokeCount = binary.LittleEndian.Uint16(data[7:9])
}

func (pm5 *Client) notify() {
	if pm5.RowState != 1 {
		pm5.Power = 0
		pm5.Speed = 0
		pm5.Spm = 0
	}
	select {
	case pm5.StatusCh <- pm5.Status:
	default:
	}
}

func pm5Time2Duration(d []byte) time.Duration {
	// 24-bit PM5 value for elapsed time to 32-bit uint
	et := make([]byte, 4)
	copy(et[0:], d[0:3])
	return time.Duration(binary.LittleEndian.Uint32(et)) * time.Millisecond * 10
}
