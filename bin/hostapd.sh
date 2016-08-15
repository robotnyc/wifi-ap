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

set -x

. $SNAP/bin/config-internal.sh

if [ "$DISABLED" == "1" ] ; then
	echo "Not starting as WiFi AP is disabled"
	exit 0
fi

iface=$WIFI_INTERFACE
if [ "$WIFI_INTERFACE_MODE" == "uap" ] ; then
	iface="uap0"
fi

# Wait a bit until our WiFi network interface is correctly
# setup by dnsmasq
grep $iface /proc/net/dev &> /dev/null
while [ $? != 0 ] ; do
	sleep 5
	grep $iface /proc/net/dev &> /dev/null
done

# Generate our configuration file
cat <<EOF > $SNAP_DATA/hostapd.conf
interface=$iface
driver=$WIFI_HOSTAPD_DRIVER
channel=$WIFI_CHANNEL
macaddr_acl=0
ignore_broadcast_ssid=0
wmm_enabled=1
ieee80211n=1
ssid=$WIFI_SSID
hw_mode=$WIFI_OPERATION_MODE
# Enable 40MHz channels with 20ns guard interval
ht_capab=[HT40][SHORT-GI-20][DSSS_CCK-40]
EOF

case "$WIFI_SECURITY" in
	open)
		cat <<-EOF >> $SNAP_DATA/hostapd.conf
		auth_algs=1
		EOF
		;;
	wpa2)
		auth_algs=2
		cat <<-EOF >> $SNAP_DATA/hostapd.conf
		wpa=2
		wpa_key_mgmt=WPA-PSK
		wpa_passphrase=$WIFI_SECURITY_PASSPHRASE
		wpa_pairwise=TKIP
		rsn_pairwise=CCMP
		EOF
		;;
	*)
		echo "Unsupported WiFi security '$WIFI_SECURITY' selected"
		exit 1
esac

EXTRA_ARGS=
if [ "$DEBUG" == "1" ] ; then
	cat $SNAP_DATA/hostapd.conf
	EXTRA_ARGS="$EXTRA_ARGS -ddd -t"
fi

exec $SNAP/usr/sbin/hostapd $EXTRA_ARGS $SNAP_DATA/hostapd.conf
