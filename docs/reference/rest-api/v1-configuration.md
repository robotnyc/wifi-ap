---
title: "/v1/configuration"
table_of_contents: False
---

## GET /v1/configuration

### Description

Retrieve all available or a single configuration items and their current values.

### Request

*keys* [optional]

An array of config item keys to return the value for. If not supplied or empty, all available configuration items will be returned.

### Response

| Response attributes

```
{
  “<config item key>”: “<config item value>”,
  ...
}
```

Each key/value pair in the in the result corresponds to one configuration item. There are no further special fields.

### Errors

The following errors can occur:

 * invalid-value
 * invalid-format

### Example

```
$ sudo wifi-ap-client /v1/configuration
{
  “result”: {
"disabled": true,
"debug": false,
"wifi.interface": "wlan0",
"wifi.address": "192.168.7.1",
"wifi.netmask": "255.255.255.0",
"wifi.interface-mode": "direct",
"wifi.hostapd-driver": "nl80211",
"wifi.ssid": "Ubuntu",
"wifi.security": "wpa2",
"wifi.security-passphrase": "12345678",
"wifi.channel": 6,
"wifi.country-code": "",
"wifi.operation-mode": "virtual",
"share.disabled": false,
"share.network-interface": "eth0",
"dhcp.range-start": "192.168.7.50",
"dhcp.range-stop": "192.168.7.200",
"dhcp.lease-time": "12h"
  },
  “status”: “OK”,
  “status-code”: 200,
  “type”: “sync”
}
```
</br>
## POST /v1/configuration

### Description

Change the value of one or multiple configuration items. When all configuration changes are applied, the AP will be restarted and all currently connected clients are disconnected.

### Request

A dictionary of key/value pairs corresponding to configuration items to change.

If multiple key/value pairs are supplied as parameter, the service will apply either all or nothing to ensure that the configuration stays in a known state.

### Result

```
{ }
```

The result does not contain any field.

### Errors

The following errors can occur:

 * invalid-value
 * invalid-format

### Example

```
$ sudo wifi-ap-client -d '{“wifi.security”: “open”, “wifi.interface”: “wlan0”}' /v1/configuration
{
  “result”: { },
  “status”: “OK”,
  “status-code”: 200,
  “type”: “sync”
}
```
