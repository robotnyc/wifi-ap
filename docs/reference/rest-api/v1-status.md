---
title: "/v1/configuration"
table_of_contents: False
---

## GET /v1/status

### Description

Retrieve various status information about the access point.

### Request

None

### Response

```
{
  “ap.active”: <boolean>
}
```

### Errors

The following errors can occur:

 * invalid-value
 * invalid-format


### Example

```
$ sudo unixhttpc /var/snap/wifi-ap/current/sockets/control  /v1/status
{
  “result”: {
     “ap.active”: “0”,
  },
  “status”: “OK”,
  “status-code”: 200,
  “type”: “sync”
}
```

## POST /v1/status

### Description

Change the status of the access point.

### Request

```
{
	"action": <string>
}
```

Operation to perform.

Possible values are:

| Value         | Description              |
|---------------|--------------------------|
| *restart-ap*  | Restart the Access Point |


### Response

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
$ sudo unixhttpc -d '{“action”:”restart-ap”}' /var/snap/wifi-ap/current/sockets/control /v1/status
{
  “result”: { },
  “status”: “OK”,
  “status-code”: 200,
  “type”: “sync”
}
```
