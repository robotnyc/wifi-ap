#!/bin/sh
. $TESTSLIB/utilities.sh

# Powercycle both interface to get them back into a sane state before
# we install the wifi-ap snap
snap install --devmode wireless-tools
for d in wlan0 wlan1 ; do
    phy=$(iw dev $d info | awk '/wiphy/{print $2}')
    /snap/bin/wireless-tools.rfkill block $phy
    /snap/bin/wireless-tools.rfkill unblock $phy
done
snap remove wireless-tools

install_snap_under_test
# Give wifi-ap a bit time to settle down to avoid clashed
sleep 5
