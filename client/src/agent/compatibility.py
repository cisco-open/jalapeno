#!/usr/bin/env python
"""Ensures that the Voltron agent is capable of running on the host.
Currently, only Linux is supported.
"""
from os import linesep
from sys import platform
import subprocess
from lib.exceptions import PlatformNotSupported, MPLSNotEnabled, MPLSNotConfigured

def check_compatibility():
    check_platform_compatibility()
    check_mpls_support()
    check_mpls_sysctl_flags()
    print('Tentatively ready to go!')

def check_platform_compatibility():
    validated_platforms = set([
        'linux', 'linux2'
    ])
    if platform not in validated_platforms:
        raise PlatformNotSupported(
            platform,
            'Platform is not supported!'
        )

def check_mpls_support():
    # http://www.samrussell.nz/2015/12/mpls-testbed-on-ubuntu-linux-with.html
    desired_modules = set([
        'mpls_router',
        'mpls_gso',
        'mpls_iptunnel'
    ])
    curr_modules = get_loaded_kernel_modules()
    if not desired_modules.issubset(curr_modules):
        raise MPLSNotEnabled(
            desired_modules,
            'MPLS modules must be loaded.'
        )

def check_mpls_sysctl_flags():
    raw_mpls_flags = _exec_system_command('sysctl -a --pattern mpls')
    if not raw_mpls_flags:
        raise MPLSNotConfigured(
            raw_mpls_flags,
            'Could not find any sysctl flags for MPLS configuration!'
        )
    mpls_flag_lines = raw_mpls_flags.split(sep=linesep)
    mpls_flags = {}
    for mpls_flag in mpls_flag_lines:
        flag, value = mpls_flag.split(sep=' = ')
        mpls_flags[flag] = int(value)
    if mpls_flags['net.mpls.platform_labels'] == 0:
        raise MPLSNotConfigured(
            mpls_flags['net.mpls.platform_labels'],
            'net.mpls.platform_labels must be set greater than 0!'
        )
    mpls_enablement_flags = [value for flag, value in mpls_flags.items() if 'net.mpls.conf' in flag]
    if not any(mpls_enablement_flags):
        raise MPLSNotConfigured(
            mpls_enablement_flags,
            'No interfaces configured for MPLS!'
        )

def get_loaded_kernel_modules():
    raw_kernel_modules = _exec_system_command('lsmod')
    kernel_module_lines = raw_kernel_modules.split(linesep)
    kernel_modules = [module_line.split()[0] for module_line in kernel_module_lines]
    return set(kernel_modules)

def _exec_system_command(command):
    return subprocess.check_output(
        command, shell=True
    ).decode('utf-8').strip()

if __name__ == '__main__':
    check_compatibility()