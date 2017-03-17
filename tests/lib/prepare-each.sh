#!/bin/sh
. $TESTSLIB/utilities.sh

# Powercycle both interface to get them back into a sane state before
# we install the wifi-ap snap
for d in wlan0 wlan1 ; do
    phy=$(iw dev $d info | awk '/wiphy/{print $2}')
    snap install --devmode wireless-tools
    /snap/bin/wireless-tools.rfkill block $phy
    /snap/bin/wireless-tools.rfkill unblock $phy
done

install_snap_under_test
