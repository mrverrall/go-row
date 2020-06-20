# go-row-cycle
A Bluetooth LE bridge, written in Go to convert data from a Concept2 PM5 rowing computer into a Bluetooth Cycle Power Service for use in common cycle gaming platforms. go-row-cycle transmits power and cadance (spm).

## Good to know...
* Runs on Debian, Ubuntu, Raspbian... etc.
* Runs well on a Raspberry Pi Zero W

## But specifically..
* Requires Bluetooth 4.1+ chipsets
* Requires kernel support for HCI_CHANNEL_USER (v3.14+)

# TL;DR
    apt-get install golang
    go get "github.com/mrverrall/go-row-cycle"
    go build ~/go/src/github.com/mrverrall/go-row-cycle/go-row-cycle.go
    sudo ./go-row-cycle

# Obtaining and Building
To Compile go-row-cycle.go into an executable the Go compliler is required:

    sudo apt-get install golang

The easiest way to get go-row-cycle along with all it's dependancies is using 'go get':

    go get "github.com/mrverrall/go-row-cycle"

Then build:

    go build ~/go/src/github.com/mrverrall/go-row-cycle/go-row-cycle.go

# First Run
Ensure your BT device meets minimum version (4.1) Raspberry Pis with built in Bluetooth chipsets are fine.

    sudo hciconfig hci0 up
    hciconfig -a

The returned output includes your HCI version, similar to:

    hci0:   Type: Primary  Bus: UART
            BD Address: B8:27:EB:31:49:4A  ACL MTU: 1021:8  SCO MTU: 64:1
            UP RUNNING
            RX bytes:1308 acl:0 sco:0 events:66 errors:0
            TX bytes:838 acl:0 sco:0 commands:66 errors:0
            Features: 0xbf 0xfe 0xcf 0xfe 0xdb 0xff 0x7b 0x87
            Packet type: DM1 DM3 DM5 DH1 DH3 DH5 HV1 HV2 HV3
            Link policy: RSWITCH SNIFF
            Link mode: SLAVE ACCEPT
            Name: 'BCM43438A1 37.4MHz Raspberry Pi 3-0062'
            Class: 0x000000
            Service Classes: Unspecified
            Device Class: Miscellaneous,
            HCI Version: 4.1 (0x7)  Revision: 0x168
            LMP Version: 4.1 (0x7)  Subversion: 0x2209
            Manufacturer: Broadcom Corporation (15)

go-row-cycle uses HCI sockets exclusively. If you have Bluez installed then the bluetooth service should be disabled.

Before starting go-row-cycle ensure BLE device is down:

    sudo hciconfig hci0 down

If you have BlueZ installed stop the built-in bluetooth server, which may interfere:

    sudo service bluetooth stop

If you plan to use go-row-cycle on boot disable the bluetooth service entirly:

    sudo service bluetooth mask

go-row-cycle requires control of your bluetooth hardware, either via running with sudo or assigning the appropriate capabilities to the go-row-cycle executable.

    sudo setcap 'cap_net_raw,cap_net_admin=eip' ./go-row-cycle

Run!

    ./go-row-cycle

# Connecting to your Rower and Game
When a PM5 rower is not connected go-row-cycle is continually scanning for one. While go-row-cyle is running simply turn on wireless on the PM5 in the usual manner, connection will occure within 5 seconds.

While running, go-row-cycle will always advertise a Cycle Power Service over Bluetooth LE. Within your Bluetooth enabled cycle game select the 'GoRowCycle' device for power and cadance.

Row!
