#!/bin/sh

wait_for_systemd_service() {
	while ! systemctl status $1 ; do
		sleep 1
	done
	sleep 1
}

wait_for_systemd_service_exit() {
	while systemctl status $1 ; do
		sleep 1
	done
	sleep 1
}

does_interface_exist() {
	[ -d /sys/class/net/$1 ]
}

wait_until_interface_is_available() {
	while ! does_interface_exist $1; do
		# Wait for 200ms
		sleep 0.2
	done
}

install_snap_under_test() {
	# If we don't install wifi-ap here we get a system
	# without any network connectivity after reboot.
	if [ -n "$SNAP_CHANNEL" ] ; then
		# Don't reinstall if we have it installed already
		if ! snap list | grep wifi-ap ; then
			snap install --$SNAP_CHANNEL wifi-ap
		fi
	else
		# Install prebuilt wifi-ap snap
		snap install --dangerous /home/wifi-ap/wifi-ap_*_amd64.snap
		# As we have a snap which we build locally it's unasserted and therefore
		# we don't have any snap-declarations in place and need to manually
		# connect all plugs.
		snap connect wifi-ap:network-control core
		snap connect wifi-ap:network-bind core
		snap connect wifi-ap:network core
		snap connect wifi-ap:firewall-control core
	fi
}

# Connects to a WiFi network
# args1: interface to use
connect_to_wifi() {
	local enc ssid pass

	enc=$(/snap/bin/wifi-ap.config get wifi.security)
	ssid=$(/snap/bin/wifi-ap.config get wifi.ssid)
	if [ "$enc" = open ]; then
		iw "$1" connect "$ssid"
	else
		pass=$(/snap/bin/wifi-ap.config get wifi.security-passphrase)
		wpa_passphrase "$ssid" "$pass" >/tmp/wpa.conf
		wpa_supplicant -B -c/tmp/wpa.conf -i"$1"

		# Need to wait for a full scan and WPA2 handshake
		for i in $(seq 60); do
			iw dev "$1" link |fgrep -xq 'Not connected.' || break
			sleep 1
		done
	fi
}
