---
title: "Snap Configuration"
table_of_contents: False
---

# Snap Configuration

In parallel to its own configuration system the wifi-ap snap provides a set of
snap configuration items which can be changed through the *snap set* system
command. These allow customization of the default behaviour of the wifi-ap snap
from a device [gadget snap](https://docs.ubuntu.com/core/en/reference/gadget).

The available configuration items are documented in the following sections.

## automatic-setup.disable

The *automatic-setup.disable* option allows a device to disable the automatic
AP setup that the wifi-ap snap comes with by default.

Possible values are:

 * *false (default):* The automatic setup of the AP on installation of the wifi-ap
   snap is **enabled**.
 * *true:* The automatic setup of the AP on installation of the wifi-ap snap is
   **disabled**.

Example:

```
$ snap set wifi-ap automatic-setup.disable=true
```

Please note that changing the configuration after installation of the wifi-ap
snap does not change its behaviour anymore. The option only has influence on the
snap when it is being installed. This can be used from a gadget snap by adding the
following lines to the gadget.yaml file:

```
defaults:
  # The alpha numeric key below is the id of the wifi-ap snap assigned in the
  # Ubuntu Store. Specifying the snap name instead is not possible.
  2rGgvyaY0CCzlWuKAPwFtCWrgwkM8lqS:
    automatic-setup.disable: true
```

After this snippet is added to the gadget.yaml and an updated version of the
gadget snap is deployed onto the device the automatic setup, of the AP is disabled
once the wifi-ap is installed from the Ubuntu Store.

## default.wifi.ssid

The *default.wifi.ssid* option allows a device to specify the default name of
the wireless network (SSID). This name will be used when creating a network
unless changed by the user.

This option is used only on installation of the wifi-ap snap, and it is ignored
on updates. Therefore, doing *snap set* on it produces no effect.  It must be
set by setting it as a default in the [gadget
snap](https://docs.ubuntu.com/core/en/reference/gadget), for instance:

```
defaults:
  # The alpha numeric key below is the id of the wifi-ap snap assigned in the
  # Ubuntu Store. Specifying the snap name instead is not possible.
  2rGgvyaY0CCzlWuKAPwFtCWrgwkM8lqS:
    default.wifi.ssid: MyDeviceSSID
```

A gadget snap with this option must be installed before installing wifi-ap for
it to have effect.
