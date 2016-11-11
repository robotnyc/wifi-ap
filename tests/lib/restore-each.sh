#!/bin/bash

. $TESTSLIB/snap-names.sh

# Remove all snaps not being the core, gadget, kernel or snap we're testing
for snap in /snap/*; do
	snap="${snap:6}"
	case "$snap" in
		"bin" | "$gadget_name" | "$kernel_name" | "$core_name" | "$SNAP_NAME" )
			;;
		*)
			snap remove "$snap"
			;;
	esac
done

# Cleanup all configuration files from the snap so that we have
# a fresh start for the next test
rm -rf /var/snap/$SNAP_NAME/common/*
rm -rf /var/snap/$SNAP_NAME/current/*
systemctl stop snap.wifi-ap.backend
systemctl stop snap.wifi-ap.management-service

# Drop any generated or modified netplan configuration files. The original
# ones will be restored below.
rm -f /etc/netplan/*

# Ensure we have the same state for snapd as we had before
systemctl stop snapd.service snapd.socket
rm -rf /var/lib/snapd/*
$(cd / && tar xzf $SPREAD_PATH/snapd-state.tar.gz)
rm -rf /root/.snap
systemctl start snapd.service snapd.socket

# Make sure the original netplan configuration is applied and active
netplan generate
netplan apply

# Start services again now that the system is restored
systemctl start snap.wifi-ap.backend
systemctl start snap.wifi-ap.management-service
