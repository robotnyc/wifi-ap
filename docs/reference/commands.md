---
title: "Available Commands"
table_of_contents: True
---

# Available Commands

The wifi-ap snap offers a few command line utility programs to configure and
control an access point.

## wifi-ap.config

The *wifi-ap.config* command allows to change one or multiple of the various available
configuration options of the access point.

Example:

```
$ wifi-ap.config set disabled=false
$ wifi-ap.config get
debug: false
dhcp.lease-time: 12h
dhcp.range-start: 10.0.60.2
dhcp.range-stop: 10.0.60.199
disabled: true
share.disabled: false
share.network-interface: wlan0
wifi.address: 10.0.60.1
wifi.channel: 6
wifi.hostapd-driver: nl80211
wifi.interface: wlan0
wifi.interface-mode: direct
wifi.netmask: 255.255.255.0
wifi.operation-mode: g
wifi.security: open
wifi.security-passphrase:
wifi.ssid: Ubuntu
```

## wifi-ap.status

The *wifi-ap.status* command allows to display the current status of the operated
access point or perform different actions.

```
$ wifi-ap.status
ap.active: false
$ wifi-ap.status restart-ap
```
