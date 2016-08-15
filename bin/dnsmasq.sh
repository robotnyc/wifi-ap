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

if [ -z "$@" ] ; then
	echo "Usage: $0 start|stop|show-config"
	exit 1
fi

generate_dnsmasq_config() {
(
	iface=$WIFI_INTERFACE
	if [ "$WIFI_INTERFACE_MODE" == "uap" ] ; then
		iface="uap0"
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

	# Create our AP interface
	if [ "$WIFI_INTERFACE_MODE" == "uap" ] ; then
		iface="uap0"
		$SNAP/sbin/iw dev $WIFI_INTERFACE interface add $iface type __ap
		sleep 2
	fi

	# Initial wifi interface configuration
	ifconfig $iface up $WIFI_ADDRESS netmask 255.255.255.0
	sleep 2

	# Enable NAT
	iptables --flush
	iptables --table nat --flush
	iptables --delete-chain
	iptables --table nat --delete-chain
	iptables --table nat --append POSTROUTING --out-interface $ETHERNET_INTERFACE -j MASQUERADE
	iptables --append FORWARD --in-interface $iface -j ACCEPT

	sysctl -w net.ipv4.ip_forward=1

	exec $SNAP/usr/local/sbin/dnsmasq -k -C $SNAP_DATA/dnsmasq.conf -l $SNAP_DATA/dnsmasq.leases -x $SNAP_DATA/dnsmasq.pid
}

stop_dnsmasq() {
	iface=$WIFI_INTERFACE

	if [ "$WIFI_INTERFACE_MODE" == "uap" ] ; then
		$SNAP/sbin/iw dev uap0 del
	fi

	# flush forwarding rules out
	iptables --table nat --delete POSTROUTING --out-interface $NETWORK_ETH -j MASQUERADE
	iptables --delete FORWARD --in-interface $iface -j ACCEPT

	# disable ipv4 forward
	sysctl -w net.ipv4.ip_forward=0
}

run_commands() {
	while [ -n "$1" ] ; do
		case "$1" in
		start)
			generate_dnsmasq_config
			start_dnsmasq
			shift
			;;
		stop)
			stop_dnsmasq
			shift
			;;
		*)
			echo "Unknown command '$1'."
			shift
			;;
		esac
	done
}

run_commands $@
