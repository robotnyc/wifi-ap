---
title: "Running a secure Access Point"
table_of_contents: True
---

# Running a secure Access Point

If you want to secure the WiFi access point with WPA 2 personal you have to change two configuration items.

```
$ wifi-ap.config set wifi.security=wpa2 wifi.security-passphrase=Test1234
```

This enables WPA2 security with the passphrase set to *Test1234*.

**WARNING:** remember to always quote or escape the value when it contains special characters or spaces, eg. 'My WiFi', 'Pa$$word' or "Alan's AP"
