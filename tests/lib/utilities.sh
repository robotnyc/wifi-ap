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
