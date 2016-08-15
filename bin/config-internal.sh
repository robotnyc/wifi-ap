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

. $SNAP/conf/default-config

# We allow the user to place two configuration files. One which
# he can provide on its own in $SNAP_USER_DATA/config and one
# which only our scripts will modify in $SNAP_DATA/config
if [ -e "$SNAP_DATA/config" ] ; then
	. $SNAP_DATA/config
fi
if [ -e "$SNAP_USER_DATA/config" ] ; then
	. $SNAP_USER_DATA/config
fi
