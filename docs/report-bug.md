---
title: "Report a Bug"
table_of_contents: False
---

# Rebort a Bug

Bugs can be reported [here](https://bugs.launchpad.net/snappy-hwe-snaps/+filebug).

When submitting a bug report, please attach:

 * */var/log/syslog*

And the output of the following two commands:

```
$ wifi-ap.config get
$ wifi-ap.status
$ journalctl --no-pager -u snap.wifi-ap.management-service
```

**NOTE:** The above commands will most likely print out your configured AP SSID
and passphrases. Please remove them from the output before attaching it to a
bug report.
