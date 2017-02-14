---
title: "Simultaneous STA / AP Mode"
table_of_contents: True
---

# Simultaneous STA / AP Mode

If your hardware and the kernel driver supports a simultaneous STA / AP mode you can stay connected to another access point while running your own. The only shortcoming here is that both have to operate on the same channel.

You can find the current channel being used for the STA connection by looking at the STA network interface with the iw command.

To get the iw command available on an Ubuntu Core system install the wireless-tools snap from the store

```
$ snap install wireless-tools
```

Connect required slots/plugs

```
$ snap connect wireless-tools:network-control core:network-control
```

Now you can run the iw command to find the current channel being used

```
$ wireless-tools.iw dev
phy#0
	Unnamed/non-netdev interface
		wdev 0x4
		addr aa:bb:cc:dd:ee:ff
		type P2P-device
	Interface wlan0
		ifindex 3
		wdev 0x1
		addr aa:bb:cc:dd:ee:ff
		type managed
		channel 1 (2412 MHz), width: 20 MHz, center1: 2412 MHz
```

**NOTE:** If the channel is not part of the output your kernel WiFi driver doesn’t report the channel.

The relevant line showing the channel being used is highlighted above. In this case it is channel 1. The next step is to configure the wifi-ap snap with the channel

```
$ wifi-ap.config set wifi.channel=1
```
