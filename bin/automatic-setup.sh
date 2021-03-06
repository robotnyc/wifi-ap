#!/bin/sh
#
# Copyright (C) 2016 Canonical Ltd
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

# Wait for the snap to be successfully setup. This will only be true
# when the snap is started the first time and the configure hook was
# never called before.
while [ ! -e $SNAP_COMMON/.setup_done ]; do
    sleep 0.5
done

[ -f "$SNAP_COMMON/.block_auto_wizard" ] && exit 0

[ -f "$SNAP_DATA/config" ] && exit 0

while ! $SNAP/bin/client status; do
    sleep .5
done

exec $SNAP/bin/client wizard --auto
