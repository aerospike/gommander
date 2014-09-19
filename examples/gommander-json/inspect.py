#!/usr/bin/python
# -*- coding: utf-8 -*-
################################################################################
# Copyright 2013-2014 Aerospike, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################

from __future__ import print_function
from json import dump
from os import path, uname
from subprocess import Popen, PIPE
from sys import stdout, exit, platform

#-------------------------------------------------------------------------------

result = {}
result['flags'] = {}

#-------------------------------------------------------------------------------

def checkPlatform():
    result['platform'] = platform
    if result['platform'].startswith('linux'):
        result['flags']['linux'] = True

def checkUname():
    (sysname, nodename, release, version, machine) = uname()

    result['uname'] = {}
    result['uname']['sysname'] = sysname.lower()
    result['uname']['nodename'] = nodename.lower()
    result['uname']['release'] = release.lower()
    result['uname']['version'] = version.lower()
    result['uname']['machine'] = machine.lower()

    result['flags']['linux'] = result['uname']['sysname'] == 'linux'

def checkLSB():
    if not result['flags']['linux']:
        return

    p = Popen("lsb_release -a", shell=True, stdin=PIPE, stdout=PIPE, stderr=PIPE, close_fds=True)
    (out, err) = p.communicate()
    if p.returncode == 0:
        lsb_raw = out.lower().strip()
        result['lsb'] = {}
        for line in lsb_raw.split('\n'):
            (k,v) = line.split('\t',2)
            result['lsb'][k.replace(' ','_')] = v

def checkRedHat():
    if path.isfile('/etc/redhat-release'):
        result['redhat'] = {}

        with open('/etc/redhat-release', 'r') as content:
            result['redhat']['version'] = content.read().lower().strip()
            result['flags']['redhat'] = True

        if result['redhat']['version'].find("centos") >= 0:
            result['redhat']['centos'] = True
            result['redhat']['distro'] = 'centos'
        elif result['redhat']['version'].find("redhat") >= 0:
            result['redhat']['redhat'] = True
            result['redhat']['distro'] = 'redhat'
        elif result['redhat']['version'].find("fedora") >= 0:
            result['redhat']['fedora'] = True
            result['redhat']['distro'] = 'fedora'
        else:
            result['distro'] = None

def checkDebian():
    if path.isfile('/etc/debian_version'):
        result['debian'] = {}
        with open('/etc/debian_version', 'r') as content:
            result['debian']['version'] = content.read().lower().strip()
            result['flags']['debian'] = True
        pass

#-------------------------------------------------------------------------------

checkPlatform()
checkUname()
checkLSB()
checkRedHat()
checkDebian()

#-------------------------------------------------------------------------------

dump(result, stdout, indent=2)
exit(0)