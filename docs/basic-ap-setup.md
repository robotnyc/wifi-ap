---
title: "Basic WiFi AP Setup"
table_of_contents: True
---

# Basic Access Point Setup

The wifi-ap snap will try to automatically configure the access point after installation with the best options it can automatically determine. You can check with

```
$ wifi-ap.status
```

if the automatic configuration was successfully and the access point is already active.

To enable the WiFi access point manually after the snap was installed and all plugs and slots are connected run the following command:

```
$ wifi-ap.config set disabled=0
```

This will mark the access point as being enabled.

Now you have an access point with the SSID Ubuntu spawned up and any WiFi device can connect to it.
