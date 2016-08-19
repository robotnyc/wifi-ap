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

DEFAULT_ACCESS_POINT_INTERFACE="ap0"

if [ "$DISABLED" == "1" ] ; then
	echo "Not starting as WiFi AP is disabled"
	exit 0
fi

generate_dnsmasq_config() {
	(
	iface=$WIFI_INTERFACE
	if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
		iface=$DEFAULT_ACCESS_POINT_INTERFACE
	fi

	cat<<-EOF
	port=53
	all-servers
	interface=$iface
	except-interface=lo
	listen-address=$WIFI_ADDRESS
	bind-interfaces
	dhcp-range=$DHCP_RANGE_START,$DHCP_RANGE_STOP,$DHCP_LEASE_TIME
	dhcp-option=6, $WIFI_ADDRESS
	EOF
	) > $SNAP_DATA/dnsmasq.conf
}

start_dnsmasq() {
	iface=$WIFI_INTERFACE

	# If WiFi interface is managed by ifupdown leave it as is
	if [ -e /etc/network/interfaces.d/$iface ]; then
		exit 0
	fi

	# Create our AP interface if required
	if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
		iface=$DEFAULT_ACCESS_POINT_INTERFACE
		$SNAP/bin/iw dev $WIFI_INTERFACE interface add $iface type __ap
		sleep 2
	fi

	# Initial wifi interface configuration
	ifconfig $iface up
	ifconfig $iface $WIFI_ADDRESS netmask 255.255.255.0
	sleep 2

	if [ "$SHARE_NETWORK_INTERFACE" != "none" ] ; then
		# Enable NAT to forward our network connection
		iptables --table nat --append POSTROUTING --out-interface $ETHERNET_INTERFACE -j MASQUERADE
		iptables --append FORWARD --in-interface $iface -j ACCEPT

		sysctl -w net.ipv4.ip_forward=1
	fi

	$SNAP/bin/dnsmasq -k -C $SNAP_DATA/dnsmasq.conf -l $SNAP_DATA/dnsmasq.leases -x $SNAP_DATA/dnsmasq.pid &
}

stop_dnsmasq() {
	DNSMASQ_PID=$(cat $SNAP_DATA/dnsmasq.pid)
	kill -TERM $DNSMASQ_PID
	wait $DNSMASQ_PID

	iface=$WIFI_INTERFACE
	if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
		iface=$DEFAULT_ACCESS_POINT_INTERFACE
	fi

	if [ "$SHARE_NETWORK_INTERFACE" != "none" ] ; then
		# flush forwarding rules out
		iptables --table nat --delete POSTROUTING --out-interface $ETHERNET_INTERFACE -j MASQUERADE
		iptables --delete FORWARD --in-interface $iface -j ACCEPT
	fi

	if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
		$SNAP/bin/iw dev $iface del
	fi

	# disable ipv4 forward
	sysctl -w net.ipv4.ip_forward=0
}

generate_dnsmasq_config
start_dnsmasq

iface=$WIFI_INTERFACE
if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
	iface=$DEFAULT_ACCESS_POINT_INTERFACE
fi

# Wait a bit until our WiFi network interface is correctly
# setup by dnsmasq
grep $iface /proc/net/dev &> /dev/null
while [ $? != 0 ] ; do
	sleep 5
	grep $iface /proc/net/dev &> /dev/null
done

driver=$WIFI_HOSTAPD_DRIVER
if [ "$driver" == "rtl8188" ] ; then
	driver=rtl871xdrv
fi

# Generate our configuration file
cat <<EOF > $SNAP_DATA/hostapd.conf
interface=$iface
driver=$driver
channel=$WIFI_CHANNEL
macaddr_acl=0
ignore_broadcast_ssid=0
wmm_enabled=1
ieee80211n=1
ssid=$WIFI_SSID
hw_mode=$WIFI_OPERATION_MODE
EOF

case "$WIFI_SECURITY" in
	open)
		cat <<-EOF >> $SNAP_DATA/hostapd.conf
		auth_algs=1
		EOF
		;;
	wpa2)
		cat <<-EOF >> $SNAP_DATA/hostapd.conf
		auth_algs=3
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
	EXTRA_ARGS="$EXTRA_ARGS -ddd -t"
fi

hostapd=$SNAP/bin/hostapd
case "$WIFI_HOSTAPD_DRIVER" in
	rtl8188)
		hostapd=$SNAP/rtl8188/hostapd
		;;
	*)
		# Fallthrough and use the default hostapd
		;;
esac

# Startup hostapd with the configuration we've put in place
$hostapd $EXTRA_ARGS $SNAP_DATA/hostapd.conf &
HOSTAPD_PID=$!

trap exit_handler EXIT
function exit_handler() {
	kill -TERM $HOSTAPD_PID
	# Wait until hostapd is correctly terminated before we continue
	# doing anything
	wait $HOSTAPD_PID
	stop_dnsmasq
	exit 0
}

wait $HOSTAPD_PID
stop_dnsmasq