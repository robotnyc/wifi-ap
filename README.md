This fork provides a captive portable functionality.

[![Snap Status](https://build.snapcraft.io/badge/robotnyc/wifi-ap.svg)](https://build.snapcraft.io/user/robotnyc/wifi-ap)

# wifi-ap

This snap provided WiFi AP functionality out of the box.

Documentation is currently available at
https://docs.google.com/document/d/1vNu3fBqpOkBkjv_Vs9NZyTv50vOEfugrQqgxD0_f0rE/edit#

## Development

To modify any of the included services written in Go you need setup
your build environment first.

```
 $ snapcraft clean
 $ snapcraft
```

Now we need to export the GOPATH and point it to the directory
snapcraft already created for us.

```
 $ export GOPATH=$PWD/parts/management-service/go
```

Now you can build the management-service by running

```
 $ cd cmd/service
 $ go build -o management-service *.go
```

If you want to start it afterwards outside of a snap environment you
need to setup the right environment variables.

```
 # Needs to be the top source dir which contains the snapcraft.yaml
 $ export SNAP=`pwd`
 $ mkdir tmp-data
 $ export SNAP_DATA=$PWD/tmp-data
 $ cmd/service/management-service
```

## Running tests

We have a set of spread (https://github.com/snapcore/spread) tests which
can be executed on a virtual machine or real hardware.

In order to run those tests you need the follow things

 * ubuntu-image
 * spread

 You can install both as a snap

 $ snap install --edge --devmode ubuntu-image
 $ snap install --devmode spread

NOTE: As of today (27/10/2016) the version of spread in the store misses
some important bug fixes so you have to build your own one for now:

 $ WORKDIR=`mktemp -d`
 $ export GOPATH=$WORKDIR
 $ go get -d -v github.com/snapcore/spread/...
 $ go build github.com/snapcore/spread/cmd/spread
 $ sudo cp spread /usr/local/bin

Make sure /usr/local/bin is in your path and is used as default:

 $ which spread
 /usr/local/bin/spread

Now you have everything to run the test suite.

  $ ./run-tests

The script will create an image via ubuntu-image and make it available
to spread by copying it to ~/.spread/qemu or ~/snap/spread/<version>/.spread/qemu
depending on if you're using a local spread version or the one from the
snap.

If you want to see more verbose debugging output of spread run

 $ ./run-tests --debug
