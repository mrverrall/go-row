# go-row
Go-row is a Bluetooth LE bridge, written in Go, to convert a Concept2 PM5 rower into a Bluetooth Cycle Power and Running Speed services.

__You can now install gow-row as a simple [deb package for Raspbian](https://github.com/mrverrall/go-row#raspberry-pi-package-deb).__

This lets you to use a Concept2 rower in cycling/runnning games such a Zwift.

Cycling cadance is set to 3x the rowing SPM and running cadance 6x.

## Good to know...
* Runs on Debian, Ubuntu, Raspbian... etc.
* Runs perfectly on a Raspberry Pi Zero W

## But specifically..
* As mentioned, works great on a Raspberry Pi Zero W
* Requires Bluetooth 4.1+ chipsets (allowing clients and servers to run simultaneously)

# Quick Start

## Raspberry Pi package (deb)
Packages can be download from the [releases page](https://github.com/mrverrall/go-row/releases). Install in your usual way, here is an example via the command line. This method also automatically installs go-row as a service on boot, perfect for a headless system.

    # Download
    wget https://github.com/mrverrall/go-row/releases/download/v0.0.1-0/go-row_0.0.1-2_armhf.deb
    
    # Install
    sudo dpkg -i go-row_0.0.1-2_armhf.deb

Go-row should now be running as a service, you can check this with,
    
    systemctl status go-row

## From Source
Assuming a clean install of Raspbian on an Raspberry Pi Zero W...

    apt-get install golang
    go get "github.com/mrverrall/go-row"
    sudo ~/go/bin/go-row

## Connecting to your Rower and Game
While go-row is running select 'connect' from the main PM5 menu, connection is then automatic.

Once connected to a PM5 go-row will advertise the cycle and running services over bluetooth. Within your game/app select the 'go-row' device.

Row!

## Installing as a service

__N.B.__ This is not needed if you installed go-row using the debian package.

It's easy to run go-row automatically on boot. This is ideal if you want a 'plug and play' setup without needing your device (like a Raspberry Pi Zero W) plugged into anything but power. In Raspbian this is acheived with systemd service.

[An example systemd service file is included in this repository](https://github.com/mrverrall/go-row/blob/main/go-row.service).

To install as a boot service with systemd, edit the "ExecStart" path in the service file to the location your compiled go-row executable.

Copy your service file to '/var/lib/systemd/system/go-row.service'. then,

    sudo systemctl daemon-reload
    sudo systemctl enable go-row.service
    sudo systemctl start go-row.service

Check your service is ruuning with,

    sudo systemctl status go-row.service
