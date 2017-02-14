---
title: "/v1/configuration"
table_of_contents: True
---

# /v1/status

# GET

## Description

Retrieve various status information about the access point.

## Parameters

None

## Result

```
{
  “ap.active”: <boolean>
}
```

## Errors

The following errors can occur:

 * invalid-value
 * invalid-format


## Example

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

# POST

## Description

Change the status of the access point.

## Parameters

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


## Result

```
{ }
```

The result does not contain any field.

## Errors

The following errors can occur:

 * invalid-value
 * invalid-format


## Example

```
$ sudo unixhttpc -d '{“action”:”restart-ap”}' /var/snap/wifi-ap/current/sockets/control /v1/status
{
  “result”: { },
  “status”: “OK”,
  “status-code”: 200,
  “type”: “sync”
}
```
