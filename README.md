# go-row
Go-row is a Bluetooth LE bridge written in Go. It re-transmits a Concept2 PM5 rowers metrics as Bluetooth 'Cycle Power' and 'Running Speed' services. This allows you to use a Concept2 rower in cycling or runnning games such a Zwift.

__You can now install gow-row as a simple [deb package for Raspbian](https://github.com/mrverrall/go-row#raspberry-pi-package-deb).__

Cycling cadance is set to 3x the rowing SPM and running cadance 6x.

# Quick Start

## Raspberry Pi package (deb)
The go-row deb package can be download from the [releases page](https://github.com/mrverrall/go-row/releases/latest). Below is an example of installing via the command line. The package will install go-row as a service that starts on boot, perfect for a headless system.

    # Download
    wget https://github.com/mrverrall/go-row/releases/download/v0.0.1-2/go-row_0.0.1-2_armhf.deb
    
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

Once connected to a PM5 go-row will advertise the cycle and running services. Within your game/app select the 'go-row' device.

Row!

## Installing as a service

__N.B.__ This is not needed if you installed go-row using the debian package.

[An example systemd service file is included in this repository](https://github.com/mrverrall/go-row/blob/main/go-row.service).

To install as a boot service with systemd, edit the "ExecStart" path in the service file to the location your compiled go-row executable.

Copy your service file to '/var/lib/systemd/system/go-row.service'. then,

    sudo systemctl daemon-reload
    sudo systemctl enable go-row.service
    sudo systemctl start go-row.service

Check your service is running with,

    sudo systemctl status go-row.service
