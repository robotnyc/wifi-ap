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

does_interface_exist() {
	grep -q $1 /proc/net/dev
}

wait_until_interface_is_available() {
	while ! does_interface_exist $iface; do
		# Wait for 10ms
		sleep 0.01
	done
}

assert_not_managed_by_ifupdown() {
	if [ -e /etc/network/interfaces.d/$1 ]; then
		echo "ERROR: Interface $1 is managed by ifupdown and can't be used"
		exit 1
	fi
}

generate_dnsmasq_config() {
	{
	iface=$WIFI_INTERFACE
	if [ "$WIFI_INTERFACE_MODE" = "virtual" ] ; then
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
	} > $1
}

is_nm_running() {
	nm_status=`$SNAP/bin/nmcli -t -f RUNNING general`
	[ "$nm_status" = "running" ]
}
