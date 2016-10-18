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

if [ $(id -u) -ne 0 ] ; then
	echo "ERROR: $0 needs to be executed as root!"
	exit 1
fi

. $SNAP/bin/config-internal.sh
. $SNAP/bin/helper.sh

DEFAULT_ACCESS_POINT_INTERFACE="ap0"

if [ $DISABLED -eq 1 ] ; then
	echo "Not starting as WiFi AP is disabled"
	exit 0
fi

# Make sure the configured WiFi interface is really available before
# doing anything.
if ! ifconfig $WIFI_INTERFACE ; then
	echo "ERROR: WiFi interface $WIFI_INTERFACE is not available!"
	exit 1
fi

cleanup_on_exit() {
	DNSMASQ_PID=$(cat $SNAP_DATA/dnsmasq.pid)
	kill -TERM $DNSMASQ_PID
	wait $DNSMASQ_PID

	iface=$WIFI_INTERFACE
	if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
		iface=$DEFAULT_ACCESS_POINT_INTERFACE
	fi

	if [ $SHARE_DISABLED -eq 0 ] ; then
		# flush forwarding rules out
		iptables --table nat --delete POSTROUTING --out-interface $SHARE_NETWORK_INTERFACE -j MASQUERADE
		iptables --delete FORWARD --in-interface $iface -j ACCEPT
		sysctl -w net.ipv4.ip_forward=0
	fi

	if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
		$SNAP/bin/iw dev $iface del
	fi
}

iface=$WIFI_INTERFACE
if [ "$WIFI_INTERFACE_MODE" == "virtual" ] ; then
	iface=$DEFAULT_ACCESS_POINT_INTERFACE

	# Make sure if the real wifi interface is connected we use
	# the same channel for our AP as otherwise the created AP
	# will not work.
	channel_in_use=$(iw dev $WIFI_INTERFACE info |awk '/channel/{print$2}')
	if [ $channel_in_use != $WIFI_CHANNEL ] ; then
		echo "ERROR: You configured a different channel than the WiFi device"
		echo "       is currently using. This will not work as most devices"
		echo "       require you to operate for AP and STA on the same channel."
		exit 1
	fi
fi

# Create our AP interface if required
if [ "$WIFI_INTERFACE_MODE" = "virtual" ] ; then
	iface=$DEFAULT_ACCESS_POINT_INTERFACE
	$SNAP/bin/iw dev $WIFI_INTERFACE interface add $iface type __ap
	sleep 2
fi
if [ "$WIFI_INTERFACE_MODE" = "direct" ] ; then
	# If WiFi interface is managed by ifupdown or network-manager leave it as is
	assert_not_managed_by_ifupdown $iface
fi


nm_status=`$SNAP/bin/nmcli -t -f RUNNING general`
if [ "$nm_status" = "running" ] ; then
	# Prevent network-manager from touching the interface we want to use. If
	# network-manager was configured to use the interface its nothing we want
	# to prevent here as this is how the user configured the system.
	$SNAP/bin/nmcli d set $iface managed no
fi

# Initial wifi interface configuration
ifconfig $iface up
if [ $? -ne 0 ] ; then
	echo "ERROR: Failed to enable WiFi network interface '$iface'"

	# Remove virtual interface again if we created one
	if [ "$WIFI_INTERFACE_MODE" = "virtual" ] ; then
		$SNAP/bin/iw dev $iface del
	fi

	if [ "$nm_status" = "running" ] ; then
		# Hand interface back to network-manager. This will also trigger the
		# auto connection process inside network-manager to get connected
		# with the previous network.
		$SNAP/bin/nmcli d set $iface managed yes
	fi

	exit 1
fi

# Configure interface and give it a moment to settle
ifconfig $iface $WIFI_ADDRESS netmask $WIFI_NETMASK
sleep 2

if [ $SHARE_DISABLED -eq 0 ] ; then
	# Enable NAT to forward our network connection
	iptables --table nat --append POSTROUTING --out-interface $SHARE_NETWORK_INTERFACE -j MASQUERADE
	iptables --append FORWARD --in-interface $iface -j ACCEPT
	sysctl -w net.ipv4.ip_forward=1
fi

generate_dnsmasq_config $SNAP_DATA/dnsmasq.conf
$SNAP/bin/dnsmasq -k -C $SNAP_DATA/dnsmasq.conf -l $SNAP_DATA/dnsmasq.leases -x $SNAP_DATA/dnsmasq.pid &

# Wait a bit until our WiFi network interface is correctly
# setup by dnsmasq
wait_until_interface_is_available $iface

driver=$WIFI_HOSTAPD_DRIVER
if [ "$driver" == "rtl8188" ] ; then
	driver=rtl871xdrv
fi

# Generate our hostapd configuration file
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
	cleanup_on_exit
	exit 0
}

wait $HOSTAPD_PID
cleanup_on_exit
