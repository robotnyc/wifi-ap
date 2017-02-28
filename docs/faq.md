---
title: "FAQ"
table_of_contents: True
---

# FAQ

This section covers some of the most commonly encountered problems and attempts
to resolve them.

## Can I run an AP in parallel to being connected to another WiFi network?

Yes this is possible. However, there are known limitations if the device only has
one WiFi network device. (e.g. both need to operate on the same channel). See
[Simultaneous STA / AP Mode](simultaneous-sta-ap-mode.md) for more details.

## Why isn't the AP automatically enabled after I've installed the snap?

Normally it should automatically come up if this isn't disabled through the
device configuration inside the gadget snap. See
[Snap Configuration](reference/snap-configuration.md#automatic-setup.disable)
for more details.

If the above doesn't help then its most likely that the automatic setup couldn't
find a good configuration for your device and you have to manually configure the AP.
See [Configuration](reference/configuration.md) for details on this.

If this still doesn't help, feel free to file a [bug report](report-bug.md).
