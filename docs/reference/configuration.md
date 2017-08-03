---
title: "Configuration"
table_of_contents: True
---

# Configuration

The WiFi access point service offers a wide range of customization options which
can be modified through the built-in configuration system. This can be either done
by the *wifi-ap.config* command line utility or directly via the REST API.

The following describes all available configuration items.

**NOTE:** All configuration items described here are not available through the
snap configuration system yet. Both configuration approaches are currently
orthogonal to each other and are used for different purposes. In the future,
both approaches will most likely be merged in a backwards compatible manner.

## disabled

Marks the access point as disabled or not.

Possible values are

 * **false**: Access point is enabled
 * **true**: Access point is disabled

Default value: *true*

Example:

```
$ wifi-ap.config set disabled=true
```

## debug

Enables verbose debug logging.

Possible values are

 * **false**: Debug logging is disabled
 * **true**: Debug logging is enabled

Default value: *false*

Example:

```
$ wifi-ap.config set debug=true
```

## wifi.interface

The network interface being used to operate the access point on.

Default value: *wlan0*

Example:

```
$ wifi-ap.config set wifi.interface=wlan1
```

## wifi.address

IP address being used as for the gateway point on the access point network.

Default value: *192.168.7.1*

Example:

```
$ wifi-ap.config set wifi.address=192.168.8.1
```

## wifi.netmask

Netmask being used for the network the access point operates on.

Default value: *255.255.255.0*

Example:

```
$ wifi-ap.config set wifi.netmask=255.255.255.0
```

## wifi.interface-mode

Interface mode which describes how the access point is created.

Possible values:

 * **direct**: The network interface specified with wifi.interface is directly used to operate the access point. It will not be available for STA connections.
 * **virtual**: This mode is being used to allow simultaneous STA and AP mode on the same WiFi device. The backend service will create a virtual network interface named ap0 to operate the access point on. The actual WiFi network interface can be continued to use for STA connections by the system network connection manager.

Default value: *direct*

Example:

```
$ wifi-ap.config set wifi.interface-mode=virtual
```

## wifi.hostapd-driver

The hostapd driver being used.

Possible values:

 * *nl80211*: Let hostapd talk through the nl80211 interface with the kernel WiFi drivers.
 * *rtl8188*: A special hostapd version will be used which is specific for WiFi chips from Realtek.

Default value: *nl80211*

Example:

```
$ wifi-ap.config set wifi.hostapd-driver=rtl8188
```

## wifi.ssid

The WiFi SSID used for the access point, up to 32 characters. UTF-8 characters can be included in the SSID.

Default value: *Ubuntu*

Example:

```
$ wifi-ap.config set wifi.ssid=MyTestSSID
```

## wifi.security

WiFi security type used for the access point.

Possible values:

 * *open*: No authentication required and no encryption of the network traffic provided.
 * *wpa2*: Using WPA 2 Personal security. Requires a passphrase being configured.

Example:

```
$ wifi-ap.config set wifi.security=wpa2
```

## wifi.security-passphrase

WiFi security passphrase.

Default value: auto-generated secure password

Example:

```
$ wifi-ap.config set wifi.security-passphrase=Test1234
```

## wifi.channel

WiFi channel the access point will be operated on.

Default value: *6*

Example:

```
$ wifi-ap.config set wifi.channel=8
```

## wifi.operation-mode

WiFi operation mode for the access point

Possible values:

 * *a*: IEEE 802.11a (5 GHz)
 * *b*: IEEE 802.11b (2.4 GHz)
 * *g*: IEEE 802.11g (2.4 GHz)
 * *ad*: IEEE 802.11ad (60 GHz)

Default value: *g*

Example:

```
$ wifi-ap.config set wifi.operation-mode=g
```

## share.disabled

Disable network sharing. Possible values are:

 * *false*: Network sharing is enabled
 * *true*: Network sharing is disabled

Default value: *false*

Example:

```
$ wifi-ap.config set share.disabled=true
```

## share.network-interface

Network interface which network will be shared with clients connected to the access point.

Default value: *eth0*

Example:

```
$ wifi-ap.config set share.network-interface=eth1
```

## dhcp.range-start

Beginning of the IP address range being used to assign IP addresses to DHCP clients

Default value: *192.168.7.5*

Example:

```
$ wifi-ap.config set dhcp.range-start=192.168.7.10
```

## dhcp.range-stop

End of the IP address range being used to assign IP addresses to DHCP clients

Default value: *192.168.7.100*

Example:

```
$ wifi-ap.config set dhcp.range-stop=192.168.7.200
```

## dhcp.lease-time

Lease time given to IP address assigned to DHCP client. The lease time is in seconds, or minutes (eg 45m) or hours (eg 1h) or "infinite".

Default value: *12h*

Example:

```
$ wifi-ap.config set dhcp.lease-time=24h
```

## wifi.country-code

Country code as specified by ISO/IEC 3166-1, used to set regulatory domain. Set
as needed to indicate country in which device is operating. This can limit
available channels and transmit power.

Possible values: see [this list](http://geotags.com/iso3166/countries.html).

Default value: empty

Example:

```
$ wifi-ap.config set wifi.country-code=US
```
