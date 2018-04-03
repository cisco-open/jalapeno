from subprocess import call
import pexpect, sys

#call(["oc", "login", "https://10.200.99.44:8443"])
oclogin = pexpect.spawn('oc login https://10.200.99.44:8443')
oclogin.delaybeforesend = 1
oclogin.expect('Username:')
oclogin.sendline('admin')
oclogin.expect('Password:')
oclogin.sendline('admin')


