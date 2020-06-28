# Preface
This service provides a defined method to serve [iPXE](https://ipxe.org/) configuration from a single YAML configuration and is the successor of [ipxe-distributor](https://git.ypbind.de/cgit/ipxe-distributor/). Allowing for better portability, e.g. no dependency for pyYAML (which is not available on some distributions) the predecessor has been rewritten in Go. (And to avoid Python 2 -> 3 migration.)

# Configuration and command line options
## Command line parameters
| *Option* | *Description* | *Default* | *Comment* |
|:---------|:--------------|:---------:|:----------|
| `--config=<file>` | Read configuration from <file> | `/etc/ipxe-distributor/config.yaml` | - |
| `--help` | Show help text | - | - |
| `--test` | Test configuration file for syntax errors | - | Returns 0 on success, 1 on failure |
| `--version` | Show version information | - | - |

## Configuration
The configuration file must be in [YAML](https://yaml.org/) format and **have to** contain the dictionaries:

| *Dictionary* | *Description* |
|:-------------|:--------------|
| `default` | Default iPXE data to send if no route matches |
| `global` | Global configuration |
| `images` | Image tag and definitions |
| `nodes` | Node/host/group tags and definitions |

### Global configuration - `global`
The `global` dictionary (obviously) defines global configuration. Supported configuration keys are:

| *Option* | *Description* | *Default* | *Comment* |
|:---------|:--------------|:---------:|:----------|
| `ipxe_append` | Array of data to append to iPXE output | - | |
| `ipxe_prepend` | Array of data to prepend to iPXE output | - | The iPXE shebang (`#!ipxe`) is always prepended on each output |
| `url` | URL to listen for request, defines scheme, host, port and base path | `http://localhost:8080` | Only HTTP is supported at the moment |

### Default configuration - `default`
Only the `default_image` key is supported in the `default` dictionary. It defines the iPXE data to return when the default URL path (`/default`) is requested. This can be used to define a default menu (e.g. for BOOTP/TFTP or for an USB stick).
The `default_image` data must be defined as an array.

### List of image/image labels - `images`
In the `images` dictionary defines a list of "image" data - identified by their unique label - and their iPXE data defined in their array `action`

### Node definitions - `nodes`
The `nodes` dictionary contains a list of "node" data identified by their (unique) label and their selector data. Valid selectors are:

| *Selector* | *Description* | *Comment* |
|:-----------|:--------------|:----------|
| `image` | Which image label/action from `images` dictionary to use for this node | **Mandatory** |
| `group` | The name of the group this node belongs to | Used for group requests on `/group/<groupname>` |
|  | | `mac` / `serial` and `group` are mutually exclusive |
| `mac` | The MAC address of this node | Used for MAC requests on `/mac/<mac>` |
|       | | Format can be either:`00:11:22:33:44:55`, `00-11-22-33-44-55` or `001122334455` | 
| `serial` | The serial number of this node | Used for requests for the serial number on `/serial/<sn>` |

**Note:** `group`, `mac` and `serial` can't contain `/`

### Example
```yaml
---
global:
    url: "http://127.0.0.1:8080/own/path"
    # prepend to _all_ IPXE data
    ipxe_prepend:
        - |
            # ! *-timeout is in ms !
            set menu-timeout 60000
            set submenu-timeout ${menu-timeout}
            
            # set default action to "exit" if not set
            isset ${menu-default} || set menu-default exit_ipxe

    # append to _all_ IPXE data
    ipxe_append:
        - "choose --timeout ${menu-timeout} --default ${menu-default} selected || goto abort"
        - "set menu-timeout 0"

default:
    # default if there is no match
    default_image:
        - "# Boot the first local HDD"
        - "sanboot --no-describe --drive 0x80"

images:
    centos_7.6:
        action:
            - "initrd http://mirror.centos.org/centos/7/os/x86_64/images/pxeboot/initrd.img"
            - "chain http://mirror.centos.org/centos/7/os/x86_64/images/pxeboot/vmlinuz net.ifnames=0 biosdevname=0 ksdevice=eth2 inst.repo=http://mirror.centos.org/centos/7/os/x86_64/ inst.lang=en_GB inst.keymap=be-latin1 inst.vnc inst.vncpassword=CHANGEME ip=x.x.x.x netmask=x.x.x.x gateway=x.x.x.x dns=x.x.x.x"

    redhat_7:
        action:
            - "initrd http://mirror.redhat.org/redhat/7/os/x86_64/images/pxeboot/initrd.img"
            - "chain http://mirror.redhat.org/redhat/7/os/x86_64/images/pxeboot/vmlinuz net.ifnames=0 biosdevname=0 ksdevice=eth2 inst.repo=http://mirror.redhat.org/redhat/7/os/x86_64/ inst.lang=en_GB inst.keymap=be-latin1 inst.vnc inst.vncpassword=CHANGEME ip=x.x.x.x netmask=x.x.x.x gateway=x.x.x.x dns=x.x.x.x"

nodes:
    singleserver01:
        mac: 12:34:56:78:9a:bc
        image: "centos_7.6"
    singleserver02:
        mac: bc-9a-78-56-34-12
        image: "redhat_7"
        serial: SN12BC56789
    singleserver03:
        mac: cafebabebabe
        image: "default"
        serial: SX123456789
    servergroup:
        image: "redhat_7"
        group: "ldapservers"
```

## Requests and workflow
The server can handle different request types:

* request default data - `/default`
* request data for a group - `/group/<groupname>`
* request data for a MAC address - `/mac/<mac>`
* request data for a serial number - `/serial/<serialnumber>`

Requests for MAC addresses and serial numbers can be scripted in iPXE (for instance in the default data) by using the [mac](https://www.ipxe.org/cfg/mac) (e.g. `/mac/${net0/mac}`) or [serial](https://www.ipxe.org/cfg/serial) (e.g. `/serial/${serial}`) iPXE variables.

Upon start the service reads the configuration file, maps group names to node labels, MAC addresses to node labels and serial numbers to node labels. Additionally the image labels and their actions are also mapped.

### Request for default data
If a request for the default data (`/default`) is received the reply consists of:

* iPXE shebang - `#!ipxe`
* data from global configuration `ipxe_prepend`
* data from default configuration `default_image`
* data from global configuration `ipxe_append`

(in this order)

### Request for a group
If a request for a group is received (on `/group/<groupname>`) the requested group is searched in the group mapping. If found the image configured for this group will be searched in the image mapping. If the image was found it's configured `action` will be read. The final
reply consists of:

* iPXE shebang - `#!ipxe`
* data from global configuration `ipxe_prepend`
* data from image configuration `action`
* data from global configuration `ipxe_append`

### Request for a MAC address
If a request for a group is received (on `/mac/<macaddress>`) the requested group is searched in the MAC address mapping. If found the image configured for this MAC address will be searched in the image mapping. If the image was found it's configured `action` will be read. The final
reply consists of:

* iPXE shebang - `#!ipxe`
* data from global configuration `ipxe_prepend`
* data from image configuration `action`
* data from global configuration `ipxe_append`

### Request for a serial number
If a request for a serial number is received (on `/serial/<serialnumber>`) the requested serial number is searched in the serial number mapping. If found the image configured for this serial number will be searched in the image mapping. If the image was found it's configured `action` will be read. The final
reply consists of:

* iPXE shebang - `#!ipxe`
* data from global configuration `ipxe_prepend`
* data from image configuration `action`
* data from global configuration `ipxe_append`

----

# Licenses
## go-ipxe-distributor
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

## go-yaml (https://github.com/go-yaml/yaml)

                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.

      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.

      "Source" form shall mean the preferred form for making modifications,
      including but not limited to software source code, documentation
      source, and configuration files.

      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form, including but
      not limited to compiled object code, generated documentation,
      and conversions to other media types.

      "Work" shall mean the work of authorship, whether in Source or
      Object form, made available under the License, as indicated by a
      copyright notice that is included in or attached to the work
      (an example is provided in the Appendix below).

      "Derivative Works" shall mean any work, whether in Source or Object
      form, that is based on (or derived from) the Work and for which the
      editorial revisions, annotations, elaborations, or other modifications
      represent, as a whole, an original work of authorship. For the purposes
      of this License, Derivative Works shall not include works that remain
      separable from, or merely link (or bind by name) to the interfaces of,
      the Work and Derivative Works thereof.

      "Contribution" shall mean any work of authorship, including
      the original version of the Work and any modifications or additions
      to that Work or Derivative Works thereof, that is intentionally
      submitted to Licensor for inclusion in the Work by the copyright owner
      or by an individual or Legal Entity authorized to submit on behalf of
      the copyright owner. For the purposes of this definition, "submitted"
      means any form of electronic, verbal, or written communication sent
      to the Licensor or its representatives, including but not limited to
      communication on electronic mailing lists, source code control systems,
      and issue tracking systems that are managed by, or on behalf of, the
      Licensor for the purpose of discussing and improving the Work, but
      excluding communication that is conspicuously marked or otherwise
      designated in writing by the copyright owner as "Not a Contribution."

      "Contributor" shall mean Licensor and any individual or Legal Entity
      on behalf of whom a Contribution has been received by Licensor and
      subsequently incorporated within the Work.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      (except as stated in this section) patent license to make, have made,
      use, offer to sell, sell, import, and otherwise transfer the Work,
      where such license applies only to those patent claims licensable
      by such Contributor that are necessarily infringed by their
      Contribution(s) alone or by combination of their Contribution(s)
      with the Work to which such Contribution(s) was submitted. If You
      institute patent litigation against any entity (including a
      cross-claim or counterclaim in a lawsuit) alleging that the Work
      or a Contribution incorporated within the Work constitutes direct
      or contributory patent infringement, then any patent licenses
      granted to You under this License for that Work shall terminate
      as of the date such litigation is filed.

   4. Redistribution. You may reproduce and distribute copies of the
      Work or Derivative Works thereof in any medium, with or without
      modifications, and in Source or Object form, provided that You
      meet the following conditions:

      (a) You must give any other recipients of the Work or
          Derivative Works a copy of this License; and

      (b) You must cause any modified files to carry prominent notices
          stating that You changed the files; and

      (c) You must retain, in the Source form of any Derivative Works
          that You distribute, all copyright, patent, trademark, and
          attribution notices from the Source form of the Work,
          excluding those notices that do not pertain to any part of
          the Derivative Works; and

      (d) If the Work includes a "NOTICE" text file as part of its
          distribution, then any Derivative Works that You distribute must
          include a readable copy of the attribution notices contained
          within such NOTICE file, excluding those notices that do not
          pertain to any part of the Derivative Works, in at least one
          of the following places: within a NOTICE text file distributed
          as part of the Derivative Works; within the Source form or
          documentation, if provided along with the Derivative Works; or,
          within a display generated by the Derivative Works, if and
          wherever such third-party notices normally appear. The contents
          of the NOTICE file are for informational purposes only and
          do not modify the License. You may add Your own attribution
          notices within Derivative Works that You distribute, alongside
          or as an addendum to the NOTICE text from the Work, provided
          that such additional attribution notices cannot be construed
          as modifying the License.

      You may add Your own copyright statement to Your modifications and
      may provide additional or different license terms and conditions
      for use, reproduction, or distribution of Your modifications, or
      for any such Derivative Works as a whole, provided Your use,
      reproduction, and distribution of the Work otherwise complies with
      the conditions stated in this License.

   5. Submission of Contributions. Unless You explicitly state otherwise,
      any Contribution intentionally submitted for inclusion in the Work
      by You to the Licensor shall be under the terms and conditions of
      this License, without any additional terms or conditions.
      Notwithstanding the above, nothing herein shall supersede or modify
      the terms of any separate license agreement you may have executed
      with Licensor regarding such Contributions.

   6. Trademarks. This License does not grant permission to use the trade
      names, trademarks, service marks, or product names of the Licensor,
      except as required for reasonable and customary use in describing the
      origin of the Work and reproducing the content of the NOTICE file.

   7. Disclaimer of Warranty. Unless required by applicable law or
      agreed to in writing, Licensor provides the Work (and each
      Contributor provides its Contributions) on an "AS IS" BASIS,
      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
      implied, including, without limitation, any warranties or conditions
      of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
      PARTICULAR PURPOSE. You are solely responsible for determining the
      appropriateness of using or redistributing the Work and assume any
      risks associated with Your exercise of permissions under this License.

   8. Limitation of Liability. In no event and under no legal theory,
      whether in tort (including negligence), contract, or otherwise,
      unless required by applicable law (such as deliberate and grossly
      negligent acts) or agreed to in writing, shall any Contributor be
      liable to You for damages, including any direct, indirect, special,
      incidental, or consequential damages of any character arising as a
      result of this License or out of the use or inability to use the
      Work (including but not limited to damages for loss of goodwill,
      work stoppage, computer failure or malfunction, or any and all
      other commercial damages or losses), even if such Contributor
      has been advised of the possibility of such damages.

   9. Accepting Warranty or Additional Liability. While redistributing
      the Work or Derivative Works thereof, You may choose to offer,
      and charge a fee for, acceptance of support, warranty, indemnity,
      or other liability obligations and/or rights consistent with this
      License. However, in accepting such obligations, You may act only
      on Your own behalf and on Your sole responsibility, not on behalf
      of any other Contributor, and only if You agree to indemnify,
      defend, and hold each Contributor harmless for any liability
      incurred by, or claims asserted against, such Contributor by reason
      of your accepting any such warranty or additional liability.

   END OF TERMS AND CONDITIONS

   APPENDIX: How to apply the Apache License to your work.

      To apply the Apache License to your work, attach the following
      boilerplate notice, with the fields enclosed by brackets "{}"
      replaced with your own identifying information. (Don't include
      the brackets!)  The text should be enclosed in the appropriate
      comment syntax for the file format. We also recommend that a
      file or class name and description of purpose be included on the
      same "printed page" as the copyright notice for easier
      identification within third-party archives.

   Copyright {yyyy} {name of copyright owner}

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

## Gorilla web toolkit (http://www.gorillatoolkit.org/pkg/mux)
Copyright (c) 2012-2018 The Gorilla Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

     * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
     * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
     * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

## logrus (https://github.com/sirupsen/logrus)
The MIT License (MIT)

Copyright (c) 2014 Simon Eskildsen

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

