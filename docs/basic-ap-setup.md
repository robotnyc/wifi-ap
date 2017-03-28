---
title: "Basic WiFi AP Setup"
table_of_contents: True
---

# Basic Access Point Setup

The wifi-ap snap will try to automatically configure the access point after installation with the best options it can automatically determine. You can check with

```
$ wifi-ap.status
```

if the automatic configuration was successful and the access point is already active.

By default the automatic wizard, which runs when the wifi-ap snap is installed,
will choose secure enough default password and will enable WPA2-PSK security.
You can find the selected password when logged into
the system the wifi-ap snap installed on by running the command:
```
$ sudo wifi-ap.config get wifi.security-passphrase
```

To enable the WiFi access point manually after the snap was installed and all plugs and slots are connected run the following command:

```
$ wifi-ap.config set disabled=false
```

This will mark the access point as being enabled.

Now you have an access point with the SSID Ubuntu spawned up and any WiFi device can connect to it.
