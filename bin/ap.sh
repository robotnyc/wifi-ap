#!/bin/bash
#
# Copyright (C) 2015-2017 Canonical Ltd
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

if [ $(id -u) -ne 0 ] ; then
	echo "ERROR: $0 needs to be executed as root!"
	exit 1
fi

. $SNAP/bin/config-internal.sh

if [ $DEBUG = "true" ]; then
	set -x
fi

# Now after we have enabled debugging or not we can safely load
# all others necessary bits.
. $SNAP/bin/helper.sh

if [ $DISABLED = "true" ] ; then
	echo "Not starting as WiFi AP is disabled"
	exit 0
fi

DEFAULT_ACCESS_POINT_INTERFACE="ap0"

# Make sure the configured WiFi interface is really available before
# doing anything.
if ! ifconfig $WIFI_INTERFACE ; then
	echo "ERROR: WiFi interface $WIFI_INTERFACE is not available!"
	exit 1
fi

cleanup_on_exit() {
	read HOSTAPD_PID <$SNAP_DATA/hostapd.pid
	if [ -n "$HOSTAPD_PID" ] ; then
		kill -TERM $HOSTAPD_PID || true
		wait $HOSTAPD_PID
	fi

	read DNSMASQ_PID <$SNAP_DATA/dnsmasq.pid
	if [ -n "$DNSMASQ_PID" ] ; then
		# If dnsmasq is already gone don't error out here
		kill -TERM $DNSMASQ_PID || true
		wait $DNSMASQ_PID
	fi

	iface=$WIFI_INTERFACE
	if [ "$WIFI_INTERFACE_MODE" = "virtual" ] ; then
		iface=$DEFAULT_ACCESS_POINT_INTERFACE
	fi

	if [ $SHARE_DISABLED = "false" ] ; then
		# flush forwarding rules out
		iptables --table nat --delete POSTROUTING --out-interface $SHARE_NETWORK_INTERFACE -j MASQUERADE
		iptables --delete FORWARD --in-interface $iface -j ACCEPT
		sysctl -w net.ipv4.ip_forward=0
	fi

	if is_nm_running ; then
		# Hand interface back to network-manager. This will also trigger the
		# auto connection process inside network-manager to get connected
		# with the previous network.
		$SNAP/bin/nmcli d set $iface managed yes
	fi

	if [ "$WIFI_INTERFACE_MODE" = "virtual" ] ; then
		$SNAP/bin/iw dev $iface del
	fi
}

# We need to install this right before we do anything to
# ensure that we cleanup everything again when we termiante.
trap cleanup_on_exit TERM

iface=$WIFI_INTERFACE
if [ "$WIFI_INTERFACE_MODE" = "virtual" ] ; then
	iface=$DEFAULT_ACCESS_POINT_INTERFACE

	# Make sure if the real wifi interface is connected we use
	# the same channel for our AP as otherwise the created AP
	# will not work.
	channel_in_use=$($SNAP/bin/iw dev $WIFI_INTERFACE info |awk '/channel/{print$2}')
	if [ -z "$channel_in_use" ]; then
		echo "WARNING: WiFi is currently not connected so we can't determine"
		echo "         which channel we can use for the AP. This will most"
		echo "         likely lead to failed connections when the STA gets"
		echo "         connected."
	elif [ "$channel_in_use" != "$WIFI_CHANNEL" ] ; then
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
	if [ $? -ne 0 ] ; then
		echo "ERROR: Failed to create virtual WiFi network interface"
		cleanup_on_exit
	fi
	wait_until_interface_is_available $iface
fi
if [ "$WIFI_INTERFACE_MODE" = "direct" ] ; then
	# If WiFi interface is managed by ifupdown or network-manager leave it as is
	assert_not_managed_by_ifupdown $iface
fi


if is_nm_running ; then
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

	if is_nm_running ; then
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

if [ $SHARE_DISABLED = "false" ] ; then
	# Enable NAT to forward our network connection
	iptables --table nat --append POSTROUTING --out-interface $SHARE_NETWORK_INTERFACE -j MASQUERADE
	iptables --append FORWARD --in-interface $iface -j ACCEPT
	sysctl -w net.ipv4.ip_forward=1
fi

generate_dnsmasq_config $SNAP_DATA/dnsmasq.conf
$SNAP/bin/dnsmasq \
	-k \
	-C $SNAP_DATA/dnsmasq.conf \
	-l $SNAP_DATA/dnsmasq.leases \
	-x $SNAP_DATA/dnsmasq.pid \
	-u root -g root \
	&

driver=$WIFI_HOSTAPD_DRIVER

# Generate our hostapd configuration file
cat <<EOF > $SNAP_DATA/hostapd.conf
interface=$iface
driver=$driver
channel=$WIFI_CHANNEL
macaddr_acl=0
ignore_broadcast_ssid=0
ieee80211n=1
ssid=$WIFI_SSID
auth_algs=1
utf8_ssid=1
hw_mode=$WIFI_OPERATION_MODE
# DTIM 3 is a good tradeoff between powersave and latency
dtim_period=3

# The wmm_* options are needed to enable AMPDU
# and get decent 802.11n throughput
# UAPSD is for stations powersave
uapsd_advertisement_enabled=1
wmm_enabled=1
wmm_ac_bk_cwmin=4
wmm_ac_bk_cwmax=10
wmm_ac_bk_aifs=7
wmm_ac_bk_txop_limit=0
wmm_ac_bk_acm=0
wmm_ac_be_aifs=3
wmm_ac_be_cwmin=4
wmm_ac_be_cwmax=10
wmm_ac_be_txop_limit=0
wmm_ac_be_acm=0
wmm_ac_vi_aifs=2
wmm_ac_vi_cwmin=3
wmm_ac_vi_cwmax=4
wmm_ac_vi_txop_limit=94
wmm_ac_vi_acm=0
wmm_ac_vo_aifs=2
wmm_ac_vo_cwmin=2
wmm_ac_vo_cwmax=3
wmm_ac_vo_txop_limit=47
wmm_ac_vo_acm=0
EOF

if [ -n "$WIFI_COUNTRY_CODE" ] ; then
	cat <<-EOF >> $SNAP_DATA/hostapd.conf
	# Regulatory domain options
	country_code=$WIFI_COUNTRY_CODE
	# Send country code in beacon frames
	ieee80211d=1
	# Enable radar detection
	ieee80211h=1
	# Send power constraint IE, 3dB below maximum allowed transmit power
	local_pwr_constraint=3
	# End reg domain options
	EOF
else
	cat <<-EOF >> $SNAP_DATA/hostapd.conf
	# Regulatory domain options
	# Country code set to global
	country_code=XX
	# End reg domain options
	EOF
fi

case "$WIFI_SECURITY" in
	open)
		cat <<-EOF >> $SNAP_DATA/hostapd.conf
		EOF
		;;
	wpa2)
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
if [ "$DEBUG" = "true" ] ; then
	EXTRA_ARGS="$EXTRA_ARGS -ddd -t"
fi

hostapd=$SNAP/bin/hostapd

# Startup hostapd with the configuration we've put in place
$hostapd $EXTRA_ARGS $SNAP_DATA/hostapd.conf &
hostapd_pid=$!
echo $hostapd_pid > $SNAP_DATA/hostapd.pid
wait $hostapd_pid

cleanup_on_exit
exit 0
