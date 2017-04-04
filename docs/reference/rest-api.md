---
title: "REST API"
table_of_contents: True
---

# REST API

The WiFi access point can be configured via a REST API service. This chapter describes the REST API and gives examples of how it can be used to change the configuration.

## API Versioning

The API is versioned. Every endpoint is prefixed with a version number. For example: /v1/configuration. The version number will be incremented with any change which breaks the current API version.

## Connection

The service will listen by default on an unix domain socket located in $SNAP_DATA/sockets/control , inside the wifi-ap snap itself, and provides a REST API endpoint.
Another snap needs to use the content interface to get access to the Unix domain socket. The plug it needs to declare can look like this:

```
name: wifi-ap-example-consumer
[...]
plugs:
  control:
    interface: content
    content: socket-directory
    target: $SNAP_DATA/sockets
[...]
```

Once the consuming snap is installed its plug needs to be connected to the wifi-ap:control slot to get access to the control socket:

```
$ snap connect wifi-ap-example-consumer:control wifi-ap:control
```
The socket will be available as $SNAP_DATA/sockets/control within the wifi-ap-example-consumer snap.

If you need a simple client to talk with the service, you can use, for example, the wifi-ap-client snap to do raw HTTP queries. You can install it with:

```
 $ snap install wifi-ap-client
 $ snap connect ..
```

## Responses

All responses are application/json unless noted otherwise. There are two return types:

 * Standard return value
 * Error

Status codes follow that of HTTP. Standard operation responses are capable of returning additional metadata key/values as part of the returned JSON object.

### Standard return value

For a standard synchronous operation, the following JSON object is returned:

```
{
 "result": {},               // Extra resource/action specific data
 "status": "OK",
 "status-code": 200,
 "type": "sync"
}
```

The HTTP code will be 200 (OK), or 201 (created, in which case the Location HTTP header will be set), as appropriate.

### Error

There are various situations in which something may immediately go wrong. In those cases, the following return value is used:

```
{
	 "result": {
		"message": "human-friendly description of the cause of the error",
   		"kind": "internal-error",  // one of a list of kinds (TBD), only present if "value" is present
	 	"value": {"...": "..."} // kind-specific object, as required
 	},
	"status": "Bad Request", // text description of status-code
"status-code": 400,      // or 401, etc. (same as HTTP code)
	"type": "error"
}
```

HTTP code must be one of of 400, 401, 403, 404, 405, 409, 412 or 500.

Possible error kinds are

| **kind** | **Value description **|
|----------|-----------------------|
|internal-error|An internal error occurred, not possible to give more information of what happened.|
|invalid-value|An invalid value was provided as input parameter.|
|invalid-format|Invalid data format.|
