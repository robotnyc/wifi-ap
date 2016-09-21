# -*- Mode:Python; indent-tabs-mode:nil; tab-width:4 -*-
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

import os

import snapcraft
import snapcraft.plugins.go

DEBIAN_GOCODE_SRC_PATH = "/usr/share/gocode/src"

class GoPlugin(snapcraft.plugins.go.GoPlugin):
    def pull(self):
        # Skip the pull step from GoPlugin here as it would repeat 90% of what
        # we're doing here and will only conflict.
        snapcraft.BasePlugin.pull(self)
        os.makedirs(self._gopath_src, exist_ok=True)

        if self.options.source:
            go_package = self._get_local_go_package()
            go_package_path = os.path.join(self._gopath_src, go_package)
            if os.path.islink(go_package_path):
                os.unlink(go_package_path)
            os.makedirs(os.path.dirname(go_package_path), exist_ok=True)
            os.symlink(self.sourcedir, go_package_path)
            # We can't just add the gocode directory to GOPATH as go
            # doesn't like this. Instead we're symlinking the necessary
            # subdirectories from the gocode directory into our already
            # existing GOPATH.
            for dir in os.listdir(DEBIAN_GOCODE_SRC_PATH):
                os.symlink(os.path.join(DEBIAN_GOCODE_SRC_PATH, dir),
                           os.path.join(self._gopath_src, dir))
