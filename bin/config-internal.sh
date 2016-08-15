#!/bin/bash
#
# Copyright (C) 2015, 2016 Canonical Ltd
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License version 3 as
# published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

DISABLED=1
DEBUG=1

# Default configuration
WIFI_INTERFACE=wlan0
WIFI_ADDRESS=192.168.7.1
WIFI_INTERFACE_MODE=uap

WIFI_HOSTAPD_DRIVER="nl80211"

WIFI_SSID="Ubuntu"

# Can be 'open' or 'wpa2'
WIFI_SECURITY="open"
# WIFI_SECURITY="wpa2"
# WIFI_SECURITY_PASSPHRASE="Ubuntu"

WIFI_CHANNEL=6
# Operation mode (a = IEEE 802.11a (5 GHz), b = IEEE 802.11b (2.4 GHz),
# g = IEEE 802.11g (2.4 GHz), ad = IEEE 802.11ad (60 GHz);
WIFI_OPERATION_MODE="g"

ETHERNET_INTERFACE=eth0

DHCP_RANGE_START=192.168.7.5
DHCP_RANGE_STOP=192.168.7.100
DHCP_LEASE_TIME="12h"

# We allow the user to place two configuration files. One which
# he can provide on its own in $SNAP_USER_DATA/config and one
# which only our scripts will modify in $SNAP_DATA/config
if [ -e "$SNAP_DATA/config" ] ; then
	. $SNAP_DATA/config
fi
if [ -e "$SNAP_USER_DATA/config" ] ; then
	. $SNAP_USER_DATA/config
fi
