---
title: "Installation"
table_of_contents: True
---

# Installation

The snap can be installed from the Ubuntu Store with the following command

```
 $ snap install wifi-ap
```

If you’re running at least snapd version 2.17 all plugs of the wifi-ap snap will be automatically connected to the right slot. If you’re running an older snapd version or a locally build snap you need to connect the plugs manually.

You can verify with the following command that the relevant plugs are connected:

```
$ snap interfaces
Slot                       Plug
[...]
:network-bind              wifi-ap
[...]
:firewall-control          wifi-ap:firewall-control
:network-control           wifi-ap:network-control
:network-manager           wifi-ap:network-manager
```

If you also have the network-manager snap installed on the system then you need
to connect the network-manager plug of the wifi-ap snap too. If you have the
network-manager installed before you install the wifi-ap snap the plug gets
automatically connected. Otherwise you have to do that manually:

```
$ snap connect wifi-ap:network-manager network-manager:service
```

# Default Configuration

| Name | Value |
|------|-------|
| **WiFi SSID** | Ubuntu |
| **WiFi Interface mode** | direct |
| **WiFi Security** | wpa2 |
| **WiFi Security Passphrase ** | Randomly chosen |
| **WiFi Channel** | 6 |
| **WiFi Network Interface** | wlan0 |
| **WiFi Address** | 192.168.7.1 |
| **WiFi Netmask** | 255.255.255.0 |
| **DHCP Range** | 192.168.7.5 - 192.168.7.100 |
| **DHCP Lease Time** | 12h |
